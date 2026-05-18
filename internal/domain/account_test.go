package domain

import (
	"testing"
	"time"
)

func TestAccountKind_Constants(t *testing.T) {
	want := map[AccountKind]string{
		AccountKindBearer: "bearer",
		AccountKindBasic:  "basic",
		AccountKindAPIKey: "apikey",
		AccountKindOAuth2: "oauth2",
		AccountKindLogin:  "login",
	}
	for k, exp := range want {
		if string(k) != exp {
			t.Fatalf("kind %q != %q", string(k), exp)
		}
	}
}

func TestAPIKeyLocation_Constants(t *testing.T) {
	if string(APIKeyInHeader) != "header" || string(APIKeyInQuery) != "query" {
		t.Fatalf("api key location constants drifted")
	}
}

func TestProjectAccount_ZeroValue(t *testing.T) {
	var a ProjectAccount
	if a.ExpiresAt != nil {
		t.Fatalf("expected nil ExpiresAt")
	}
	if a.IsDefault || a.SortOrder != 0 {
		t.Fatalf("zero account not empty: %+v", a)
	}
}

func TestProjectAccount_WithExpires(t *testing.T) {
	now := time.Now().UTC()
	a := ProjectAccount{Kind: AccountKindBearer, ExpiresAt: &now}
	if a.ExpiresAt == nil || !a.ExpiresAt.Equal(now) {
		t.Fatalf("expires mismatch")
	}
}
