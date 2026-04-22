package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// GetMeInput is the typed input for nextdoor_get_me.
type GetMeInput struct{}

func getMe(ctx context.Context, c *nextdoor.Client, _ GetMeInput) (any, error) {
	return c.GetMe(ctx)
}

// GetProfileInput is the typed input for nextdoor_get_profile.
type GetProfileInput struct {
	UserID string `json:"user_id" jsonschema:"description=Nextdoor user ID,required"`
}

func getProfile(ctx context.Context, c *nextdoor.Client, in GetProfileInput) (any, error) {
	return c.GetProfile(ctx, in.UserID)
}

var profileTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, GetMeInput](
		"nextdoor_get_me",
		"Fetch the authenticated Nextdoor user's profile (display name, neighborhood, pronouns)",
		"GetMe",
		getMe,
	),
	mcptool.Define[*nextdoor.Client, GetProfileInput](
		"nextdoor_get_profile",
		"Fetch a Nextdoor user profile by user ID",
		"GetProfile",
		getProfile,
	),
}
