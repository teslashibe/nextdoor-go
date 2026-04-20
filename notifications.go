package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
)

const notificationFeedQuery = `query NotificationFeed {
  me {
    notificationFeed {
      badgeCount
      feedItems {
        __typename
        ... on FeedItemNotification {
          notification {
            id
            body { text }
            isRead
            createdAt { epochSeconds }
          }
        }
      }
    }
  }
}`

// GetNotifications returns the user's notification feed.
func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	data, err := c.gql(ctx, "NotificationFeed", notificationFeedQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("GetNotifications: %w", err)
	}

	var resp notificationFeedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetNotifications: %w: %v", ErrRequestFailed, err)
	}

	var notifications []Notification
	for _, item := range resp.Me.NotificationFeed.FeedItems {
		if item.Typename != "FeedItemNotification" {
			continue
		}
		n := item.Notification
		notifications = append(notifications, Notification{
			ID:        n.ID,
			Body:      n.Body.Text,
			Read:      n.IsRead,
			CreatedAt: epochToTime(n.CreatedAt.EpochSeconds),
		})
	}
	return notifications, nil
}
