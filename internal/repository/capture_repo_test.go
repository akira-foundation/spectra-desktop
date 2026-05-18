package repository

import (
	"context"
	"testing"

	"spectra-desktop/internal/domain"
)

func sampleCaptures() []domain.EndpointCapture {
	return []domain.EndpointCapture{
		{Name: "token", Source: "body", Path: "data.token"},
		{Name: "id", Source: "body", Path: "data.id"},
	}
}

func TestCaptureRepository_ReplaceAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCaptureRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cap")
	if err := repo.Replace(ctx, p.ID, "GET /users", sampleCaptures()); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, err := repo.List(ctx, p.ID, "GET /users")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	if list[0].Name != "token" || list[0].SortOrder != 0 {
		t.Fatalf("unexpected order: %+v", list[0])
	}
	if list[1].SortOrder != 1 {
		t.Fatalf("expected SortOrder 1, got %d", list[1].SortOrder)
	}
	if list[0].ID == "" {
		t.Fatalf("expected generated ID")
	}
}

func TestCaptureRepository_Replace_EmptyClears(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCaptureRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cap")
	if err := repo.Replace(ctx, p.ID, "k", sampleCaptures()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.Replace(ctx, p.ID, "k", nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, "k")
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

func TestCaptureRepository_ScopedByEndpointKey(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCaptureRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cap")
	if err := repo.Replace(ctx, p.ID, "a", sampleCaptures()); err != nil {
		t.Fatalf("a: %v", err)
	}
	if err := repo.Replace(ctx, p.ID, "b", sampleCaptures()[:1]); err != nil {
		t.Fatalf("b: %v", err)
	}
	a, _ := repo.List(ctx, p.ID, "a")
	b, _ := repo.List(ctx, p.ID, "b")
	if len(a) != 2 || len(b) != 1 {
		t.Fatalf("unexpected counts a=%d b=%d", len(a), len(b))
	}
}

func TestCaptureRepository_DeleteByEndpoint(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCaptureRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cap")
	if err := repo.Replace(ctx, p.ID, "a", sampleCaptures()); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if err := repo.DeleteByEndpoint(ctx, p.ID, "a"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, "a")
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

func TestCaptureRepository_PreservesProvidedID(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCaptureRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cap")
	captures := []domain.EndpointCapture{{ID: "fixed", Name: "n", Source: "body", Path: "x"}}
	if err := repo.Replace(ctx, p.ID, "k", captures); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, "k")
	if len(list) != 1 || list[0].ID != "fixed" {
		t.Fatalf("expected fixed id, got %+v", list)
	}
}
