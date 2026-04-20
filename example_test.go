package nextdoor_test

import (
	"context"
	"fmt"
	"log"

	"github.com/teslashibe/nextdoor-go"
)

func Example() {
	auth := nextdoor.Auth{
		CSRFToken:   "your-csrf-token",
		AccessToken: "your-ndbr-at-token",
	}

	c, err := nextdoor.New(auth)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	me, err := c.GetMe(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Hello, %s from %s!\n", me.DisplayName, me.NeighborhoodName)

	feed, err := c.GetFeed(ctx,
		nextdoor.WithOrderingMode(nextdoor.OrderRecentPosts),
		nextdoor.WithPageSize(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range feed.Posts {
		fmt.Printf("- [%s] %s\n", p.AuthorName, p.Subject)
	}
}

func ExampleClient_SearchPosts() {
	auth := nextdoor.Auth{
		CSRFToken:   "your-csrf-token",
		AccessToken: "your-ndbr-at-token",
	}

	c, err := nextdoor.New(auth)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	results, err := c.SearchPosts(ctx, "lost dog")
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range results {
		fmt.Printf("%s: %s\n", r.Type, r.Title)
	}
}
