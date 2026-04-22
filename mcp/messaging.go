package mcp

import (
	"context"

	nextdoor "github.com/teslashibe/nextdoor-go"
	"github.com/teslashibe/mcptool"
)

// CreateChannelInput is the typed input for nextdoor_create_channel.
type CreateChannelInput struct {
	LegacyUserIDs []string `json:"legacy_user_ids" jsonschema:"description=numeric Nextdoor user IDs (as strings) to include in the new DM channel,required"`
}

func createChannel(ctx context.Context, c *nextdoor.Client, in CreateChannelInput) (any, error) {
	channel, err := c.CreateChannel(ctx, in.LegacyUserIDs)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "channel": channel}, nil
}

// SendMessageInput is the typed input for nextdoor_send_message.
type SendMessageInput struct {
	ChannelID string `json:"channel_id" jsonschema:"description=Nextdoor DM channel ID (from nextdoor_create_channel or nextdoor_get_channels),required"`
	Body      string `json:"body" jsonschema:"description=plain-text message body,required"`
}

func sendMessage(ctx context.Context, c *nextdoor.Client, in SendMessageInput) (any, error) {
	msg, err := c.SendMessage(ctx, in.ChannelID, in.Body)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "message": msg}, nil
}

// DeleteMessageInput is the typed input for nextdoor_delete_message.
type DeleteMessageInput struct {
	MessageID string `json:"message_id" jsonschema:"description=ID of the DM message to delete,required"`
}

func deleteMessage(ctx context.Context, c *nextdoor.Client, in DeleteMessageInput) (any, error) {
	if err := c.DeleteMessage(ctx, in.MessageID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "message_id": in.MessageID}, nil
}

// GetChannelsInput is the typed input for nextdoor_get_channels.
type GetChannelsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"description=max channels to return,minimum=1,maximum=50,default=20"`
}

func getChannels(ctx context.Context, c *nextdoor.Client, in GetChannelsInput) (any, error) {
	res, err := c.GetChannels(ctx)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res, "", limit), nil
}

// GetMessagesInput is the typed input for nextdoor_get_messages.
type GetMessagesInput struct {
	ChannelID string `json:"channel_id" jsonschema:"description=Nextdoor DM channel ID,required"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max messages to return,minimum=1,maximum=50,default=20"`
}

func getMessages(ctx context.Context, c *nextdoor.Client, in GetMessagesInput) (any, error) {
	res, err := c.GetMessages(ctx, in.ChannelID)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res, "", limit), nil
}

var messagingTools = []mcptool.Tool{
	mcptool.Define[*nextdoor.Client, CreateChannelInput](
		"nextdoor_create_channel",
		"Create a Nextdoor DM channel with one or more legacy user IDs",
		"CreateChannel",
		createChannel,
	),
	mcptool.Define[*nextdoor.Client, SendMessageInput](
		"nextdoor_send_message",
		"Send a text message in a Nextdoor DM channel",
		"SendMessage",
		sendMessage,
	),
	mcptool.Define[*nextdoor.Client, DeleteMessageInput](
		"nextdoor_delete_message",
		"Delete a Nextdoor DM message by ID",
		"DeleteMessage",
		deleteMessage,
	),
	mcptool.Define[*nextdoor.Client, GetChannelsInput](
		"nextdoor_get_channels",
		"List the authenticated user's Nextdoor DM channels",
		"GetChannels",
		getChannels,
	),
	mcptool.Define[*nextdoor.Client, GetMessagesInput](
		"nextdoor_get_messages",
		"Fetch recent messages from a Nextdoor DM channel",
		"GetMessages",
		getMessages,
	),
}
