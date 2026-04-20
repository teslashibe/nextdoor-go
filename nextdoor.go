package nextdoor

import (
	"net/http"
	"sync"
	"time"
)

const (
	baseURL          = "https://nextdoor.com"
	gqlPath          = "/api/gql/"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	defaultRetries   = 3
	defaultRetryBase = 500 * time.Millisecond
)

// Client communicates with Nextdoor's internal APIs.
type Client struct {
	auth       Auth
	httpClient *http.Client
	userAgent  string
	maxRetries int
	retryBase  time.Duration

	streamOnce   sync.Once
	streamConfig *StreamConfig
	streamErr    error
}

// Option configures a Client.
type Option func(*Client)

// WithUserAgent overrides the default browser User-Agent string.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithRetry sets the maximum retry count and base backoff duration.
func WithRetry(maxRetries int, base time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryBase = base
	}
}

// WithHTTPClient overrides the default http.Client. Nil is ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// New creates a new Nextdoor client. Returns an error if the required
// auth credentials (CSRFToken, AccessToken) are missing.
func New(auth Auth, opts ...Option) (*Client, error) {
	if auth.CSRFToken == "" || auth.AccessToken == "" {
		return nil, ErrInvalidAuth
	}
	c := &Client{
		auth:       auth,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  defaultUserAgent,
		maxRetries: defaultRetries,
		retryBase:  defaultRetryBase,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}
