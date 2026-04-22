package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// SearchPostsInput is the typed input for nextdoor_search_posts.
type SearchPostsInput struct {
	Query string `json:"query" jsonschema:"description=keywords to search Nextdoor posts for,required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=max results to return,minimum=1,maximum=50,default=10"`
}

func searchPosts(ctx context.Context, c *nextdoor.Client, in SearchPostsInput) (any, error) {
	res, err := c.SearchPosts(ctx, in.Query)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res, "", limit), nil
}

// SearchNeighborsInput is the typed input for nextdoor_search_neighbors.
type SearchNeighborsInput struct {
	Query string `json:"query" jsonschema:"description=keywords to search Nextdoor neighbors for,required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=max results to return,minimum=1,maximum=50,default=10"`
}

func searchNeighbors(ctx context.Context, c *nextdoor.Client, in SearchNeighborsInput) (any, error) {
	res, err := c.SearchNeighbors(ctx, in.Query)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res, "", limit), nil
}

var searchTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, SearchPostsInput](
		"nextdoor_search_posts",
		"Search Nextdoor posts by query string",
		"SearchPosts",
		searchPosts,
	),
	mcptool.Define[*nextdoor.Client, SearchNeighborsInput](
		"nextdoor_search_neighbors",
		"Search Nextdoor neighbors by query string",
		"SearchNeighbors",
		searchNeighbors,
	),
}
