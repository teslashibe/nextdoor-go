package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

// PostOption configures a CreatePost call.
type PostOption func(*postOpts)

type postOpts struct {
	neighborhoodID string
}

// WithNeighborhoodID targets the post to a specific neighborhood.
// If not set, the user's home neighborhood is used via GetMe.
func WithNeighborhoodID(id string) PostOption {
	return func(o *postOpts) { o.neighborhoodID = id }
}

const getPostQuery = `query GetPost($postId: ID!) {
  post(postId: $postId) {
    id
    subject
    body
    author { displayName url }
    createdAt { epochSeconds }
    mediaAttachments { __typename url }
  }
}`

// GetPost retrieves a single post by ID.
func (c *Client) GetPost(ctx context.Context, postID string) (*Post, error) {
	if postID == "" {
		return nil, fmt.Errorf("GetPost: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{"postId": postID}
	data, err := c.gql(ctx, "GetPost", getPostQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("GetPost: %w", err)
	}

	var resp struct {
		Post postNode `json:"post"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetPost: %w: %v", ErrRequestFailed, err)
	}

	p := postFromNode(resp.Post)
	return &p, nil
}

const createPostMutation = `mutation CreatePostV3($input: CreatePostV2Input!) {
  createPostV3(input: $input) {
    ... on CreatePostPayloadV2 {
      feedPostItem {
        ... on FeedItemPost {
          post {
            id
            subject
            body
            author { displayName url }
            createdAt { epochSeconds }
            mediaAttachments { __typename url }
          }
        }
      }
    }
  }
}`

// CreatePost creates a new post in the user's neighborhood. The body text
// is required. Nextdoor auto-extracts the subject from the first sentence.
//
// By default the post targets the authenticated user's home neighborhood.
// Use WithNeighborhoodID to override.
func (c *Client) CreatePost(ctx context.Context, body string, opts ...PostOption) (*Post, error) {
	if body == "" {
		return nil, fmt.Errorf("CreatePost: %w: body required", ErrInvalidParams)
	}

	po := postOpts{}
	for _, fn := range opts {
		fn(&po)
	}

	hoodID := po.neighborhoodID
	if hoodID == "" {
		me, err := c.GetMe(ctx)
		if err != nil {
			return nil, fmt.Errorf("CreatePost: resolving neighborhood: %w", err)
		}
		hoodID = stripPrefix(me.NeighborhoodID, "neighborhood_")
	}

	input := map[string]any{
		"body": body,
		"postAudienceAndDistribution": map[string]any{
			"neighborhoodId": hoodID,
		},
	}

	vars := map[string]any{"input": input}
	data, err := c.gql(ctx, "CreatePostV3", createPostMutation, vars)
	if err != nil {
		return nil, fmt.Errorf("CreatePost: %w", err)
	}

	var resp createPostResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreatePost: %w: %v", ErrRequestFailed, err)
	}

	p := postFromNode(resp.CreatePostV3.FeedPostItem.Post)
	return &p, nil
}

const deletePostMutation = `mutation deletePost($input: DeletePostInput!) {
  deletePost(input: $input) {
    __typename
  }
}`

// DeletePost deletes a post by ID.
func (c *Client) DeletePost(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("DeletePost: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{"input": map[string]any{"postId": postID}}
	_, err := c.gql(ctx, "deletePost", deletePostMutation, vars)
	if err != nil {
		return fmt.Errorf("DeletePost: %w", err)
	}
	return nil
}

const addReactionMutation = `mutation AddReactionToPost($input: AddReactionToPostInput!) {
  addReactionToPost(input: $input) {
    post {
      reactionSummaries {
        summaries {
          userReactionId
        }
      }
    }
  }
}`

// ReactToPost adds a reaction to a post and returns the reaction ID needed
// for later removal via RemoveReaction.
func (c *Client) ReactToPost(ctx context.Context, postID string, reaction ReactionType) (string, error) {
	if postID == "" {
		return "", fmt.Errorf("ReactToPost: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"postId":       postID,
			"reactionName": string(reaction),
		},
	}
	data, err := c.gql(ctx, "AddReactionToPost", addReactionMutation, vars)
	if err != nil {
		return "", fmt.Errorf("ReactToPost: %w", err)
	}

	var resp addReactionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("ReactToPost: %w: %v", ErrRequestFailed, err)
	}

	summaries := resp.AddReactionToPost.Post.ReactionSummaries.Summaries
	if len(summaries) > 0 {
		return summaries[0].UserReactionID, nil
	}
	return "", nil
}

const removeReactionMutation = `mutation RemoveReactionFromPost($input: RemoveReactionInput!) {
  removeReactionFromPost(input: $input) {
    __typename
  }
}`

// RemoveReaction removes a reaction by its reaction ID (obtained from
// ReactToPost or from post.reactionSummaries).
func (c *Client) RemoveReaction(ctx context.Context, reactionID string) error {
	if reactionID == "" {
		return fmt.Errorf("RemoveReaction: %w: reactionID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{"reactionId": reactionID},
	}
	_, err := c.gql(ctx, "RemoveReactionFromPost", removeReactionMutation, vars)
	if err != nil {
		return fmt.Errorf("RemoveReaction: %w", err)
	}
	return nil
}

func stripPrefix(s, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
