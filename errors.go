package nextdoor

import (
	"errors"
	"fmt"
)

// Sentinel errors.
var (
	ErrInvalidAuth   = errors.New("nextdoor: missing or invalid auth credentials")
	ErrUnauthorized  = errors.New("nextdoor: unauthorized (invalid session)")
	ErrForbidden     = errors.New("nextdoor: forbidden")
	ErrNotFound      = errors.New("nextdoor: not found")
	ErrRateLimited   = errors.New("nextdoor: rate limited")
	ErrInvalidParams = errors.New("nextdoor: invalid parameters")
	ErrRequestFailed = errors.New("nextdoor: request failed")
	ErrGraphQL       = errors.New("nextdoor: graphql error")
)

// HTTPError is returned for unexpected non-2xx HTTP responses.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("nextdoor: HTTP %d: %s", e.StatusCode, e.Body)
}

// GraphQLError wraps one or more errors returned in a GraphQL response.
type GraphQLError struct {
	Errors []gqlErrorEntry
}

func (e *GraphQLError) Error() string {
	if len(e.Errors) == 0 {
		return "nextdoor: unknown graphql error"
	}
	return fmt.Sprintf("nextdoor: graphql: %s", e.Errors[0].Message)
}

func (e *GraphQLError) Unwrap() error { return ErrGraphQL }
