package nextdoor

import "time"

// Auth holds the session cookies required for authenticating with Nextdoor.
type Auth struct {
	// CSRFToken is the csrftoken cookie value (also sent as X-CSRFToken header).
	CSRFToken string
	// AccessToken is the ndbr_at cookie value.
	AccessToken string

	// Optional additional cookies.
	DAID      string
	WE        string
	WE3P      string
	SessionID string
}

// ReactionType represents a reaction on a post.
type ReactionType string

const (
	ReactionLike       ReactionType = "like2"
	ReactionInsightful ReactionType = "thank"
	ReactionAgree      ReactionType = "agree"
	ReactionHaha       ReactionType = "funny"
	ReactionWow        ReactionType = "shock"
	ReactionSad        ReactionType = "sad"
)

// OrderingMode controls feed sort order.
type OrderingMode string

const (
	OrderRecentPosts    OrderingMode = "recent_posts"
	OrderTopPosts       OrderingMode = "top_posts"
	OrderRecentActivity OrderingMode = "recent_activity"
)

// Post represents a Nextdoor post.
type Post struct {
	ID         string    `json:"id"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	AuthorName string    `json:"authorName"`
	AuthorURL  string    `json:"authorUrl"`
	MediaURLs  []string  `json:"mediaUrls,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// FeedPage is a paginated slice of posts.
type FeedPage struct {
	Posts      []Post `json:"posts"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasNext    bool   `json:"hasNext"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID         string    `json:"id"`
	AuthorName string    `json:"authorName"`
	AuthorURL  string    `json:"authorUrl"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CommentPage is a paginated slice of comments.
type CommentPage struct {
	Comments   []Comment `json:"comments"`
	TotalCount int       `json:"totalCount,omitempty"`
	NextCursor string    `json:"nextCursor,omitempty"`
	HasNext    bool      `json:"hasNext"`
}

// Profile represents a Nextdoor user profile.
type Profile struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	GivenName        string `json:"givenName"`
	NeighborhoodID   string `json:"neighborhoodId"`
	NeighborhoodName string `json:"neighborhoodName"`
	Pronouns         string `json:"pronouns,omitempty"`
}

// Message represents a chat message.
type Message struct {
	ID         string    `json:"id"`
	ChannelID  string    `json:"channelId"`
	AuthorID   string    `json:"authorId"`
	AuthorName string    `json:"authorName,omitempty"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Channel represents a chat channel.
type Channel struct {
	ID           string   `json:"id"`
	Participants []string `json:"participants"`
}

// Notification represents a notification item.
type Notification struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

// SearchResult represents a single search hit.
type SearchResult struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // "post", "neighbor", "neighborhood", "business", "page"
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// RateLimitState captures rate-limit information from the most recently observed
// response headers. All fields are zero-valued until a response with rate-limit
// headers is received.
type RateLimitState struct {
	Limit      int           `json:"limit"`       // max requests per window (0 = not reported)
	Remaining  int           `json:"remaining"`   // requests left in the current window
	Reset      time.Time     `json:"reset"`       // when the window resets (UTC)
	RetryAfter time.Duration `json:"retry_after"` // set to Retry-After duration after a 429
}

// IsLimited reports whether the current state indicates requests are blocked.
func (r RateLimitState) IsLimited() bool {
	if !r.Reset.IsZero() && r.Remaining == 0 && time.Now().Before(r.Reset) {
		return true
	}
	return r.RetryAfter > 0
}

// ResetIn returns how long until the rate-limit window resets.
// Returns 0 if Reset is in the past or not set.
func (r RateLimitState) ResetIn() time.Duration {
	if r.Reset.IsZero() {
		return 0
	}
	if d := time.Until(r.Reset); d > 0 {
		return d
	}
	return 0
}
