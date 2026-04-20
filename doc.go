// Package nextdoor provides a Go client for Nextdoor's internal APIs.
//
// Authentication uses browser session cookies (cookie-based auth).
// All methods accept a context.Context and return typed domain objects.
//
// The client communicates with Nextdoor's GraphQL endpoint and a small
// number of REST endpoints for messaging and ancillary features.
//
// Usage:
//
//	auth := nextdoor.Auth{
//		CSRFToken:   "...",
//		AccessToken: "...",
//	}
//	c, err := nextdoor.New(auth)
//
//	me, err := c.GetMe(ctx)
package nextdoor
