package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFetchOAuth2Token_ClientCredentials(t *testing.T) {
	var (
		gotForm   url.Values
		gotMethod string
		gotCT     string
		gotAccept string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		gotForm = r.PostForm
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "abc.def",
			"refresh_token": "r-1",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()

	cfg := OAuth2Config{
		TokenURL:     srv.URL,
		ClientID:     "cid",
		ClientSecret: "csec",
		Scopes:       []string{"read", "write"},
		Audience:     "api://x",
	}
	tok, err := FetchOAuth2Token(context.Background(), cfg, "", "")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if tok.AccessToken != "abc.def" || tok.RefreshToken != "r-1" {
		t.Fatalf("token mismatch: %+v", tok)
	}
	if tok.ExpiresAt.IsZero() || time.Until(tok.ExpiresAt) <= 0 {
		t.Fatalf("expires_at not populated: %v", tok.ExpiresAt)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s", gotMethod)
	}
	if gotCT != "application/x-www-form-urlencoded" {
		t.Fatalf("content-type = %q", gotCT)
	}
	if gotAccept != "application/json" {
		t.Fatalf("accept = %q", gotAccept)
	}
	if gotForm.Get("grant_type") != "client_credentials" {
		t.Fatalf("grant_type = %q", gotForm.Get("grant_type"))
	}
	if gotForm.Get("client_id") != "cid" || gotForm.Get("client_secret") != "csec" {
		t.Fatalf("client creds missing: %+v", gotForm)
	}
	if gotForm.Get("scope") != "read write" {
		t.Fatalf("scope = %q", gotForm.Get("scope"))
	}
	if gotForm.Get("audience") != "api://x" {
		t.Fatalf("audience = %q", gotForm.Get("audience"))
	}
}

func TestFetchOAuth2Token_PasswordGrant(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		_, _ = w.Write([]byte(`{"access_token":"pw-tok"}`))
	}))
	defer srv.Close()

	cfg := OAuth2Config{
		TokenURL:  srv.URL,
		GrantType: "password",
		Username:  "alice",
	}
	tok, err := FetchOAuth2Token(context.Background(), cfg, "", "s3cret")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if tok.AccessToken != "pw-tok" {
		t.Fatalf("access = %q", tok.AccessToken)
	}
	if gotForm.Get("grant_type") != "password" {
		t.Fatalf("grant_type = %q", gotForm.Get("grant_type"))
	}
	if gotForm.Get("username") != "alice" || gotForm.Get("password") != "s3cret" {
		t.Fatalf("user/pwd missing: %+v", gotForm)
	}
}

func TestFetchOAuth2Token_RefreshTokenTakesPriority(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		_, _ = w.Write([]byte(`{"access_token":"new-tok","refresh_token":"new-rt","expires_in":60}`))
	}))
	defer srv.Close()

	cfg := OAuth2Config{TokenURL: srv.URL, GrantType: "password", Username: "u"}
	tok, err := FetchOAuth2Token(context.Background(), cfg, "old-rt", "pw")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if tok.AccessToken != "new-tok" {
		t.Fatalf("token = %q", tok.AccessToken)
	}
	if gotForm.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type = %q", gotForm.Get("grant_type"))
	}
	if gotForm.Get("refresh_token") != "old-rt" {
		t.Fatalf("refresh_token = %q", gotForm.Get("refresh_token"))
	}
	if gotForm.Get("username") != "" || gotForm.Get("password") != "" {
		t.Fatalf("password fields leaked: %+v", gotForm)
	}
}

func TestFetchOAuth2Token_MissingTokenURL(t *testing.T) {
	_, err := FetchOAuth2Token(context.Background(), OAuth2Config{}, "", "")
	if err == nil || !strings.Contains(err.Error(), "token URL") {
		t.Fatalf("expected token URL error, got %v", err)
	}
}

func TestFetchOAuth2Token_HTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer srv.Close()

	_, err := FetchOAuth2Token(context.Background(), OAuth2Config{TokenURL: srv.URL}, "", "")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 error, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("expected error body in message, got %v", err)
	}
}

func TestFetchOAuth2Token_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	_, err := FetchOAuth2Token(context.Background(), OAuth2Config{TokenURL: srv.URL}, "", "")
	if err == nil || !strings.Contains(err.Error(), "parse token response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestFetchOAuth2Token_MissingAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	_, err := FetchOAuth2Token(context.Background(), OAuth2Config{TokenURL: srv.URL}, "", "")
	if err == nil || !strings.Contains(err.Error(), "missing access_token") {
		t.Fatalf("expected missing access_token error, got %v", err)
	}
}

func TestFetchOAuth2Token_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"x"}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := FetchOAuth2Token(ctx, OAuth2Config{TokenURL: srv.URL}, "", "")
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestTruncate_ShortAndLong(t *testing.T) {
	if got := truncate("hi", 10); got != "hi" {
		t.Fatalf("short = %q", got)
	}
	if got := truncate("abcdefghij", 3); got != "abc..." {
		t.Fatalf("long = %q", got)
	}
}
