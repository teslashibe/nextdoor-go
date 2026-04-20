package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

const searchPostFeedQuery = `query SearchPostFeed($query: String!) {
  searchPostFeed(query: $query) {
    results {
      id
      title
      body
      url
      type
    }
  }
}`

// SearchPosts searches for posts matching the query string.
func (c *Client) SearchPosts(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("SearchPosts: %w: query required", ErrInvalidParams)
	}

	vars := map[string]any{"query": query}
	data, err := c.gql(ctx, "SearchPostFeed", searchPostFeedQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("SearchPosts: %w", err)
	}

	var resp searchPostFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("SearchPosts: %w: %v", ErrRequestFailed, err)
	}

	return searchNodesTo(resp.SearchPostFeed.Results, "post"), nil
}

const searchNeighborsQuery = `query SearchNeighbor($query: String!) {
  searchNeighbor(query: $query) {
    results {
      id
      title
      body
      url
      type
    }
  }
}`

// SearchNeighbors searches for neighbors matching the query string.
func (c *Client) SearchNeighbors(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("SearchNeighbors: %w: query required", ErrInvalidParams)
	}

	vars := map[string]any{"query": query}
	data, err := c.gql(ctx, "SearchNeighbor", searchNeighborsQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("SearchNeighbors: %w", err)
	}

	var resp searchNeighborsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("SearchNeighbors: %w: %v", ErrRequestFailed, err)
	}

	return searchNodesTo(resp.SearchNeighbor.Results, "neighbor"), nil
}

func searchNodesTo(nodes []searchNode, defaultType string) []SearchResult {
	results := make([]SearchResult, 0, len(nodes))
	for _, n := range nodes {
		t := n.Type
		if t == "" {
			t = defaultType
		}
		results = append(results, SearchResult{
			ID:    n.ID,
			Type:  t,
			Title: n.Title,
			Body:  n.Body,
			URL:   n.URL,
		})
	}
	return results
}
