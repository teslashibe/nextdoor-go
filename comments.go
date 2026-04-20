package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

const getCommentsQuery = `query PagedComments($postId: ID!, $cursor: String, $pageSize: Int) {
  pagedComments(postId: $postId, cursor: $cursor, pageSize: $pageSize) {
    comments {
      id
      author { displayName url }
      body
      createdAt { epochSeconds }
    }
    nextPage
  }
}`

// GetComments returns the first page of comments for a post.
func (c *Client) GetComments(ctx context.Context, postID string) (CommentPage, error) {
	return c.getComments(ctx, postID, "")
}

// GetCommentsPage returns a page of comments starting from a cursor.
func (c *Client) GetCommentsPage(ctx context.Context, postID, cursor string) (CommentPage, error) {
	return c.getComments(ctx, postID, cursor)
}

func (c *Client) getComments(ctx context.Context, postID, cursor string) (CommentPage, error) {
	if postID == "" {
		return CommentPage{}, fmt.Errorf("GetComments: %w: postID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"postId":   postID,
		"pageSize": 20,
	}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	data, err := c.gql(ctx, "PagedComments", getCommentsQuery, vars)
	if err != nil {
		return CommentPage{}, fmt.Errorf("GetComments: %w", err)
	}

	var resp pagedCommentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return CommentPage{}, fmt.Errorf("GetComments: %w: %v", ErrRequestFailed, err)
	}

	pc := resp.PagedComments
	var comments []Comment
	for _, n := range pc.Comments {
		comments = append(comments, commentFromNode(n))
	}

	return CommentPage{
		Comments:   comments,
		NextCursor: pc.NextPage,
		HasNext:    pc.NextPage != "",
	}, nil
}

const createCommentMutation = `mutation CreateCommentV3($input: CreateCommentInput!) {
  create_comment_v3(input: $input) {
    comment {
      id
      author { displayName url }
      body
      createdAt { epochSeconds }
    }
  }
}`

// CreateComment adds a comment to a post.
func (c *Client) CreateComment(ctx context.Context, postID, body string) (*Comment, error) {
	if postID == "" || body == "" {
		return nil, fmt.Errorf("CreateComment: %w: postID and body required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"postId": postID,
			"body":   body,
		},
	}

	data, err := c.gql(ctx, "CreateCommentV3", createCommentMutation, vars)
	if err != nil {
		return nil, fmt.Errorf("CreateComment: %w", err)
	}

	var resp createCommentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateComment: %w: %v", ErrRequestFailed, err)
	}

	cm := commentFromNode(resp.CreateCommentV3.Comment)
	return &cm, nil
}

const deleteCommentMutation = `mutation DeleteComment($commentId: ID!) {
  delete_comment(commentId: $commentId) {
    success
  }
}`

// DeleteComment removes a comment by ID.
func (c *Client) DeleteComment(ctx context.Context, commentID string) error {
	if commentID == "" {
		return fmt.Errorf("DeleteComment: %w: commentID required", ErrInvalidParams)
	}

	vars := map[string]any{"commentId": commentID}
	data, err := c.gql(ctx, "DeleteComment", deleteCommentMutation, vars)
	if err != nil {
		return fmt.Errorf("DeleteComment: %w", err)
	}

	var resp deleteCommentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("DeleteComment: %w: %v", ErrRequestFailed, err)
	}
	if !resp.DeleteComment.Success {
		return fmt.Errorf("DeleteComment: %w: server returned success=false", ErrRequestFailed)
	}
	return nil
}

func commentFromNode(n commentNode) Comment {
	return Comment{
		ID:         n.ID,
		AuthorName: n.Author.DisplayName,
		AuthorURL:  n.Author.URL,
		Body:       n.Body,
		CreatedAt:  epochToTime(n.CreatedAt.EpochSeconds),
	}
}
