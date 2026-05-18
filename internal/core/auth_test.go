package core

import "testing"

func TestAuthRole_Constants(t *testing.T) {
	if string(AuthRoleNone) != "" {
		t.Fatalf("AuthRoleNone should be empty")
	}
	if string(AuthRoleLogin) != "login" {
		t.Fatalf("AuthRoleLogin: %q", AuthRoleLogin)
	}
	if string(AuthRoleLogout) != "logout" {
		t.Fatalf("AuthRoleLogout: %q", AuthRoleLogout)
	}
	if string(AuthRoleRefresh) != "refresh" {
		t.Fatalf("AuthRoleRefresh: %q", AuthRoleRefresh)
	}
	if string(AuthRoleCSRF) != "csrf" {
		t.Fatalf("AuthRoleCSRF: %q", AuthRoleCSRF)
	}
}

func TestAuthScheme_Constants(t *testing.T) {
	want := map[AuthScheme]string{
		AuthSchemeNone:   "",
		AuthSchemeBearer: "bearer",
		AuthSchemeCookie: "cookie",
		AuthSchemeBasic:  "basic",
		AuthSchemeAPIKey: "api_key",
		AuthSchemeCustom: "custom",
	}
	for s, exp := range want {
		if string(s) != exp {
			t.Fatalf("scheme %q != %q", string(s), exp)
		}
	}
}

func TestAuthConfidence_Constants(t *testing.T) {
	if string(AuthConfidenceHigh) != "high" || string(AuthConfidenceMedium) != "medium" || string(AuthConfidenceLow) != "low" {
		t.Fatalf("confidence constants drifted")
	}
}

func TestAuthRoleHint_ZeroValue(t *testing.T) {
	var h AuthRoleHint
	if h.Role != AuthRoleNone || h.Confidence != "" || h.Reason != "" {
		t.Fatalf("zero hint not empty: %+v", h)
	}
}

func TestAuthContext_NilOptionals(t *testing.T) {
	c := AuthContext{Scheme: AuthSchemeBearer, Token: "tok"}
	if c.User != nil {
		t.Fatalf("User should be nil")
	}
	if c.ExpiresAt != nil {
		t.Fatalf("ExpiresAt should be nil")
	}
}
