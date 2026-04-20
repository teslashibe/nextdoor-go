package nextdoor

import (
	"context"
	"fmt"
)

// GetNotifications returns the user's notification feed.
//
// UNVERIFIED: Live testing found no root query field for notifications.
// They may only be accessible via the notification page HTML or a
// different mechanism. This method is a stub and will return an error.
func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	return nil, fmt.Errorf("GetNotifications: %w: notification query not available in GQL schema (unverified)", ErrRequestFailed)
}
