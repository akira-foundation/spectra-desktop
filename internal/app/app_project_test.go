package app

import (
	"testing"

	"spectra-desktop/internal/core"
)

func TestScoreAuthPath_LoginVariants(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"/api/login", 5},
		{"/api/login/extra", 3},
		{"/auth/signin", 2},
		{"/auth/sign-in", 2},
		{"/authenticate", 2},
		{"/auth/token", 2},
		{"/unrelated", 0},
	}
	for _, c := range cases {
		if got := scoreAuthPath(c.path, core.AuthRoleLogin); got != c.want {
			t.Fatalf("%q: got %d want %d", c.path, got, c.want)
		}
	}
}

func TestScoreAuthPath_LogoutVariants(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"/api/logout", 5},
		{"/api/logout/all", 3},
		{"/api/signout", 2},
		{"/api/sign-out", 2},
		{"/api/login", 0},
	}
	for _, c := range cases {
		if got := scoreAuthPath(c.path, core.AuthRoleLogout); got != c.want {
			t.Fatalf("%q: got %d want %d", c.path, got, c.want)
		}
	}
}

func TestScoreAuthPath_UnknownRoleReturnsZero(t *testing.T) {
	if got := scoreAuthPath("/api/login", core.AuthRole("other")); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestScoreAuthPath_CaseInsensitive(t *testing.T) {
	if got := scoreAuthPath("/API/LOGIN", core.AuthRoleLogin); got != 5 {
		t.Fatalf("got %d", got)
	}
}

func TestSummarizeAuth_PicksBestLoginAndLogout(t *testing.T) {
	eps := []core.Endpoint{
		{Method: core.MethodGet, Path: "/auth/signin", AuthRole: core.AuthRoleLogin},
		{Method: core.MethodPost, Path: "/api/login", AuthRole: core.AuthRoleLogin},
		{Method: core.MethodPost, Path: "/api/logout", AuthRole: core.AuthRoleLogout},
		{Method: core.MethodGet, Path: "/other", AuthRole: core.AuthRoleNone},
	}
	login, logout := summarizeAuth(eps)
	if login != "POST /api/login" {
		t.Fatalf("login: %q", login)
	}
	if logout != "POST /api/logout" {
		t.Fatalf("logout: %q", logout)
	}
}

func TestSummarizeAuth_NoCandidatesReturnsEmpty(t *testing.T) {
	login, logout := summarizeAuth(nil)
	if login != "" || logout != "" {
		t.Fatalf("got %q, %q", login, logout)
	}
}
