package repository

import (
	"context"
	"testing"
	"time"

	"spectra-desktop/internal/core"
)

func makeEndpoints() []core.Endpoint {
	return []core.Endpoint{
		{
			Method:     core.MethodGet,
			Path:       "/users",
			Handler:    "UserController@index",
			Middleware: []string{"auth", "api"},
			Tags:       []string{"users"},
			Source:     core.EndpointSource{File: "routes/api.php", Line: 10},
			Framework:  "laravel",
			Confidence: 0.9,
		},
		{
			Method:  core.MethodPost,
			Path:    "/users",
			Handler: "UserController@store",
			Source:  core.EndpointSource{File: "routes/api.php", Line: 12},
		},
	}
}

func TestEndpointRepository_ReplaceAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")

	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("replace: %v", err)
	}

	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 endpoints, got %d", len(list))
	}
	if list[0].Path != "/users" || list[0].Method != core.MethodGet {
		t.Fatalf("unexpected ordering: %+v", list[0])
	}
	if len(list[0].Middleware) != 2 {
		t.Fatalf("middleware decode: %+v", list[0].Middleware)
	}
}

func TestEndpointRepository_GetByID_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewEndpointRepository(s.DB)

	got, err := repo.GetByID(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil endpoint, got %+v", got)
	}
}

func TestEndpointRepository_GetByID_Found(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("replace: %v", err)
	}

	list, _ := repo.List(ctx, p.ID)
	got, err := repo.GetByID(ctx, list[0].ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.ID != list[0].ID {
		t.Fatalf("expected found endpoint, got %+v", got)
	}
}

func TestEndpointRepository_ProjectIDOf(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)

	got, err := repo.ProjectIDOf(ctx, list[0].ID)
	if err != nil {
		t.Fatalf("project id: %v", err)
	}
	if got != p.ID {
		t.Fatalf("expected %q, got %q", p.ID, got)
	}

	missing, err := repo.ProjectIDOf(ctx, "nope")
	if err != nil {
		t.Fatalf("missing project id: %v", err)
	}
	if missing != "" {
		t.Fatalf("expected empty project id, got %q", missing)
	}
}

func TestEndpointRepository_Replace_PreservesAuthOverrides(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)

	if err := repo.UpdateAuthOverride(ctx, list[0].ID, core.AuthRoleLogin, "data.token"); err != nil {
		t.Fatalf("override: %v", err)
	}

	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("replace again: %v", err)
	}

	after, _ := repo.List(ctx, p.ID)
	var matched core.Endpoint
	for _, ep := range after {
		if ep.Method == core.MethodGet && ep.Path == "/users" {
			matched = ep
			break
		}
	}
	if matched.AuthRoleOverride != core.AuthRoleLogin {
		t.Fatalf("expected override preserved, got %+v", matched)
	}
	if matched.TokenPathOverride != "data.token" {
		t.Fatalf("expected token path preserved, got %q", matched.TokenPathOverride)
	}
}

func TestEndpointRepository_Replace_EmptyClearsRows(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.Replace(ctx, p.ID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}

func TestEndpointRepository_DeleteByProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.DeleteByProject(ctx, p.ID); err != nil {
		t.Fatalf("delete by project: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 0 {
		t.Fatalf("expected wiped, got %d", len(list))
	}
}

func TestEndpointRepository_ProjectDeleteCascades(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ep")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := projects.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 0 {
		t.Fatalf("foreign key cascade did not fire; rows=%d", len(list))
	}
}

func TestEndpointRepository_Stats(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "stats")
	if err := repo.Replace(ctx, p.ID, makeEndpoints()); err != nil {
		t.Fatalf("seed: %v", err)
	}

	stats, err := repo.Stats(ctx, p.ID)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Routes != 2 {
		t.Fatalf("routes = %d", stats.Routes)
	}
	if stats.Controllers != 1 {
		t.Fatalf("controllers = %d", stats.Controllers)
	}
	if stats.Middleware != 2 {
		t.Fatalf("middleware = %d", stats.Middleware)
	}
	if stats.LastScannedAt == nil {
		t.Fatal("expected LastScannedAt")
	}
	if time.Since(*stats.LastScannedAt) > time.Minute {
		t.Fatalf("scanned_at too old: %v", stats.LastScannedAt)
	}
}

func TestEndpointRepository_Stats_Empty(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "stats-empty")

	stats, err := repo.Stats(ctx, p.ID)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Routes != 0 || stats.Controllers != 0 || stats.Middleware != 0 {
		t.Fatalf("expected zero stats, got %+v", stats)
	}
	if stats.LastScannedAt != nil {
		t.Fatalf("expected nil LastScannedAt, got %v", stats.LastScannedAt)
	}
}

func TestNormalizeController(t *testing.T) {
	if normalizeController("UserController@index") != "UserController" {
		t.Fatal("strip @")
	}
	if normalizeController("UserController") != "UserController" {
		t.Fatal("passthrough")
	}
	if normalizeController("") != "" {
		t.Fatal("empty")
	}
}
