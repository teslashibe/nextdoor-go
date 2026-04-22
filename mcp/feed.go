package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// GetFeedInput is the typed input for nextdoor_get_feed.
type GetFeedInput struct {
	Ordering string `json:"ordering,omitempty" jsonschema:"description=feed ordering mode; allowed: recent_posts,top_posts,recent_activity,default=recent_posts"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"description=number of posts per page,minimum=1,maximum=50,default=10"`
}

func getFeed(ctx context.Context, c *nextdoor.Client, in GetFeedInput) (any, error) {
	var opts []nextdoor.FeedOption
	if in.Ordering != "" {
		opts = append(opts, nextdoor.WithOrderingMode(nextdoor.OrderingMode(in.Ordering)))
	}
	if in.PageSize > 0 {
		opts = append(opts, nextdoor.WithPageSize(in.PageSize))
	}
	res, err := c.GetFeed(ctx, opts...)
	if err != nil {
		return nil, err
	}
	limit := in.PageSize
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res.Posts, res.NextCursor, limit), nil
}

// GetFeedPageInput is the typed input for nextdoor_get_feed_page.
type GetFeedPageInput struct {
	Cursor string `json:"cursor" jsonschema:"description=pagination cursor returned by a previous nextdoor_get_feed call,required"`
}

func getFeedPage(ctx context.Context, c *nextdoor.Client, in GetFeedPageInput) (any, error) {
	res, err := c.GetFeedPage(ctx, in.Cursor)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(res.Posts, res.NextCursor, 10), nil
}

var feedTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, GetFeedInput](
		"nextdoor_get_feed",
		"Fetch the authenticated user's personalized Nextdoor feed (recent posts, top posts, or recent activity)",
		"GetFeed",
		getFeed,
	),
	mcptool.Define[*nextdoor.Client, GetFeedPageInput](
		"nextdoor_get_feed_page",
		"Fetch the next page of the personalized Nextdoor feed using a cursor",
		"GetFeedPage",
		getFeedPage,
	),
}
