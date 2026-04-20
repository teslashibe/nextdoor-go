package nextdoor_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/teslashibe/nextdoor-go"
)

type cookieEntry struct {
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

	var cookies []cookieEntry
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

func newClient(t *testing.T) *nextdoor.Client {
	t.Helper()
	auth := loadAuth(t)
	c, err := nextdoor.New(auth)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewValidatesAuth(t *testing.T) {
	_, err := nextdoor.New(nextdoor.Auth{})
	if err == nil {
		t.Fatal("expected error for empty auth")
	}

	_, err = nextdoor.New(nextdoor.Auth{CSRFToken: "x"})
	if err == nil {
		t.Fatal("expected error for missing AccessToken")
	}

	c, err := nextdoor.New(nextdoor.Auth{CSRFToken: "x", AccessToken: "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetMe(t *testing.T) {
	c := newClient(t)
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
	c := newClient(t)
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
	c := newClient(t)
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

	t.Logf("comments for post %s: %d (total: %d)", postID, len(cp.Comments), cp.TotalCount)
	for i, cm := range cp.Comments {
		t.Logf("comment[%d]: id=%s author=%s body=%q", i, cm.ID, cm.AuthorName, cm.Body)
	}
}

func TestCreateAndDeletePost(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	post, err := c.CreatePost(ctx, "Automated test from nextdoor-go library — will be deleted in a few seconds. Please ignore!")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	if post.ID == "" {
		t.Fatal("expected non-empty post ID")
	}
	t.Logf("created post: id=%s subject=%q", post.ID, post.Subject)

	err = c.DeletePost(ctx, post.ID)
	if err != nil {
		t.Fatalf("DeletePost(%s): %v", post.ID, err)
	}
	t.Logf("deleted post: %s", post.ID)
}

func TestReactAndRemoveReaction(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	page, err := c.GetFeed(ctx, nextdoor.WithPageSize(3))
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(page.Posts) == 0 {
		t.Skip("no posts in feed")
	}

	postID := page.Posts[0].ID
	t.Logf("reacting to post %s", postID)

	reactionID, err := c.ReactToPost(ctx, postID, nextdoor.ReactionLike)
	if err != nil {
		t.Fatalf("ReactToPost: %v", err)
	}
	t.Logf("reaction added, reactionID=%s", reactionID)

	if reactionID != "" {
		err = c.RemoveReaction(ctx, reactionID)
		if err != nil {
			t.Fatalf("RemoveReaction(%s): %v", reactionID, err)
		}
		t.Logf("reaction removed: %s", reactionID)
	}
}

func TestSearchPosts(t *testing.T) {
	c := newClient(t)
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
		t.Logf("result[%d]: id=%s title=%q url=%s", i, r.ID, r.Title, r.URL)
	}
}

func TestSearchNeighbors(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	results, err := c.SearchNeighbors(ctx, "John")
	if err != nil {
		t.Fatalf("SearchNeighbors: %v", err)
	}

	t.Logf("search returned %d results", len(results))
	for i, r := range results {
		if i >= 3 {
			break
		}
		t.Logf("result[%d]: id=%s title=%q url=%s", i, r.ID, r.Title, r.URL)
	}
}

func TestSendMessage(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	userID := os.Getenv("NEXTDOOR_TEST_USER_ID")
	if userID == "" {
		t.Skip("NEXTDOOR_TEST_USER_ID not set")
	}

	ch, err := c.CreateChannel(ctx, []string{userID})
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	t.Logf("channel created: %s", ch.ID)

	msg, err := c.SendMessage(ctx, ch.ID, "Test message from nextdoor-go library")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	t.Logf("message sent: id=%s body=%q", msg.ID, msg.Body)
}

func TestGetNotificationsStub(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	_, err := c.GetNotifications(ctx)
	if err == nil {
		t.Fatal("expected error from stub GetNotifications")
	}
	t.Logf("GetNotifications correctly returned error: %v", err)
}
