package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// GetCommentsInput is the typed input for nextdoor_get_comments.
type GetCommentsInput struct {
	PostID string `json:"post_id" jsonschema:"description=Nextdoor post ID,required"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=max comments to return,minimum=1,maximum=50,default=20"`
}

func getComments(ctx context.Context, c *nextdoor.Client, in GetCommentsInput) (any, error) {
	res, err := c.GetComments(ctx, in.PostID)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Comments, res.NextCursor, limit), nil
}

// GetCommentsPageInput is the typed input for nextdoor_get_comments_page.
type GetCommentsPageInput struct {
	PostID string `json:"post_id" jsonschema:"description=Nextdoor post ID,required"`
	Cursor string `json:"cursor" jsonschema:"description=pagination cursor returned by a previous nextdoor_get_comments call,required"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=max comments to return,minimum=1,maximum=50,default=20"`
}

func getCommentsPage(ctx context.Context, c *nextdoor.Client, in GetCommentsPageInput) (any, error) {
	res, err := c.GetCommentsPage(ctx, in.PostID, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Comments, res.NextCursor, limit), nil
}

// CreateCommentInput is the typed input for nextdoor_create_comment.
type CreateCommentInput struct {
	PostID string `json:"post_id" jsonschema:"description=Nextdoor post ID to comment on,required"`
	Body   string `json:"body" jsonschema:"description=plain-text comment body,required"`
}

func createComment(ctx context.Context, c *nextdoor.Client, in CreateCommentInput) (any, error) {
	comment, err := c.CreateComment(ctx, in.PostID, in.Body)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment": comment}, nil
}

// DeleteCommentInput is the typed input for nextdoor_delete_comment.
type DeleteCommentInput struct {
	CommentID string `json:"comment_id" jsonschema:"description=ID of the comment to delete; must be authored by the authenticated user,required"`
}

func deleteComment(ctx context.Context, c *nextdoor.Client, in DeleteCommentInput) (any, error) {
	if err := c.DeleteComment(ctx, in.CommentID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment_id": in.CommentID}, nil
}

var commentTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, GetCommentsInput](
		"nextdoor_get_comments",
		"Fetch the first page of comments on a Nextdoor post",
		"GetComments",
		getComments,
	),
	mcptool.Define[*nextdoor.Client, GetCommentsPageInput](
		"nextdoor_get_comments_page",
		"Fetch a page of comments on a Nextdoor post by cursor (currently unsupported by the upstream API)",
		"GetCommentsPage",
		getCommentsPage,
	),
	mcptool.Define[*nextdoor.Client, CreateCommentInput](
		"nextdoor_create_comment",
		"Post a comment on a Nextdoor post",
		"CreateComment",
		createComment,
	),
	mcptool.Define[*nextdoor.Client, DeleteCommentInput](
		"nextdoor_delete_comment",
		"Delete one of the authenticated user's Nextdoor comments by ID",
		"DeleteComment",
		deleteComment,
	),
}
