package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// FeedOption configures a feed request.
type FeedOption func(*feedOpts)

type feedOpts struct {
	ordering OrderingMode
	pageSize int
}

// WithOrderingMode sets the feed ordering mode.
func WithOrderingMode(m OrderingMode) FeedOption {
	return func(o *feedOpts) { o.ordering = m }
}

// WithPageSize sets the number of items per page.
func WithPageSize(n int) FeedOption {
	return func(o *feedOpts) { o.pageSize = n }
}

const feedQuery = `query PersonalizedFeed($mainFeedArgs: MainFeedArgs!) {
  me {
    personalizedFeed(mainFeedArgs: $mainFeedArgs) {
      feedItems {
        __typename
        ... on FeedItemPost {
          post {
            id
            subject
            body
            author { displayName url }
            createdAt { epochSeconds }
            mediaAttachments { __typename }
          }
        }
      }
      nextPage
    }
  }
}`

// GetFeed returns the first page of the user's personalized feed.
func (c *Client) GetFeed(ctx context.Context, opts ...FeedOption) (FeedPage, error) {
	fo := feedOpts{
		ordering: OrderRecentPosts,
		pageSize: 10,
	}
	for _, fn := range opts {
		fn(&fo)
	}

	vars := map[string]any{
		"mainFeedArgs": map[string]any{
			"orderingMode": string(fo.ordering),
			"pageSize":     fo.pageSize,
		},
	}

	data, err := c.gql(ctx, "PersonalizedFeed", feedQuery, vars)
	if err != nil {
		return FeedPage{}, fmt.Errorf("GetFeed: %w", err)
	}

	var fr feedResponse
	if err := json.Unmarshal(data, &fr); err != nil {
		return FeedPage{}, fmt.Errorf("GetFeed: %w: %v", ErrRequestFailed, err)
	}

	return parseFeedResponse(fr), nil
}

const feedPageQuery = `query PersonalizedFeed($mainFeedArgs: MainFeedArgs!) {
  me {
    personalizedFeed(mainFeedArgs: $mainFeedArgs) {
      feedItems {
        __typename
        ... on FeedItemPost {
          post {
            id
            subject
            body
            author { displayName url }
            createdAt { epochSeconds }
            mediaAttachments { __typename }
          }
        }
      }
      nextPage
    }
  }
}`

// GetFeedPage fetches the next page of the feed using a cursor.
func (c *Client) GetFeedPage(ctx context.Context, cursor string) (FeedPage, error) {
	vars := map[string]any{
		"mainFeedArgs": map[string]any{
			"cursor":       cursor,
			"orderingMode": string(OrderRecentPosts),
			"pageSize":     10,
		},
	}

	data, err := c.gql(ctx, "PersonalizedFeed", feedPageQuery, vars)
	if err != nil {
		return FeedPage{}, fmt.Errorf("GetFeedPage: %w", err)
	}

	var fr feedResponse
	if err := json.Unmarshal(data, &fr); err != nil {
		return FeedPage{}, fmt.Errorf("GetFeedPage: %w: %v", ErrRequestFailed, err)
	}

	return parseFeedResponse(fr), nil
}

func parseFeedResponse(fr feedResponse) FeedPage {
	pf := fr.Me.PersonalizedFeed
	var posts []Post
	for _, item := range pf.FeedItems {
		if item.Typename != "FeedItemPost" {
			continue
		}
		posts = append(posts, postFromNode(item.Post))
	}
	return FeedPage{
		Posts:      posts,
		NextCursor: pf.NextPage,
		HasNext:    pf.NextPage != "",
	}
}

func postFromNode(n postNode) Post {
	t := time.Unix(int64(n.CreatedAt.EpochSeconds), 0)
	return Post{
		ID:         n.ID,
		Subject:    n.Subject,
		Body:       n.Body,
		AuthorName: n.Author.DisplayName,
		AuthorURL:  n.Author.URL,
		CreatedAt:  t,
	}
}
