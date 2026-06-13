package nextdoor

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

// Credential login for Nextdoor (#269).
//
// Reverse-engineered from the live web login (Playwright capture): Nextdoor
// authenticates with an OAuth2 password grant, then exchanges the resulting
// tokens for a web session.
//
//  1. GET https://nextdoor.com/login/            -> seeds the csrftoken cookie
//  2. POST https://auth.nextdoor.com/v2/token    -> {access_token, id_token}
//     (form: grant_type=password, client_id=NEXTDOOR-WEB, login_type=basic,
//      scope=openid, username, password, state)
//  3. POST https://nextdoor.com/session/         -> sets ndbr_at + csrftoken
//     (JSON {access_token, id_token}; requires X-CSRFToken header from step 1)
//
// The minted csrftoken + ndbr_at cookies feed straight into New(Auth{...}).
//
// Nextdoor is geofenced to the account's region; route the login through a
// residential egress in the account's country (ProxyURL) so it isn't blocked.

const (
	ndLoginPageURL = "https://nextdoor.com/login/"
	ndTokenURL     = "https://auth.nextdoor.com/v2/token"
	ndSessionURL   = "https://nextdoor.com/session/"
	ndClientID     = "NEXTDOOR-WEB"
	ndUserAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
	ndSessionMime  = "application/vnd.nextdoor/schema/session-request.v1+json"
)

// LoginParams holds Nextdoor credentials.
type LoginParams struct {
	Email    string
	Password string
	// UserAgent overrides the browser UA used for the login + minted session.
	UserAgent string
	// ProxyURL routes the login through an HTTP/S proxy (residential egress).
	ProxyURL string
	// OTP is an email verification code, used when Nextdoor requires one
	// (new device/IP -> auth_code_required).
	OTP string
	// VerificationProvider fetches the email OTP on demand (e.g. polls Gmail)
	// when Nextdoor requires one and no static OTP was supplied.
	VerificationProvider func(ctx context.Context) (string, error)
}

// ErrVerificationRequired is returned when Nextdoor requires an email OTP and
// none was supplied (neither OTP nor VerificationProvider yielded a code).
var ErrVerificationRequired = fmt.Errorf("nextdoor: email verification code required (retry with OTP)")

// LoginResult is the minted session, ready for New.
type LoginResult struct {
	Auth      Auth
	UserAgent string
}

