package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// GetPostInput is the typed input for nextdoor_get_post.
type GetPostInput struct {
	PostID string `json:"post_id" jsonschema:"description=Nextdoor post ID,required"`
}

func getPost(ctx context.Context, c *nextdoor.Client, in GetPostInput) (any, error) {
	return c.GetPost(ctx, in.PostID)
}

// CreatePostInput is the typed input for nextdoor_create_post.
type CreatePostInput struct {
	Body           string `json:"body" jsonschema:"description=plain-text post body; Nextdoor auto-extracts the subject from the first sentence,required"`
	NeighborhoodID string `json:"neighborhood_id,omitempty" jsonschema:"description=neighborhood ID to target; defaults to the authenticated user's home neighborhood"`
}

func createPost(ctx context.Context, c *nextdoor.Client, in CreatePostInput) (any, error) {
	var opts []nextdoor.PostOption
	if in.NeighborhoodID != "" {
		opts = append(opts, nextdoor.WithNeighborhoodID(in.NeighborhoodID))
	}
	post, err := c.CreatePost(ctx, in.Body, opts...)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post": post}, nil
}

// DeletePostInput is the typed input for nextdoor_delete_post.
type DeletePostInput struct {
	PostID string `json:"post_id" jsonschema:"description=ID of the post to delete; must be authored by the authenticated user,required"`
}

func deletePost(ctx context.Context, c *nextdoor.Client, in DeletePostInput) (any, error) {
	if err := c.DeletePost(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

// ReactToPostInput is the typed input for nextdoor_react_to_post.
type ReactToPostInput struct {
	PostID   string `json:"post_id" jsonschema:"description=ID of the post to react to,required"`
	Reaction string `json:"reaction" jsonschema:"description=reaction type; allowed: like2,thank,agree,funny,shock,sad,required"`
}

func reactToPost(ctx context.Context, c *nextdoor.Client, in ReactToPostInput) (any, error) {
	reactionID, err := c.ReactToPost(ctx, in.PostID, nextdoor.ReactionType(in.Reaction))
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "reaction_id": reactionID}, nil
}

// RemoveReactionInput is the typed input for nextdoor_remove_reaction.
type RemoveReactionInput struct {
	ReactionID string `json:"reaction_id" jsonschema:"description=reaction ID returned by nextdoor_react_to_post,required"`
}

func removeReaction(ctx context.Context, c *nextdoor.Client, in RemoveReactionInput) (any, error) {
	if err := c.RemoveReaction(ctx, in.ReactionID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "reaction_id": in.ReactionID}, nil
}

var postTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, GetPostInput](
		"nextdoor_get_post",
		"Fetch a single Nextdoor post by ID",
		"GetPost",
		getPost,
	),
	mcptool.Define[*nextdoor.Client, CreatePostInput](
		"nextdoor_create_post",
		"Create a Nextdoor post in the user's neighborhood (or a specified neighborhood)",
		"CreatePost",
		createPost,
	),
	mcptool.Define[*nextdoor.Client, DeletePostInput](
		"nextdoor_delete_post",
		"Delete one of the authenticated user's Nextdoor posts by ID",
		"DeletePost",
		deletePost,
	),
	mcptool.Define[*nextdoor.Client, ReactToPostInput](
		"nextdoor_react_to_post",
		"Add a reaction (like2, thank, agree, funny, shock, sad) to a Nextdoor post",
		"ReactToPost",
		reactToPost,
	),
	mcptool.Define[*nextdoor.Client, RemoveReactionInput](
		"nextdoor_remove_reaction",
		"Remove a previously-added reaction from a Nextdoor post by reaction ID",
		"RemoveReaction",
		removeReaction,
	),
}
