package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

var requestCounter uint64

func nextRequestID() string {
	n := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), strconv.FormatUint(n, 10))
}

const searchPostFeedQuery = `query SearchPostFeed($postSearchArgs: PostSearchArgs!) {
  searchPostFeed(postSearchArgs: $postSearchArgs) {
    searchResultView {
      __typename
      ... on SearchResultSection {
        searchResultItems {
          edges {
            node {
              title { text }
              body { text }
              url
              contentId
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  }
}`

// SearchPosts searches for posts matching the query string.
func (c *Client) SearchPosts(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("SearchPosts: %w: query required", ErrInvalidParams)
	}

	vars := map[string]any{
		"postSearchArgs": map[string]any{
			"query":     query,
			"requestId": nextRequestID(),
		},
	}
	data, err := c.gql(ctx, "SearchPostFeed", searchPostFeedQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("SearchPosts: %w", err)
	}

	var resp searchPostFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("SearchPosts: %w: %v", ErrRequestFailed, err)
	}

	return extractSearchResults(resp.SearchPostFeed.SearchResultView, "post")
}

const searchNeighborFeedQuery = `query SearchNeighborFeed($neighborSearchArgs: NeighborSearchArgs!) {
  searchNeighborFeed(neighborSearchArgs: $neighborSearchArgs) {
    searchResultView {
      __typename
      ... on SearchResultSection {
        searchResultItems {
          edges {
            node {
              title { text }
              body { text }
              url
              contentId
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  }
}`

// SearchNeighbors searches for neighbors matching the query string.
func (c *Client) SearchNeighbors(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("SearchNeighbors: %w: query required", ErrInvalidParams)
	}

	vars := map[string]any{
		"neighborSearchArgs": map[string]any{
			"query":     query,
			"requestId": nextRequestID(),
		},
	}
	data, err := c.gql(ctx, "SearchNeighborFeed", searchNeighborFeedQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("SearchNeighbors: %w", err)
	}

	var resp searchNeighborFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("SearchNeighbors: %w: %v", ErrRequestFailed, err)
	}

	return extractSearchResults(resp.SearchNeighborFeed.SearchResultView, "neighbor")
}

func extractSearchResults(views []json.RawMessage, defaultType string) ([]SearchResult, error) {
	var results []SearchResult
	for _, raw := range views {
		var section searchResultSection
		if err := json.Unmarshal(raw, &section); err != nil {
			continue
		}
		if section.Typename != "SearchResultSection" {
			continue
		}
		for _, e := range section.SearchResultItems.Edges {
			n := e.Node
			results = append(results, SearchResult{
				ID:    n.ContentID,
				Type:  defaultType,
				Title: n.Title.Text,
				Body:  n.Body.Text,
				URL:   n.URL,
			})
		}
	}
	return results, nil
}
