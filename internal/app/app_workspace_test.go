package app

import (
	"testing"

	"spectra-desktop/internal/core"
)

func TestPickAuthEndpoint_BestLoginByPathScore(t *testing.T) {
	eps := []core.Endpoint{
		{ID: "a", Path: "/auth/signin", AuthRole: core.AuthRoleLogin},
		{ID: "b", Path: "/api/login", AuthRole: core.AuthRoleLogin},
		{ID: "c", Path: "/whatever", AuthRole: core.AuthRoleLogin},
	}
	if got := pickAuthEndpoint(eps, core.AuthRoleLogin); got != "b" {
		t.Fatalf("got %q", got)
	}
}

func TestPickAuthEndpoint_BestLogoutByPathScore(t *testing.T) {
	eps := []core.Endpoint{
		{ID: "x", Path: "/api/signout", AuthRole: core.AuthRoleLogout},
		{ID: "y", Path: "/api/logout", AuthRole: core.AuthRoleLogout},
	}
	if got := pickAuthEndpoint(eps, core.AuthRoleLogout); got != "y" {
		t.Fatalf("got %q", got)
	}
}

func TestPickAuthEndpoint_IgnoresWrongRole(t *testing.T) {
	eps := []core.Endpoint{
		{ID: "a", Path: "/api/login", AuthRole: core.AuthRoleLogout},
	}
	if got := pickAuthEndpoint(eps, core.AuthRoleLogin); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestPickAuthEndpoint_NoCandidatesReturnsEmpty(t *testing.T) {
	if got := pickAuthEndpoint(nil, core.AuthRoleLogin); got != "" {
		t.Fatalf("got %q", got)
	}
}
