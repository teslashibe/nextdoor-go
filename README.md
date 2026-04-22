# nextdoor-go

Go client for Nextdoor's internal APIs. Uses cookie-based authentication with browser session cookies.

## Installation

```bash
go get github.com/teslashibe/nextdoor-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    nextdoor "github.com/teslashibe/nextdoor-go"
)

func main() {
    auth := nextdoor.Auth{
        CSRFToken:   "your-csrftoken-cookie",
        AccessToken: "your-ndbr_at-cookie",
        // Optional additional cookies:
        // DAID:      "...",
        // WE:        "...",
        // WE3P:      "...",
        // SessionID: "...",
    }

    c, err := nextdoor.New(auth)
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()

    // Get authenticated user
    me, err := c.GetMe(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Hello, %s from %s!\n", me.DisplayName, me.NeighborhoodName)

    // Fetch feed
    feed, err := c.GetFeed(ctx,
        nextdoor.WithOrderingMode(nextdoor.OrderRecentPosts),
        nextdoor.WithPageSize(5),
    )
    if err != nil {
        log.Fatal(err)
    }
    for _, p := range feed.Posts {
        fmt.Printf("- %s: %s\n", p.AuthorName, p.Subject)
    }
}
```

## Supported Operations

| Area | Methods |
|------|---------|
| **Profile** | `GetMe`, `GetProfile` |
| **Feed** | `GetFeed`, `GetFeedPage` |
| **Posts** | `GetPost`, `CreatePost`, `DeletePost`, `ReactToPost`, `RemoveReaction` |
| **Comments** | `GetComments`, `GetCommentsPage`, `CreateComment`, `DeleteComment` |
| **Messaging** | `CreateChannel`, `SendMessage`, `GetChannels`, `DeleteMessage` |
| **Search** | `SearchPosts`, `SearchNeighbors` |
| **Notifications** | `GetNotifications` |

## Authentication

Extract cookies from an authenticated Nextdoor browser session. Required cookies:

- `csrftoken` → `Auth.CSRFToken`
- `ndbr_at` → `Auth.AccessToken`

Optional cookies for full compatibility: `DAID`, `WE`, `WE3P`, `ndp_session_id`.

## Testing

Integration tests require a cookies file:

```bash
NEXTDOOR_COOKIES_FILE=./cookies.json go test -v -run TestGetMe
```

The cookies file should be a JSON array of `{"name": "...", "value": "..."}` objects (browser cookie export format).

## MCP support

This package ships an [MCP](https://modelcontextprotocol.io/) tool surface in `./mcp` for use with [`teslashibe/mcptool`](https://github.com/teslashibe/mcptool)-compatible hosts (e.g. [`teslashibe/agent-setup`](https://github.com/teslashibe/agent-setup)). 21 tools cover the full client API: profile (me/by-id), feed (list/paginate), posts (fetch/create/delete/react/unreact), comments (list/paginate/create/delete), messaging (channels list/create, messages list/send/delete), notifications, and post/neighbor search.

```go
import (
    "github.com/teslashibe/mcptool"
    nextdoor "github.com/teslashibe/nextdoor-go"
    ndmcp "github.com/teslashibe/nextdoor-go/mcp"
)

client, _ := nextdoor.New(nextdoor.Auth{...})
provider := ndmcp.Provider{}
for _, tool := range provider.Tools() {
    // register tool with your MCP server, passing client as the
    // opaque client argument when invoking
}
```

A coverage test in `mcp/mcp_test.go` fails if a new exported method is added to `*Client` without either being wrapped by an MCP tool or being added to `mcp.Excluded` with a reason — keeping the MCP surface in lockstep with the package API is enforced by CI rather than convention.

## Dependencies

Zero production dependencies — stdlib only. The `./mcp` subpackage adds [`teslashibe/mcptool`](https://github.com/teslashibe/mcptool) (and its single transitive `invopop/jsonschema` dep) for schema reflection.
