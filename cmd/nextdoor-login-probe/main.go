// Command nextdoor-login-probe smoke-tests Nextdoor email/password login end
// to end: it logs in, mints csrftoken + ndbr_at, builds a client, and calls
// GetMe for liveness. Credentials come from env:
//
//	NEXTDOOR_EMAIL, NEXTDOOR_PASSWORD, NEXTDOOR_PROXY_URL (optional)
//
// Exit 0 on success, non-zero otherwise.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	nextdoor "github.com/teslashibe/nextdoor-go"
)

func main() {
	email := strings.TrimSpace(os.Getenv("NEXTDOOR_EMAIL"))
	pass := os.Getenv("NEXTDOOR_PASSWORD")
	if email == "" || pass == "" {
		fmt.Fprintln(os.Stderr, "nextdoor-login-probe: set NEXTDOOR_EMAIL and NEXTDOOR_PASSWORD")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	params := nextdoor.LoginParams{
		Email:    email,
		Password: pass,
		ProxyURL: strings.TrimSpace(os.Getenv("NEXTDOOR_PROXY_URL")),
		OTP:      strings.TrimSpace(os.Getenv("NEXTDOOR_OTP")),
	}
	if params.OTP == "" {
		if codeFile := strings.TrimSpace(os.Getenv("NEXTDOOR_CODE_FILE")); codeFile != "" {
			params.VerificationProvider = func(ctx context.Context) (string, error) {
				deadline := time.Now().Add(100 * time.Second)
				for time.Now().Before(deadline) {
					if b, err := os.ReadFile(codeFile); err == nil {
						if c := strings.TrimSpace(string(b)); c != "" {
							return c, nil
						}
					}
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(3 * time.Second):
					}
				}
				return "", fmt.Errorf("timed out waiting for code in %s", codeFile)
			}
		}
	}

	res, err := nextdoor.Login(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "login: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("login ok: csrftoken_len=%d ndbr_at_len=%d\n", len(res.Auth.CSRFToken), len(res.Auth.AccessToken))

	c, err := nextdoor.New(res.Auth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new client: %v\n", err)
		os.Exit(1)
	}
	me, err := c.GetMe(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetMe(): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("liveness ok: %+v\n", me)
}
