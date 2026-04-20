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
	ID         string
	Subject    string
	Body       string
	AuthorName string
	AuthorURL  string
	CreatedAt  time.Time
}

// FeedPage is a paginated slice of posts.
type FeedPage struct {
	Posts      []Post
	NextCursor string
	HasNext    bool
}

// Comment represents a comment on a post.
type Comment struct {
	ID         string
	AuthorName string
	AuthorURL  string
	Body       string
	CreatedAt  time.Time
}

// CommentPage is a paginated slice of comments.
type CommentPage struct {
	Comments   []Comment
	NextCursor string
	HasNext    bool
}

// Profile represents a Nextdoor user profile.
type Profile struct {
	ID               string
	DisplayName      string
	GivenName        string
	NeighborhoodID   string
	NeighborhoodName string
	Pronouns         string
}

// Message represents a chat message.
type Message struct {
	ID        string
	ChannelID string
	AuthorID  string
	Body      string
	CreatedAt time.Time
}

// Channel represents a chat channel.
type Channel struct {
	ID           string
	Participants []string
}

// Notification represents a notification item.
type Notification struct {
	ID        string
	Title     string
	Body      string
	Link      string
	Read      bool
	CreatedAt time.Time
}

// SearchResult represents a single search hit.
type SearchResult struct {
	ID    string
	Type  string // "post", "neighbor", "neighborhood", "business", "page"
	Title string
	Body  string
	URL   string
}
