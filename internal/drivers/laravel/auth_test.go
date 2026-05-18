package laravel

import (
	"net/http"
	"strings"
	"testing"

	"spectra-desktop/internal/core"
)

func TestAuthCapability_DefaultScheme(t *testing.T) {
	if (AuthCapability{}).DefaultScheme() != core.AuthSchemeBearer {
		t.Fatal("want bearer default")
	}
}

func TestDetectAuthRole_CSRF(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodGet, Path: "/sanctum/csrf-cookie"}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleCSRF || got.Confidence != core.AuthConfidenceHigh {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_LogoutHighWithAuthMiddleware(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodPost, Path: "/api/logout", Middleware: []string{"auth:sanctum"}}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleLogout || got.Confidence != core.AuthConfidenceHigh {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_LogoutMediumWithoutAuthMiddleware(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodPost, Path: "/logout"}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleLogout || got.Confidence != core.AuthConfidenceMedium {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_Refresh(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodPost, Path: "/auth/refresh"}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleRefresh {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_LoginHighWithGuest(t *testing.T) {
	ep := core.Endpoint{
		Method:     core.MethodPost,
		Path:       "/api/login",
		Handler:    "App\\Http\\Controllers\\AuthController@login",
		Middleware: []string{"guest"},
	}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleLogin || got.Confidence != core.AuthConfidenceHigh {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_LoginNonPostIsNone(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodGet, Path: "/login"}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleNone {
		t.Fatalf("got %+v", got)
	}
}

func TestDetectAuthRole_NoneForUnrelated(t *testing.T) {
	ep := core.Endpoint{Method: core.MethodGet, Path: "/api/users"}
	got := (AuthCapability{}).DetectAuthRole(ep)
	if got.Role != core.AuthRoleNone {
		t.Fatalf("got %+v", got)
	}
}

func TestApplyAuth_BearerToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://x", nil)
	(AuthCapability{}).ApplyAuth(req, core.AuthContext{Scheme: core.AuthSchemeBearer, Token: "abc"})
	if req.Header.Get("Authorization") != "Bearer abc" {
		t.Fatalf("got %s", req.Header.Get("Authorization"))
	}
}

func TestApplyAuth_HeadersAndCookies(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://x", nil)
	(AuthCapability{}).ApplyAuth(req, core.AuthContext{
		Headers: map[string]string{"X-API-Key": "k1"},
		Cookies: []http.Cookie{{Name: "session", Value: "v"}},
	})
	if req.Header.Get("X-API-Key") != "k1" {
		t.Fatal("want X-API-Key set")
	}
	c, err := req.Cookie("session")
	if err != nil || c.Value != "v" {
		t.Fatalf("cookie not set: %v", err)
	}
}

func TestApplyAuth_DoesNotOverrideExistingHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://x", nil)
	req.Header.Set("X-API-Key", "existing")
	(AuthCapability{}).ApplyAuth(req, core.AuthContext{Headers: map[string]string{"X-API-Key": "new"}})
	if req.Header.Get("X-API-Key") != "existing" {
		t.Fatalf("got %s", req.Header.Get("X-API-Key"))
	}
}

func TestExtractCredentials_TokenAndUser(t *testing.T) {
	body := []byte(`{"data":{"token":"tk","user":{"id":"1","email":"e@x","name":"E"}}}`)
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Status: 200, Body: body})
	if !ok || got == nil {
		t.Fatal("want extraction")
	}
	if got.Token != "tk" || got.TokenPath != "data.token" {
		t.Fatalf("token: %+v", got)
	}
	if got.User == nil || got.User.Email != "e@x" || got.UserPath != "data.user" {
		t.Fatalf("user: %+v", got.User)
	}
}

func TestExtractCredentials_TopLevelToken(t *testing.T) {
	body := []byte(`{"access_token":"tk2"}`)
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Body: body})
	if !ok || got.Token != "tk2" || got.TokenPath != "access_token" {
		t.Fatalf("got %+v", got)
	}
}

func TestExtractCredentials_CookiesOnly(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "laravel_session=abc; Path=/")
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Headers: h, Body: nil})
	if ok {
		t.Fatal("empty body should return ok=false even with cookies")
	}
	if got == nil || len(got.Cookies) != 1 || got.Cookies[0].Name != "laravel_session" {
		t.Fatalf("want cookies, got %+v", got)
	}
}

func TestExtractCredentials_InvalidJSON(t *testing.T) {
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Body: []byte("not json")})
	if ok || got != nil {
		t.Fatalf("want nil/false, got %+v ok=%v", got, ok)
	}
}

func TestExtractCredentials_EmptyResponse(t *testing.T) {
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{})
	if ok || got != nil {
		t.Fatalf("want nil/false, got %+v ok=%v", got, ok)
	}
}

func TestExtractCredentials_NoMatchReturnsFalse(t *testing.T) {
	body := []byte(`{"something":{"else":1}}`)
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Body: body})
	if ok || got != nil {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}

func TestBuildUser_RequiresAtLeastOneField(t *testing.T) {
	if buildUser(map[string]any{}) != nil {
		t.Fatal("empty user must be nil")
	}
	u := buildUser(map[string]any{"role": "admin"})
	if u != nil {
		t.Fatalf("role-only must not produce a user (no id/name/username/email): %+v", u)
	}
	u = buildUser(map[string]any{"id": "1"})
	if u == nil || u.ID != "1" {
		t.Fatalf("want id=1 user, got %+v", u)
	}
}

func TestLookupString_TrimsAndType(t *testing.T) {
	m := map[string]any{"a": map[string]any{"b": " v "}}
	v, ok := lookupString(m, []string{"a", "b"})
	if !ok || v != "v" {
		t.Fatalf("want v, got %q ok=%v", v, ok)
	}
	if _, ok := lookupString(m, []string{"a", "missing"}); ok {
		t.Fatal("want missing path => false")
	}
	if _, ok := lookupString(map[string]any{"x": 5}, []string{"x"}); ok {
		t.Fatal("want non-string => false")
	}
}

func TestFirstString_OrderedFallback(t *testing.T) {
	m := map[string]any{"full_name": "Jane"}
	if firstString(m, "name", "full_name") != "Jane" {
		t.Fatal("want Jane")
	}
	if firstString(m, "x") != "" {
		t.Fatal("want empty")
	}
}

func TestLowerSlice(t *testing.T) {
	got := lowerSlice([]string{"AUTH", "Throttle"})
	if !equalStringSlice(got, []string{"auth", "throttle"}) {
		t.Fatalf("got %v", got)
	}
}

func TestParseSetCookies_Nil(t *testing.T) {
	if parseSetCookies(nil) != nil {
		t.Fatal("want nil")
	}
}

func TestExtractCredentials_RawUserContainsEmail(t *testing.T) {
	body := []byte(`{"user":{"email":"a@b"}}`)
	got, ok := (AuthCapability{}).ExtractCredentials(core.AuthResponse{Body: body})
	if !ok || got.User == nil {
		t.Fatal("want user")
	}
	if !strings.Contains(got.User.Raw, "a@b") {
		t.Fatalf("raw missing email: %s", got.User.Raw)
	}
}