// Login authenticates with an email + password and returns minted session
// cookies (csrftoken, ndbr_at). It does not require pre-existing cookies.
func Login(ctx context.Context, p LoginParams) (*LoginResult, error) {
	if strings.TrimSpace(p.Email) == "" || p.Password == "" {
		return nil, fmt.Errorf("nextdoor: email and password are required")
	}
	ua := p.UserAgent
	if ua == "" {
		ua = ndUserAgent
	}
	debug := strings.TrimSpace(os.Getenv("NEXTDOOR_LOGIN_DEBUG")) != ""

	jar, _ := cookiejar.New(nil)
	hc := &http.Client{Jar: jar, Timeout: 30 * time.Second}
	if p.ProxyURL != "" {
		if parsed, err := url.Parse(p.ProxyURL); err == nil {
			tr := http.DefaultTransport.(*http.Transport).Clone()
			tr.Proxy = http.ProxyURL(parsed)
			hc.Transport = tr
		}
	}

	// 1. Seed the csrftoken + DAID cookies.
	if err := ndGet(ctx, hc, ndLoginPageURL, ua); err != nil {
		return nil, fmt.Errorf("nextdoor: load login page: %w", err)
	}

	// The auth server ties the password grant to a device anonymous id: the
	// web client sends the DAID cookie as both the `state` param and the
	// `session-daid` header, plus a device-id and activity-id. Reuse the DAID
	// cookie when present (set by GET /login/), else mint one.
	daid := ndCookie(jar, "DAID")
	if daid == "" {
		daid = ndState()
	}
	deviceID := ndState()
	activityID := ndUUID()

	hdrs := map[string]string{
		"Content-Type":     "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":       ua,
		"Origin":           "https://nextdoor.com",
		"Referer":          "https://nextdoor.com/",
		"Accept":           "application/json, text/plain, */*",
		"session-daid":     daid,
		"device-id":        deviceID,
		"x-nd-activity-id": activityID,
		"device-fp":        "v2" + ndHex(32) + ",v6" + deviceID + ",",
	}

	// 2. OAuth2 password grant. From a new device/IP Nextdoor responds with
	//    auth_code_required and emails an OTP; re-submit with the auth_code.
	tok, status, body, err := ndTokenGrant(ctx, hc, p, daid, "", hdrs)
	if err != nil {
		return nil, err
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[nextdoor-login-debug] token status=%d body=%s\n", status, ndTruncate(body, 300))
	}
	if status == http.StatusBadRequest && strings.Contains(body, "auth_code_required") {
		code := strings.TrimSpace(p.OTP)
		if code == "" && p.VerificationProvider != nil {
			c, perr := p.VerificationProvider(ctx)
			if perr != nil {
				return nil, fmt.Errorf("nextdoor: fetch verification code: %w", perr)
			}
			code = strings.TrimSpace(c)
		}
		if code == "" {
			return nil, ErrVerificationRequired
		}
		tok, status, body, err = ndTokenGrant(ctx, hc, p, daid, code, hdrs)
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[nextdoor-login-debug] token(otp) status=%d body=%s\n", status, ndTruncate(body, 300))
		}
	}
	switch {
	case status == http.StatusUnauthorized:
		return nil, fmt.Errorf("nextdoor: wrong email or password")
	case status == http.StatusBadRequest && (strings.Contains(body, "invalid") && strings.Contains(body, "code")):
		return nil, fmt.Errorf("nextdoor: invalid or expired verification code")
	case status != http.StatusOK:
		return nil, fmt.Errorf("nextdoor: token endpoint status %d: %s", status, ndTruncate(body, 200))
	}
	if tok.AccessToken == "" || tok.IDToken == "" {
		return nil, fmt.Errorf("nextdoor: token response missing access_token/id_token")
	}

	// 3. Exchange tokens for a web session (sets ndbr_at + csrftoken).
	sessBody, _ := json.Marshal(map[string]string{"access_token": tok.AccessToken, "id_token": tok.IDToken})
	sessReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ndSessionURL, strings.NewReader(string(sessBody)))
	if err != nil {
		return nil, err
	}
	sessReq.Header.Set("Content-Type", ndSessionMime)
	sessReq.Header.Set("User-Agent", ua)
	sessReq.Header.Set("Origin", "https://nextdoor.com")
	sessReq.Header.Set("Referer", ndLoginPageURL)
	sessReq.Header.Set("Accept", "application/json")
	if csrf := ndCookie(jar, "csrftoken"); csrf != "" {
		sessReq.Header.Set("X-CSRFToken", csrf)
	}
	sessResp, err := hc.Do(sessReq)
	if err != nil {
		return nil, fmt.Errorf("nextdoor: session request: %w", err)
	}
	sessRaw, _ := io.ReadAll(io.LimitReader(sessResp.Body, 1<<20))
	sessResp.Body.Close()
	if debug {
		fmt.Fprintf(os.Stderr, "[nextdoor-login-debug] session status=%d\n", sessResp.StatusCode)
	}
	if sessResp.StatusCode != http.StatusOK && sessResp.StatusCode != http.StatusCreated && sessResp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("nextdoor: session exchange status %d: %s", sessResp.StatusCode, ndTruncate(string(sessRaw), 200))
	}

	auth := Auth{
		CSRFToken:   ndCookie(jar, "csrftoken"),
		AccessToken: ndCookie(jar, "ndbr_at"),
	}
	if auth.AccessToken == "" || auth.CSRFToken == "" {
		return nil, fmt.Errorf("nextdoor: login did not establish a session (csrftoken/ndbr_at missing)")
	}
	return &LoginResult{Auth: auth, UserAgent: ua}, nil
}

type ndToken struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

// ndTokenGrant performs the OAuth2 password grant. When authCode is non-empty
// it is included (the email-OTP step). Returns the decoded token, HTTP status,
// and raw body.
func ndTokenGrant(ctx context.Context, hc *http.Client, p LoginParams, daid, authCode string, hdrs map[string]string) (ndToken, int, string, error) {
	form := url.Values{
		"scope":      {"openid"},
		"client_id":  {ndClientID},
		"login_type": {"basic"},
		"grant_type": {"password"},
		"username":   {p.Email},
		"password":   {p.Password},
		"state":      {daid},
	}
	if authCode != "" {
		form.Set("auth_code", authCode)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ndTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ndToken{}, 0, "", err
	}
	for k, v := range hdrs {
		req.Header.Set(k, v)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return ndToken{}, 0, "", fmt.Errorf("nextdoor: token request: %w", err)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	var tok ndToken
	_ = json.Unmarshal(body, &tok)
	return tok, resp.StatusCode, string(body), nil
}

func ndGet(ctx context.Context, hc *http.Client, target, ua string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	return nil
}

func ndCookie(jar http.CookieJar, name string) string {
	for _, host := range []string{"https://nextdoor.com", "https://auth.nextdoor.com"} {
		u, err := url.Parse(host)
		if err != nil {
			continue
		}
		for _, ck := range jar.Cookies(u) {
			if ck.Name == name && ck.Value != "" {
				return ck.Value
			}
		}
	}
	return ""
}

// ndState builds the OAuth `state` param: a random UUIDv4 plus a short
// date suffix, mirroring the web client's format.
func ndState() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return uuid + time.Now().Format("060102")
}

// ndUUID returns a random UUIDv4.
func ndUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ndHex returns n random lowercase hex characters.
func ndHex(n int) string {
	b := make([]byte, (n+1)/2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:n]
}

func ndTruncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
