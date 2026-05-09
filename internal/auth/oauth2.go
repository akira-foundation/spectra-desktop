package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuth2Config struct {
	GrantType    string   `json:"grantType"`
	TokenURL     string   `json:"tokenUrl"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Scopes       []string `json:"scopes,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	Username     string   `json:"username,omitempty"`
}

type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"-"`
}

func FetchOAuth2Token(ctx context.Context, cfg OAuth2Config, refreshToken string, password string) (*OAuth2Token, error) {
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("oauth2: token URL is required")
	}

	form := url.Values{}
	switch {
	case refreshToken != "":
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", refreshToken)
	case strings.EqualFold(cfg.GrantType, "password"):
		form.Set("grant_type", "password")
		form.Set("username", cfg.Username)
		form.Set("password", password)
	default:
		form.Set("grant_type", "client_credentials")
	}
	if cfg.ClientID != "" {
		form.Set("client_id", cfg.ClientID)
	}
	if cfg.ClientSecret != "" {
		form.Set("client_secret", cfg.ClientSecret)
	}
	if len(cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	if cfg.Audience != "" {
		form.Set("audience", cfg.Audience)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("oauth2: token endpoint returned %d: %s", resp.StatusCode, truncate(string(body), 256))
	}

	var token OAuth2Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("oauth2: parse token response: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("oauth2: response missing access_token")
	}
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return &token, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
