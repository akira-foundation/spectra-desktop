package auth

import (
	"context"
	"encoding/base64"
	"testing"

	"spectra-desktop/internal/domain"
)

func TestResolver_Resolve_NilAccount(t *testing.T) {
	r := NewResolver(nil, newVault(t))
	inj, err := r.Resolve(context.Background(), nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if inj != (HeaderInjection{}) {
		t.Fatalf("expected zero injection, got %+v", inj)
	}
}

func TestResolver_Resolve_UnknownKindReturnsZero(t *testing.T) {
	r := NewResolver(nil, newVault(t))
	acc := &domain.ProjectAccount{Kind: domain.AccountKind("nope")}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if inj != (HeaderInjection{}) {
		t.Fatalf("expected zero injection, got %+v", inj)
	}
}

func TestResolver_Bearer_DefaultScheme(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{Kind: domain.AccountKindBearer, TokenEnc: encrypt(t, v, "tok-123")}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Header != "Authorization" || inj.Value != "Bearer tok-123" {
		t.Fatalf("got %+v", inj)
	}
}

func TestResolver_Bearer_CustomScheme(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:     domain.AccountKindLogin,
		Scheme:   "Token",
		TokenEnc: encrypt(t, v, "abc"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Token abc" {
		t.Fatalf("value = %q", inj.Value)
	}
}

func TestResolver_Bearer_EmptyTokenReturnsZero(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{Kind: domain.AccountKindBearer}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj != (HeaderInjection{}) {
		t.Fatalf("expected zero, got %+v", inj)
	}
}

func TestResolver_Bearer_DecryptError(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{Kind: domain.AccountKindBearer, TokenEnc: "v1:not-base64!!!"}
	_, err := r.Resolve(context.Background(), acc)
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestResolver_Basic_Encoding(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:        domain.AccountKindBasic,
		Username:    "alice",
		PasswordEnc: encrypt(t, v, "wonder"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:wonder"))
	if inj.Header != "Authorization" || inj.Value != want {
		t.Fatalf("got %+v want %s", inj, want)
	}
}

func TestResolver_Basic_EmptyBothReturnsZero(t *testing.T) {
	r := NewResolver(nil, newVault(t))
	acc := &domain.ProjectAccount{Kind: domain.AccountKindBasic}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj != (HeaderInjection{}) {
		t.Fatalf("expected zero, got %+v", inj)
	}
}

func TestResolver_APIKey_HeaderDefault(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:      domain.AccountKindAPIKey,
		APIKeyEnc: encrypt(t, v, "k-1"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Header != "X-API-Key" || inj.Value != "k-1" {
		t.Fatalf("got %+v", inj)
	}
	if inj.QueryKey != "" {
		t.Fatalf("unexpected query key: %q", inj.QueryKey)
	}
}

func TestResolver_APIKey_HeaderCustom(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:         domain.AccountKindAPIKey,
		APIKeyHeader: "X-My-Key",
		APIKeyEnc:    encrypt(t, v, "k-2"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Header != "X-My-Key" || inj.Value != "k-2" {
		t.Fatalf("got %+v", inj)
	}
}

func TestResolver_APIKey_QueryDefault(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:      domain.AccountKindAPIKey,
		APIKeyIn:  domain.APIKeyInQuery,
		APIKeyEnc: encrypt(t, v, "k-3"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.QueryKey != "api_key" || inj.QueryValue != "k-3" {
		t.Fatalf("got %+v", inj)
	}
	if inj.Header != "" {
		t.Fatalf("unexpected header: %q", inj.Header)
	}
}

func TestResolver_APIKey_QueryCustomName(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		Kind:         domain.AccountKindAPIKey,
		APIKeyIn:     domain.APIKeyInQuery,
		APIKeyHeader: "token",
		APIKeyEnc:    encrypt(t, v, "k-4"),
	}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.QueryKey != "token" || inj.QueryValue != "k-4" {
		t.Fatalf("got %+v", inj)
	}
}

func TestResolver_APIKey_EmptyReturnsZero(t *testing.T) {
	r := NewResolver(nil, newVault(t))
	acc := &domain.ProjectAccount{Kind: domain.AccountKindAPIKey}
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj != (HeaderInjection{}) {
		t.Fatalf("expected zero, got %+v", inj)
	}
}
