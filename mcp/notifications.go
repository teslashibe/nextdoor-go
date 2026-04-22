package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// GetNotificationsInput is the typed input for nextdoor_get_notifications.
type GetNotificationsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"description=max notifications to return,minimum=1,maximum=100,default=20"`
}

func getNotifications(ctx context.Context, c *nextdoor.Client, in GetNotificationsInput) (any, error) {
	res, err := c.GetNotifications(ctx)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res, "", limit), nil
}

var notificationTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, GetNotificationsInput](
		"nextdoor_get_notifications",
		"List the authenticated user's Nextdoor notification feed",
		"GetNotifications",
		getNotifications,
	),
}
