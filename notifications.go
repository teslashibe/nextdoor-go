package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const notificationsQuery = `query NotificationFeed($args: NotificationFeedArgs) {
  notificationFeed(args: $args) {
    notifications {
      id
      title
      body
      link
      read
      createdAt { epochSeconds }
    }
  }
}`

// GetNotifications returns the user's notification feed.
func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	data, err := c.gql(ctx, "NotificationFeed", notificationsQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("GetNotifications: %w", err)
	}

	var resp notificationFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetNotifications: %w: %v", ErrRequestFailed, err)
	}

	notifs := make([]Notification, 0, len(resp.NotificationFeed.Notifications))
	for _, n := range resp.NotificationFeed.Notifications {
		t := time.Unix(int64(n.CreatedAt.EpochSeconds), 0)
		notifs = append(notifs, Notification{
			ID:        n.ID,
			Title:     n.Title,
			Body:      n.Body,
			Link:      n.Link,
			Read:      n.Read,
			CreatedAt: t,
		})
	}
	return notifs, nil
}
