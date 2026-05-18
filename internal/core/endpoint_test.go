package core

import "testing"

func TestEndpoint_EffectiveAuthRole_NoOverride(t *testing.T) {
	e := Endpoint{AuthRole: AuthRoleLogin}
	if got := e.EffectiveAuthRole(); got != AuthRoleLogin {
		t.Fatalf("want %q, got %q", AuthRoleLogin, got)
	}
}

func TestEndpoint_EffectiveAuthRole_OverrideWins(t *testing.T) {
	e := Endpoint{AuthRole: AuthRoleLogin, AuthRoleOverride: AuthRoleRefresh}
	if got := e.EffectiveAuthRole(); got != AuthRoleRefresh {
		t.Fatalf("want %q, got %q", AuthRoleRefresh, got)
	}
}

func TestEndpoint_EffectiveAuthRole_OverrideNoneClears(t *testing.T) {
	e := Endpoint{AuthRole: AuthRoleLogin, AuthRoleOverride: AuthRole("none")}
	if got := e.EffectiveAuthRole(); got != AuthRoleNone {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestEndpoint_EffectiveAuthRole_AllZero(t *testing.T) {
	e := Endpoint{}
	if got := e.EffectiveAuthRole(); got != AuthRoleNone {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestHTTPMethod_Constants(t *testing.T) {
	pairs := map[HTTPMethod]string{
		MethodGet:     "GET",
		MethodPost:    "POST",
		MethodPut:     "PUT",
		MethodPatch:   "PATCH",
		MethodDelete:  "DELETE",
		MethodHead:    "HEAD",
		MethodOptions: "OPTIONS",
	}
	for m, want := range pairs {
		if string(m) != want {
			t.Fatalf("method %q != %q", string(m), want)
		}
	}
}
