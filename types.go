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
	ID        string    `json:"id"`
	ChannelID string    `json:"channelId"`
	AuthorID  string    `json:"authorId"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Channel represents a chat channel.
type Channel struct {
	ID           string   `json:"id"`
	Participants []string `json:"participants"`
}

// Notification represents a notification item.
type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Link      string    `json:"link"`
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
