package nextdoor_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/teslashibe/nextdoor-go"
)

// cookieFile represents a single cookie entry from a browser cookie export.
type cookieFile struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func loadAuth(t *testing.T) nextdoor.Auth {
	t.Helper()
	path := os.Getenv("NEXTDOOR_COOKIES_FILE")
	if path == "" {
		t.Skip("NEXTDOOR_COOKIES_FILE not set")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading cookies file: %v", err)
	}

	var cookies []cookieFile
	if err := json.Unmarshal(raw, &cookies); err != nil {
		t.Fatalf("parsing cookies JSON: %v", err)
	}

	var auth nextdoor.Auth
	for _, c := range cookies {
		switch c.Name {
		case "csrftoken":
			auth.CSRFToken = c.Value
		case "ndbr_at":
			auth.AccessToken = c.Value
		case "DAID":
			auth.DAID = c.Value
		case "WE":
			auth.WE = c.Value
		case "WE3P":
			auth.WE3P = c.Value
		case "ndp_session_id":
			auth.SessionID = c.Value
		}
	}

	if auth.CSRFToken == "" || auth.AccessToken == "" {
		t.Fatal("cookies file missing csrftoken or ndbr_at")
	}
	return auth
}

func TestGetMe(t *testing.T) {
	auth := loadAuth(t)
	c := nextdoor.New(auth)
	ctx := context.Background()

	me, err := c.GetMe(ctx)
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}

	if me.ID == "" {
		t.Error("expected non-empty ID")
	}
	if me.DisplayName == "" {
		t.Error("expected non-empty DisplayName")
	}
	t.Logf("me: %+v", me)
}

func TestGetFeed(t *testing.T) {
	auth := loadAuth(t)
	c := nextdoor.New(auth)
	ctx := context.Background()

	page, err := c.GetFeed(ctx, nextdoor.WithPageSize(3))
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}

	if len(page.Posts) == 0 {
		t.Error("expected at least one post")
	}
	for i, p := range page.Posts {
		t.Logf("post[%d]: id=%s author=%s subject=%q", i, p.ID, p.AuthorName, p.Subject)
	}
}

func TestGetComments(t *testing.T) {
	auth := loadAuth(t)
	c := nextdoor.New(auth)
	ctx := context.Background()

	page, err := c.GetFeed(ctx, nextdoor.WithPageSize(5))
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}

	if len(page.Posts) == 0 {
		t.Skip("no posts in feed")
	}
	postID := page.Posts[0].ID

	cp, err := c.GetComments(ctx, postID)
	if err != nil {
		t.Fatalf("GetComments(%s): %v", postID, err)
	}

	t.Logf("comments for post %s: %d", postID, len(cp.Comments))
	for i, cm := range cp.Comments {
		t.Logf("comment[%d]: id=%s author=%s body=%q", i, cm.ID, cm.AuthorName, cm.Body)
	}
}

func TestSearchPosts(t *testing.T) {
	auth := loadAuth(t)
	c := nextdoor.New(auth)
	ctx := context.Background()

	results, err := c.SearchPosts(ctx, "neighborhood")
	if err != nil {
		t.Fatalf("SearchPosts: %v", err)
	}

	t.Logf("search returned %d results", len(results))
	for i, r := range results {
		if i >= 3 {
			break
		}
		t.Logf("result[%d]: id=%s title=%q", i, r.ID, r.Title)
	}
}
