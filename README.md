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

    c := nextdoor.New(auth)
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
        fmt.Printf("- %s: %s\n", p.AuthorName, p.Title)
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

## Dependencies

Zero production dependencies — stdlib only.
