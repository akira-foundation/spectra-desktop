package app

import (
	"testing"
	"time"

	"spectra-desktop/internal/domain"
)

func TestAccountToDTO_BooleanFlagsReflectEncryptedPresence(t *testing.T) {
	now := time.Now()
	a := &App{}
	acc := domain.ProjectAccount{
		ID:              "id",
		ProjectID:       "p",
		Label:           "L",
		Kind:            domain.AccountKindBearer,
		PasswordEnc:     "x",
		APIKeyEnc:       "x",
		TokenEnc:        "x",
		RefreshTokenEnc: "x",
		TOTPSecretEnc:   "x",
		CookiesJSON:     `["c"]`,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	dto := a.accountToDTO(acc)
	if !dto.HasPassword || !dto.HasAPIKey || !dto.HasToken || !dto.HasRefreshToken || !dto.HasTOTP || !dto.HasCookies {
		t.Fatalf("flags: %+v", dto)
	}
	if dto.Kind != string(domain.AccountKindBearer) {
		t.Fatalf("kind: %q", dto.Kind)
	}
}

func TestAccountToDTO_EmptyEncryptedFieldsYieldFalseFlags(t *testing.T) {
	a := &App{}
	dto := a.accountToDTO(domain.ProjectAccount{})
	if dto.HasPassword || dto.HasAPIKey || dto.HasToken || dto.HasRefreshToken || dto.HasTOTP || dto.HasCookies {
		t.Fatalf("expected false flags: %+v", dto)
	}
}

func TestAccountToDTO_EmptyCookiesArrayCountsAsNoCookies(t *testing.T) {
	a := &App{}
	dto := a.accountToDTO(domain.ProjectAccount{CookiesJSON: "[]"})
	if dto.HasCookies {
		t.Fatal("[] should not count as cookies")
	}
}

func TestAccountToDTO_ParsesUserAndExtraHeaders(t *testing.T) {
	a := &App{}
	acc := domain.ProjectAccount{
		UserJSON:    `{"id":1,"email":"e@x"}`,
		HeadersJSON: `{"X-A":"1"}`,
	}
	dto := a.accountToDTO(acc)
	if dto.User["email"] != "e@x" {
		t.Fatalf("user: %v", dto.User)
	}
	if dto.ExtraHeaders["X-A"] != "1" {
		t.Fatalf("headers: %v", dto.ExtraHeaders)
	}
}

func TestAccountToDTO_ParsesOAuthConfig(t *testing.T) {
	a := &App{}
	acc := domain.ProjectAccount{
		OAuthConfigJSON: `{"grantType":"password","tokenURL":"https://x/token","clientID":"c","clientSecret":"s","scopes":["a","b"],"audience":"aud","username":"u"}`,
	}
	dto := a.accountToDTO(acc)
	if dto.OAuth == nil {
		t.Fatal("oauth nil")
	}
	if dto.OAuth.GrantType != "password" || dto.OAuth.TokenURL != "https://x/token" || dto.OAuth.ClientID != "c" {
		t.Fatalf("oauth fields: %+v", dto.OAuth)
	}
	if !dto.OAuth.HasSecret {
		t.Fatal("expected HasSecret true")
	}
	if len(dto.OAuth.Scopes) != 2 || dto.OAuth.Audience != "aud" || dto.OAuth.Username != "u" {
		t.Fatalf("oauth extras: %+v", dto.OAuth)
	}
}

func TestAccountToDTO_OAuthConfigInvalidJSONYieldsNil(t *testing.T) {
	a := &App{}
	dto := a.accountToDTO(domain.ProjectAccount{OAuthConfigJSON: "garbage"})
	if dto.OAuth != nil {
		t.Fatalf("expected nil oauth, got %+v", dto.OAuth)
	}
}
