package model

import (
	"testing"
	"time"

	"spectra-desktop/internal/domain"
)

func TestProject_FromDomainToDomain_RoundTrip(t *testing.T) {
	now := time.Now().UTC()
	synced := now.Add(time.Minute)
	in := domain.Project{
		ID:                  "p1",
		Name:                "Demo",
		Path:                "/x",
		Framework:           "laravel",
		FrameworkVersion:    "11.0",
		Status:              domain.ProjectStatusConnected,
		APIFilterMode:       "auto",
		APIFilterValue:      "api",
		BaseURL:             "http://x.test",
		LoginEndpointID:     "ep-login",
		LoginTokenPath:      "data.token",
		LogoutEndpointID:    "ep-logout",
		ActiveEnvironmentID: "env",
		CreatedAt:           now,
		UpdatedAt:           now,
		LastSyncedAt:        &synced,
	}
	out := FromDomain(in).ToDomain()
	if out != in {
		t.Fatalf("roundtrip mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestProject_FromDomain_NilSynced(t *testing.T) {
	in := domain.Project{ID: "x", Path: "/p", Framework: "laravel", Status: domain.ProjectStatusDisconnected}
	m := FromDomain(in)
	if m.LastSyncedAt != nil {
		t.Fatalf("expected nil LastSyncedAt")
	}
	if m.Status != "disconnected" {
		t.Fatalf("status: %q", m.Status)
	}
}

func TestProject_ToDomain_StatusCast(t *testing.T) {
	p := Project{Status: "syncing"}
	if p.ToDomain().Status != domain.ProjectStatusSyncing {
		t.Fatalf("status cast failed")
	}
}
