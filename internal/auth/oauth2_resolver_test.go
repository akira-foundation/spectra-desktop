package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"spectra-desktop/internal/domain"
)

func TestResolver_OAuth2_ValidTokenSkipsRefresh(t *testing.T) {
	v := newVault(t)
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("token endpoint should not be called when token is fresh")
	}))
	defer srv.Close()

	future := time.Now().UTC().Add(10 * time.Minute)
	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL, ClientID: "c"})
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		TokenEnc:        encrypt(t, v, "live-tok"),
		ExpiresAt:       &future,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(newFakeAccountRepo(), v)
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Bearer live-tok" {
		t.Fatalf("value = %q", inj.Value)
	}
}

func TestResolver_OAuth2_RefreshOnExpiry(t *testing.T) {
	v := newVault(t)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "refresh_token" {
			t.Fatalf("expected refresh_token grant, got %q", r.PostForm.Get("grant_type"))
		}
		if r.PostForm.Get("refresh_token") != "rt-old" {
			t.Fatalf("refresh token = %q", r.PostForm.Get("refresh_token"))
		}
		_, _ = w.Write([]byte(`{"access_token":"rotated","refresh_token":"rt-new","expires_in":3600}`))
	}))
	defer srv.Close()

	past := time.Now().UTC().Add(-time.Minute)
	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL, ClientID: "cid"})
	repo := newFakeAccountRepo()
	acc := &domain.ProjectAccount{
		ID:              "a1",
		Kind:            domain.AccountKindOAuth2,
		TokenEnc:        encrypt(t, v, "stale"),
		RefreshTokenEnc: encrypt(t, v, "rt-old"),
		ExpiresAt:       &past,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(repo, v)
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Bearer rotated" {
		t.Fatalf("value = %q", inj.Value)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("calls = %d", calls)
	}
	if repo.saveCount() != 1 {
		t.Fatalf("saves = %d", repo.saveCount())
	}
	stored, _ := repo.Get(context.Background(), "a1")
	if stored == nil {
		t.Fatal("expected account persisted")
	}
	newToken, err := v.Decrypt(stored.TokenEnc)
	if err != nil || newToken != "rotated" {
		t.Fatalf("stored token = %q err=%v", newToken, err)
	}
	newRT, _ := v.Decrypt(stored.RefreshTokenEnc)
	if newRT != "rt-new" {
		t.Fatalf("stored refresh = %q", newRT)
	}
	if stored.ExpiresAt == nil || stored.ExpiresAt.Before(time.Now().UTC()) {
		t.Fatalf("expires_at not updated: %v", stored.ExpiresAt)
	}
}

func TestResolver_OAuth2_RefreshFailureWithExistingTokenFallsBack(t *testing.T) {
	v := newVault(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	past := time.Now().UTC().Add(-time.Minute)
	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL})
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		TokenEnc:        encrypt(t, v, "still-here"),
		ExpiresAt:       &past,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(newFakeAccountRepo(), v)
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Bearer still-here" {
		t.Fatalf("value = %q", inj.Value)
	}
}

func TestResolver_OAuth2_RefreshFailureNoExistingTokenReturnsError(t *testing.T) {
	v := newVault(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL})
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(newFakeAccountRepo(), v)
	_, err := r.Resolve(context.Background(), acc)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected 400 in error, got %v", err)
	}
}

func TestResolver_OAuth2_MissingConfig(t *testing.T) {
	v := newVault(t)
	acc := &domain.ProjectAccount{
		Kind:  domain.AccountKindOAuth2,
		Label: "demo",
	}
	r := NewResolver(newFakeAccountRepo(), v)
	_, err := r.Resolve(context.Background(), acc)
	if err == nil || !strings.Contains(err.Error(), "missing config") {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func TestResolver_OAuth2_InvalidConfigJSON(t *testing.T) {
	v := newVault(t)
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		OAuthConfigJSON: "{not json",
	}
	r := NewResolver(newFakeAccountRepo(), v)
	_, err := r.Resolve(context.Background(), acc)
	if err == nil || !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("expected parse config error, got %v", err)
	}
}

func TestResolver_OAuth2_NearExpiryTriggersRefresh(t *testing.T) {
	v := newVault(t)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"access_token":"fresh","expires_in":60}`))
	}))
	defer srv.Close()

	soon := time.Now().UTC().Add(10 * time.Second)
	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL})
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		TokenEnc:        encrypt(t, v, "old"),
		ExpiresAt:       &soon,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(newFakeAccountRepo(), v)
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Bearer fresh" {
		t.Fatalf("value = %q", inj.Value)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestResolver_OAuth2_NoRepoStillReturnsToken(t *testing.T) {
	v := newVault(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"norepo"}`))
	}))
	defer srv.Close()

	cfg, _ := json.Marshal(OAuth2Config{TokenURL: srv.URL})
	acc := &domain.ProjectAccount{
		Kind:            domain.AccountKindOAuth2,
		OAuthConfigJSON: string(cfg),
	}
	r := NewResolver(nil, v)
	inj, err := r.Resolve(context.Background(), acc)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if inj.Value != "Bearer norepo" {
		t.Fatalf("value = %q", inj.Value)
	}
}

func TestResolver_MergeTOTP_NilOrEmpty(t *testing.T) {
	r := NewResolver(nil, newVault(t))
	inj, err := r.MergeTOTP(nil)
	if err != nil || inj != (HeaderInjection{}) {
		t.Fatalf("nil: %+v err=%v", inj, err)
	}
	inj, err = r.MergeTOTP(&domain.ProjectAccount{})
	if err != nil || inj != (HeaderInjection{}) {
		t.Fatalf("empty: %+v err=%v", inj, err)
	}
}

func TestResolver_MergeTOTP_HeaderDefault(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{TOTPSecretEnc: encrypt(t, v, "JBSWY3DPEHPK3PXP")}
	inj, err := r.MergeTOTP(acc)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if inj.Header != "X-OTP" || len(inj.Value) != 6 {
		t.Fatalf("got %+v", inj)
	}
}

func TestResolver_MergeTOTP_QueryParam(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		TOTPSecretEnc: encrypt(t, v, "JBSWY3DPEHPK3PXP"),
		TOTPParam:     "?otp",
	}
	inj, err := r.MergeTOTP(acc)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if inj.QueryKey != "otp" || len(inj.QueryValue) != 6 {
		t.Fatalf("got %+v", inj)
	}
	if inj.Header != "" {
		t.Fatalf("unexpected header: %q", inj.Header)
	}
}

func TestResolver_MergeTOTP_CustomHeader(t *testing.T) {
	v := newVault(t)
	r := NewResolver(nil, v)
	acc := &domain.ProjectAccount{
		TOTPSecretEnc: encrypt(t, v, "JBSWY3DPEHPK3PXP"),
		TOTPParam:     "X-MFA",
	}
	inj, err := r.MergeTOTP(acc)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if inj.Header != "X-MFA" || len(inj.Value) != 6 {
		t.Fatalf("got %+v", inj)
	}
}
