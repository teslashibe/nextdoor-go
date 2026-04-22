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
	"strconv"
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
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

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

	c.updateRateLimit(resp.Header)

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
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		c.rlMu.Lock()
		c.rlState.Remaining = 0
		c.rlState.RetryAfter = wait
		if c.rlState.Reset.IsZero() || time.Until(c.rlState.Reset) < wait {
			c.rlState.Reset = time.Now().Add(wait)
		}
		c.rlMu.Unlock()
		c.gapMu.Lock()
		if earliest := time.Now().Add(wait); c.lastReqAt.Before(earliest) {
			c.lastReqAt = earliest
		}
		c.gapMu.Unlock()
		return nil, fmt.Errorf("%w: retry after %s", ErrRateLimited, wait)
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

func (c *Client) updateRateLimit(h http.Header) {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	if v := rlHeader(h, "Limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rlState.Limit = n
		}
	}
	if v := rlHeader(h, "Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rlState.Remaining = n
		}
	}
	if v := rlHeader(h, "Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			if ts > 1_000_000_000 {
				c.rlState.Reset = time.Unix(ts, 0)
			} else {
				c.rlState.Reset = time.Now().Add(time.Duration(ts) * time.Second)
			}
		}
	}
}

func rlHeader(h http.Header, suffix string) string {
	for _, p := range []string{"X-RateLimit-", "X-Rate-Limit-", "X-Ratelimit-", "RateLimit-"} {
		if v := strings.TrimSpace(h.Get(p + suffix)); v != "" {
			return v
		}
	}
	return ""
}

func (c *Client) adaptiveGap() time.Duration {
	c.rlMu.Lock()
	rs := c.rlState
	c.rlMu.Unlock()

	if rs.Remaining == 0 && !rs.Reset.IsZero() {
		if d := time.Until(rs.Reset); d > 0 {
			return d + 50*time.Millisecond
		}
	}
	if rs.Remaining > 0 && !rs.Reset.IsZero() {
		if d := time.Until(rs.Reset); d > 0 {
			spread := d / time.Duration(float64(rs.Remaining)*0.9)
			if spread > c.minGap {
				return spread
			}
		}
	}
	return c.minGap
}

func (c *Client) waitForGap(ctx context.Context) {
	gap := c.adaptiveGap()
	c.gapMu.Lock()
	now := time.Now()
	next := c.lastReqAt.Add(gap)
	if now.After(next) {
		next = now
	}
	c.lastReqAt = next
	c.gapMu.Unlock()

	if wait := time.Until(next); wait > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
	}
	c.rlMu.Lock()
	c.rlState.RetryAfter = 0
	c.rlMu.Unlock()
}

func parseRetryAfter(val string, fallback time.Duration) time.Duration {
	if val == "" {
		return fallback
	}
	trimmed := strings.TrimSpace(val)
	if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		if n > 1_000_000_000 {
			if d := time.Until(time.Unix(n, 0)); d > 0 {
				return d
			}
			return fallback
		}
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(trimmed); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return fallback
}
