package nextdoor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// gql executes a GraphQL operation against /api/gql/ and returns the raw
// data payload. It returns a *GraphQLError when the response contains
// errors.
func (c *Client) gql(ctx context.Context, operationName, query string, variables any) (json.RawMessage, error) {
	var varsRaw json.RawMessage
	if variables != nil {
		b, err := json.Marshal(variables)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
		}
		varsRaw = b
	}

	body, err := json.Marshal(gqlRequest{
		OperationName: operationName,
		Query:         query,
		Variables:     varsRaw,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	raw, err := c.makeRequest(ctx, http.MethodPost, baseURL+gqlPath, body)
	if err != nil {
		return nil, err
	}

	var resp gqlResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed response: %v", ErrRequestFailed, err)
	}
	if len(resp.Errors) > 0 {
		return nil, &GraphQLError{Errors: resp.Errors}
	}
	return resp.Data, nil
}

// makeRequest performs an HTTP request with automatic retry on 429 and 5xx.
func (c *Client) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			wait := c.backoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		raw, err := c.doRequest(ctx, method, url, body)
		if err == nil {
			return raw, nil
		}

		lastErr = err

		if errors.Is(err, ErrRateLimited) {
			continue
		}
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode >= 500 {
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

// doRequest performs a single HTTP round-trip.
func (c *Client) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("X-CSRFToken", c.auth.CSRFToken)
	req.Header.Set("Cookie", c.cookieString())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return raw, nil
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("%w: %s", ErrRateLimited, truncate(string(raw), 256))
	default:
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: truncate(string(raw), 256)}
	}
}

// cookieString builds the Cookie header value from Auth fields.
func (c *Client) cookieString() string {
	var parts []string
	if c.auth.CSRFToken != "" {
		parts = append(parts, "csrftoken="+c.auth.CSRFToken)
	}
	if c.auth.AccessToken != "" {
		parts = append(parts, "ndbr_at="+c.auth.AccessToken)
	}
	if c.auth.DAID != "" {
		parts = append(parts, "DAID="+c.auth.DAID)
	}
	if c.auth.WE != "" {
		parts = append(parts, "WE="+c.auth.WE)
	}
	if c.auth.WE3P != "" {
		parts = append(parts, "WE3P="+c.auth.WE3P)
	}
	if c.auth.SessionID != "" {
		parts = append(parts, "ndp_session_id="+c.auth.SessionID)
	}
	return strings.Join(parts, "; ")
}

// backoff returns the exponential-backoff delay for the given attempt (1-indexed).
func (c *Client) backoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt-1))) * c.retryBase
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
