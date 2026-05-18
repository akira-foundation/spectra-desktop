package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func TestProjectRepository_SaveAndGetByID(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	in := domain.ProjectInput{
		ID:        uuid.NewString(),
		Name:      "  Acme  ",
		Path:      "/srv/acme",
		Framework: "laravel",
	}
	saved, err := repo.Save(ctx, in)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.Name != "Acme" {
		t.Fatalf("expected trimmed name, got %q", saved.Name)
	}
	if saved.Status != domain.ProjectStatusDisconnected {
		t.Fatalf("default status = %q", saved.Status)
	}
	if saved.APIFilterMode != domain.APIFilterModeAuto {
		t.Fatalf("default api filter = %q", saved.APIFilterMode)
	}

	got, err := repo.GetByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != saved.ID || got.Path != "/srv/acme" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestProjectRepository_GetByID_MissingReturnsErrNotFound(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)

	_, err := repo.GetByID(context.Background(), "does-not-exist")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_GetByPath_MissingReturnsErrNotFound(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)

	_, err := repo.GetByPath(context.Background(), "/nope")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_Save_UpdatesExistingByPath(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	first, err := repo.Save(ctx, domain.ProjectInput{Name: "v1", Path: "/srv/x", Framework: "laravel"})
	if err != nil {
		t.Fatalf("save first: %v", err)
	}

	second, err := repo.Save(ctx, domain.ProjectInput{
		Name:      " v2 ",
		Path:      "/srv/x",
		Framework: "rails",
		BaseURL:   "http://localhost",
	})
	if err != nil {
		t.Fatalf("save second: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same id, got %s vs %s", second.ID, first.ID)
	}
	if second.Name != "v2" || second.Framework != "rails" || second.BaseURL != "http://localhost" {
		t.Fatalf("expected updated fields: %+v", second)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project after update, got %d", len(list))
	}
}

func TestProjectRepository_List_OrdersByCreatedAt(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p1 := seedProject(t, repo, "first")
	p2 := seedProject(t, repo, "second")

	rows, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].ID != p1.ID || rows[1].ID != p2.ID {
		t.Fatalf("unexpected order: %s, %s", rows[0].ID, rows[1].ID)
	}
}

func TestProjectRepository_Delete(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, repo, "kill-me")

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.Delete(ctx, p.ID); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound on second delete, got %v", err)
	}
}

func TestProjectRepository_UpdateStatus(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, repo, "status")

	if err := repo.UpdateStatus(ctx, p.ID, domain.ProjectStatusError); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := repo.GetByID(ctx, p.ID)
	if got.Status != domain.ProjectStatusError {
		t.Fatalf("status = %q", got.Status)
	}

	if err := repo.UpdateStatus(ctx, "missing", domain.ProjectStatusError); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_MarkSynced(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, repo, "sync")

	if err := repo.MarkSynced(ctx, p.ID); err != nil {
		t.Fatalf("mark synced: %v", err)
	}
	got, _ := repo.GetByID(ctx, p.ID)
	if got.Status != domain.ProjectStatusConnected {
		t.Fatalf("status = %q", got.Status)
	}
	if got.LastSyncedAt == nil {
		t.Fatal("expected LastSyncedAt populated")
	}

	if err := repo.MarkSynced(ctx, "nope"); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_UpdateBaseURL(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, repo, "base")

	if err := repo.UpdateBaseURL(ctx, p.ID, "  https://api.local  "); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := repo.GetByID(ctx, p.ID)
	if got.BaseURL != "https://api.local" {
		t.Fatalf("base url = %q", got.BaseURL)
	}

	if err := repo.UpdateBaseURL(ctx, "missing", "x"); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_UpdateAuthRoutes(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, repo, "auth")

	if err := repo.UpdateAuthRoutes(ctx, p.ID, "login-1", "logout-1", "data.token"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := repo.GetByID(ctx, p.ID)
	if got.LoginEndpointID != "login-1" || got.LogoutEndpointID != "logout-1" || got.LoginTokenPath != "data.token" {
		t.Fatalf("auth fields: %+v", got)
	}

	if err := repo.UpdateAuthRoutes(ctx, "missing", "", "", ""); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_UpdateActiveEnvironment_NoErrorOnMissing(t *testing.T) {
	s := newStorage(t)
	repo := NewProjectRepository(s.DB)

	if err := repo.UpdateActiveEnvironment(context.Background(), "nope", "env-1"); err != nil {
		t.Fatalf("update missing should not error, got %v", err)
	}
}

func TestNormalizeAPIFilter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		mode      string
		value     string
		wantMode  string
		wantValue string
	}{
		{"unknown mode falls back to auto", "weird", "v", domain.APIFilterModeAuto, "v"},
		{"middleware passes through", " MIDDLEWARE ", " api ", domain.APIFilterModeMiddleware, "api"},
		{"all clears value", "all", "ignored", domain.APIFilterModeAll, ""},
		{"prefix kept", "prefix", "/api", domain.APIFilterModePrefix, "/api"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mode, value := normalizeAPIFilter(tc.mode, tc.value)
			if mode != tc.wantMode || value != tc.wantValue {
				t.Fatalf("mode=%q value=%q", mode, value)
			}
		})
	}
}
