// Package mcp exposes the nextdoor-go [nextdoor.Client] surface as a set of
// MCP (Model Context Protocol) tools that any host application can mount on
// its own MCP server.
//
// All tools wrap exported methods on *nextdoor.Client. Each tool is defined
// via [mcptool.Define] so the JSON input schema is reflected from the typed
// input struct — no hand-maintained schemas, no drift.
//
// Usage from a host application:
//
//	import (
//	    "github.com/teslashibe/mcptool"
//	    nextdoor "github.com/teslashibe/nextdoor-go"
//	    ndmcp "github.com/teslashibe/nextdoor-go/mcp"
//	)
//
//	client, _ := nextdoor.New(nextdoor.Auth{...})
//	for _, tool := range ndmcp.Provider{}.Tools() {
//	    // register tool with your MCP server, passing client as the client arg
//	    // when invoking
//	}
//
// The [Excluded] map documents methods on *Client that are intentionally not
// exposed via MCP, with a one-line reason. The coverage test in mcp_test.go
// fails if a new exported method is added without either being wrapped by a
// tool or appearing in [Excluded].
package mcp

import "github.com/teslashibe/mcptool"

// Provider implements [mcptool.Provider] for nextdoor-go. The zero value is
// ready to use.
type Provider struct{}

// Platform returns "nextdoor".
func (Provider) Platform() string { return "nextdoor" }

// Tools returns every nextdoor-go MCP tool, in registration order.
func (Provider) Tools() []mcptool.Tool {
	out := make([]mcptool.Tool, 0,
		len(profileTools)+
			len(feedTools)+
			len(postTools)+
			len(commentTools)+
			len(messagingTools)+
			len(notificationTools)+
			len(searchTools),
	)
	out = append(out, profileTools...)
	out = append(out, feedTools...)
	out = append(out, postTools...)
	out = append(out, commentTools...)
	out = append(out, messagingTools...)
	out = append(out, notificationTools...)
	out = append(out, searchTools...)
	return out
}
