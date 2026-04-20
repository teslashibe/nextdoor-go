package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

// PostOption configures a CreatePost call.
type PostOption func(*postOpts)

type postOpts struct {
	subject string
}

// WithSubject sets the post subject line.
func WithSubject(s string) PostOption {
	return func(o *postOpts) { o.subject = s }
}

const getPostQuery = `query GetPost($postId: ID!) {
  post(postId: $postId) {
    id
    subject
    body
    author { displayName url }
    createdAt { epochSeconds }
    mediaAttachments { __typename }
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
    post {
      id
      subject
      body
      author { displayName url }
      createdAt { epochSeconds }
      mediaAttachments { __typename }
    }
  }
}`

// CreatePost creates a new post in the user's neighborhood.
func (c *Client) CreatePost(ctx context.Context, body string, opts ...PostOption) (*Post, error) {
	if body == "" {
		return nil, fmt.Errorf("CreatePost: %w: body required", ErrInvalidParams)
	}

	po := postOpts{}
	for _, fn := range opts {
		fn(&po)
	}

	input := map[string]any{"body": body}
	if po.subject != "" {
		input["subject"] = po.subject
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

	p := postFromNode(resp.CreatePostV3.Post)
	return &p, nil
}

const deletePostMutation = `mutation DeletePost($input: DeletePostInput!) {
  deletePost(input: $input) {
    success
  }
}`

// DeletePost deletes a post by ID.
func (c *Client) DeletePost(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("DeletePost: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{"input": map[string]any{"postId": postID}}
	data, err := c.gql(ctx, "DeletePost", deletePostMutation, vars)
	if err != nil {
		return fmt.Errorf("DeletePost: %w", err)
	}

	var resp deletePostResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("DeletePost: %w: %v", ErrRequestFailed, err)
	}
	if !resp.DeletePost.Success {
		return fmt.Errorf("DeletePost: %w: server returned success=false", ErrRequestFailed)
	}
	return nil
}

const addReactionMutation = `mutation AddReactionToPost($input: AddReactionToPostInput!) {
  addReactionToPost(input: $input) {
    success
  }
}`

// ReactToPost adds a reaction to a post.
func (c *Client) ReactToPost(ctx context.Context, postID string, reaction ReactionType) error {
	if postID == "" {
		return fmt.Errorf("ReactToPost: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"postId":       postID,
			"reactionType": string(reaction),
		},
	}
	data, err := c.gql(ctx, "AddReactionToPost", addReactionMutation, vars)
	if err != nil {
		return fmt.Errorf("ReactToPost: %w", err)
	}

	var resp addReactionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("ReactToPost: %w: %v", ErrRequestFailed, err)
	}
	return nil
}

const removeReactionMutation = `mutation RemoveReactionFromPost($input: RemoveReactionInput!) {
  removeReactionFromPost(input: $input) {
    success
  }
}`

// RemoveReaction removes the current user's reaction from a post.
func (c *Client) RemoveReaction(ctx context.Context, postID string) error {
	if postID == "" {
		return fmt.Errorf("RemoveReaction: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{"postId": postID},
	}
	data, err := c.gql(ctx, "RemoveReactionFromPost", removeReactionMutation, vars)
	if err != nil {
		return fmt.Errorf("RemoveReaction: %w", err)
	}

	var resp removeReactionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("RemoveReaction: %w: %v", ErrRequestFailed, err)
	}
	return nil
}
