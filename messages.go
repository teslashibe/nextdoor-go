package nextdoor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const createChannelMutation = `mutation CreateRtmChannel($input: CreateRtmChannelInput!) {
  createRtmChannel(input: $input) {
    channelId
  }
}`

// CreateChannel creates a new messaging channel with the given participants.
func (c *Client) CreateChannel(ctx context.Context, participantIDs []string) (*Channel, error) {
	if len(participantIDs) == 0 {
		return nil, fmt.Errorf("CreateChannel: %w: participantIDs required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"participantIds": participantIDs,
		},
	}

	data, err := c.gql(ctx, "CreateRtmChannel", createChannelMutation, vars)
	if err != nil {
		return nil, fmt.Errorf("CreateChannel: %w", err)
	}

	var resp createChannelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateChannel: %w: %v", ErrRequestFailed, err)
	}

	return &Channel{
		ID:           resp.CreateRtmChannel.ChannelID,
		Participants: participantIDs,
	}, nil
}

// SendMessage sends a text message to a channel via the REST chat API.
func (c *Client) SendMessage(ctx context.Context, channelID, body string) (*Message, error) {
	if channelID == "" || body == "" {
		return nil, fmt.Errorf("SendMessage: %w: channelID and body required", ErrInvalidParams)
	}

	payload, err := json.Marshal(map[string]any{
		"channel_id": channelID,
		"body":       body,
	})
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}

	raw, err := c.makeRequest(ctx, http.MethodPost, baseURL+"/api/chat/chats", payload)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w", err)
	}

	var resp struct {
		ID        string `json:"id"`
		ChannelID string `json:"channel_id"`
		AuthorID  string `json:"author_id"`
		Body      string `json:"body"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}

	return &Message{
		ID:        resp.ID,
		ChannelID: resp.ChannelID,
		AuthorID:  resp.AuthorID,
		Body:      resp.Body,
	}, nil
}

// GetChannels lists the user's messaging channels via the REST chat API.
func (c *Client) GetChannels(ctx context.Context) ([]Channel, error) {
	raw, err := c.makeRequest(ctx, http.MethodGet, baseURL+"/api/chat/chats", nil)
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w", err)
	}

	var resp chatListResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("GetChannels: %w: %v", ErrRequestFailed, err)
	}

	channels := make([]Channel, 0, len(resp.Chats))
	for _, ch := range resp.Chats {
		channels = append(channels, Channel{
			ID:           ch.ID,
			Participants: ch.Participants,
		})
	}
	return channels, nil
}

const deleteMessageMutation = `mutation DeleteRtmMessageV2($input: DeleteRtmMessageInput!) {
  deleteRtmMessageV2(input: $input) {
    success
  }
}`

// DeleteMessage deletes a message by ID.
func (c *Client) DeleteMessage(ctx context.Context, messageID string) error {
	if messageID == "" {
		return fmt.Errorf("DeleteMessage: %w: messageID required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"messageId": messageID,
		},
	}

	data, err := c.gql(ctx, "DeleteRtmMessageV2", deleteMessageMutation, vars)
	if err != nil {
		return fmt.Errorf("DeleteMessage: %w", err)
	}

	var resp deleteMessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("DeleteMessage: %w: %v", ErrRequestFailed, err)
	}
	if !resp.DeleteRtmMessageV2.Success {
		return fmt.Errorf("DeleteMessage: %w: server returned success=false", ErrRequestFailed)
	}
	return nil
}
