package nextdoor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultStreamAPIKey = "gvfqwq34swkh"
	streamAPIURL        = "https://chat.stream-io-api.com"
)

// StreamConfig holds the credentials needed for Stream Chat messaging.
type StreamConfig struct {
	APIKey string `json:"apiKey"`
	Token  string `json:"token"`
	UserID string `json:"userId"`
}

var rtmConfigRe = regexp.MustCompile(`RTM_CONFIG\s*=\s*(\{[^}]+\})`)

const createChannelMutation = `mutation CreateRtmChannel($input: CreateRtmChannelInput!) {
  createRtmChannel(input: $input) {
    channelId
  }
}`

// CreateChannel creates a new messaging channel with the given legacy user
// profile IDs (numeric IDs as strings, e.g. "55720928").
func (c *Client) CreateChannel(ctx context.Context, legacyUserIDs []string) (*Channel, error) {
	if len(legacyUserIDs) == 0 {
		return nil, fmt.Errorf("CreateChannel: %w: legacyUserIDs required", ErrInvalidParams)
	}

	vars := map[string]any{
		"input": map[string]any{
			"legacyUserIds": legacyUserIDs,
			"userIds":       []string{},
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
		Participants: legacyUserIDs,
	}, nil
}

// SendMessage sends a text message to a Stream Chat channel. The channel
// must already exist (see CreateChannel). On first call, the client
// bootstraps RTM credentials from the Nextdoor page.
func (c *Client) SendMessage(ctx context.Context, channelID, body string) (*Message, error) {
	if channelID == "" || body == "" {
		return nil, fmt.Errorf("SendMessage: %w: channelID and body required", ErrInvalidParams)
	}

	cfg, err := c.getStreamConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w", err)
	}

	cid := channelID
	if idx := strings.Index(cid, ":"); idx >= 0 {
		cid = cid[idx+1:]
	}

	url := fmt.Sprintf("https://chat.stream-io-api.com/channels/messaging/%s/message?api_key=%s", cid, cfg.APIKey)

	payload, err := json.Marshal(map[string]any{
		"message": map[string]any{
			"text": body,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cfg.Token)
	req.Header.Set("stream-auth-type", "jwt")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w: reading body: %v", ErrRequestFailed, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("SendMessage: %w: Stream API %d: %s", ErrRequestFailed, resp.StatusCode, truncate(string(raw), 256))
	}

	var streamResp streamSendMessageResponse
	if err := json.Unmarshal(raw, &streamResp); err != nil {
		return nil, fmt.Errorf("SendMessage: %w: %v", ErrRequestFailed, err)
	}

	return &Message{
		ID:        streamResp.Message.ID,
		ChannelID: channelID,
		AuthorID:  streamResp.Message.User.ID,
		Body:      streamResp.Message.Text,
	}, nil
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

// GetChannels lists the user's DM channels via the Stream Chat API.
func (c *Client) GetChannels(ctx context.Context) ([]Channel, error) {
	cfg, err := c.getStreamConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w", err)
	}

	url := fmt.Sprintf("%s/channels?api_key=%s", streamAPIURL, cfg.APIKey)

	payload, err := json.Marshal(map[string]any{
		"filter_conditions": map[string]any{
			"members": map[string]any{
				"$in": []string{cfg.UserID},
			},
		},
		"sort":  []map[string]any{{"field": "last_message_at", "direction": -1}},
		"limit": 20,
	})
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w: %v", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w: %v", ErrRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cfg.Token)
	req.Header.Set("stream-auth-type", "jwt")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetChannels: %w: reading body: %v", ErrRequestFailed, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GetChannels: %w: Stream API %d: %s", ErrRequestFailed, resp.StatusCode, truncate(string(raw), 256))
	}

	var listResp streamChannelListResponse
	if err := json.Unmarshal(raw, &listResp); err != nil {
		return nil, fmt.Errorf("GetChannels: %w: %v", ErrRequestFailed, err)
	}

	var channels []Channel
	for _, ch := range listResp.Channels {
		var participants []string
		for _, m := range ch.Members {
			name := m.User.Name
			if name == "" {
				name = m.User.ID
			}
			participants = append(participants, name)
		}
		channels = append(channels, Channel{
			ID:           ch.Channel.ID,
			Participants: participants,
		})
	}
	return channels, nil
}

// GetMessages returns the message history for a Stream Chat channel.
func (c *Client) GetMessages(ctx context.Context, channelID string) ([]Message, error) {
	if channelID == "" {
		return nil, fmt.Errorf("GetMessages: %w: channelID required", ErrInvalidParams)
	}

	cfg, err := c.getStreamConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMessages: %w", err)
	}

	cid := channelID
	if idx := strings.Index(cid, ":"); idx >= 0 {
		cid = cid[idx+1:]
	}

	url := fmt.Sprintf("%s/channels/messaging/%s/query?api_key=%s", streamAPIURL, cid, cfg.APIKey)

	payload, err := json.Marshal(map[string]any{
		"messages": map[string]any{"limit": 50},
		"state":    true,
	})
	if err != nil {
		return nil, fmt.Errorf("GetMessages: %w: %v", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("GetMessages: %w: %v", ErrRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cfg.Token)
	req.Header.Set("stream-auth-type", "jwt")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetMessages: %w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetMessages: %w: reading body: %v", ErrRequestFailed, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GetMessages: %w: Stream API %d: %s", ErrRequestFailed, resp.StatusCode, truncate(string(raw), 256))
	}

	var queryResp streamChannelQueryResponse
	if err := json.Unmarshal(raw, &queryResp); err != nil {
		return nil, fmt.Errorf("GetMessages: %w: %v", ErrRequestFailed, err)
	}

	var messages []Message
	for _, sm := range queryResp.Messages {
		msg := Message{
			ID:         sm.ID,
			ChannelID:  channelID,
			AuthorID:   sm.User.ID,
			AuthorName: sm.User.Name,
			Body:       sm.Text,
		}
		if t, err := time.Parse(time.RFC3339Nano, sm.CreatedAt); err == nil {
			msg.CreatedAt = t
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// getStreamConfig returns cached Stream Chat credentials, bootstrapping
// them from the Nextdoor page HTML on first call.
func (c *Client) getStreamConfig(ctx context.Context) (*StreamConfig, error) {
	c.streamOnce.Do(func() {
		c.streamConfig, c.streamErr = c.bootstrapRTM(ctx)
	})
	if c.streamErr != nil {
		c.streamOnce = sync.Once{}
		return nil, c.streamErr
	}
	return c.streamConfig, nil
}

// bootstrapRTM fetches the Nextdoor homepage and extracts RTM_CONFIG
// containing the Stream Chat API key and JWT token.
func (c *Client) bootstrapRTM(ctx context.Context) (*StreamConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("bootstrapRTM: %w: %v", ErrRequestFailed, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cookie", c.cookieString())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bootstrapRTM: %w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bootstrapRTM: %w: reading body: %v", ErrRequestFailed, err)
	}

	matches := rtmConfigRe.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("bootstrapRTM: %w: RTM_CONFIG not found in page", ErrRequestFailed)
	}

	var raw struct {
		APIKey string `json:"api_key"`
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(matches[1], &raw); err != nil {
		return nil, fmt.Errorf("bootstrapRTM: %w: parsing RTM_CONFIG: %v", ErrRequestFailed, err)
	}

	cfg := &StreamConfig{
		APIKey: raw.APIKey,
		Token:  raw.Token,
		UserID: raw.UserID,
	}
	if cfg.APIKey == "" {
		cfg.APIKey = defaultStreamAPIKey
	}

	return cfg, nil
}
