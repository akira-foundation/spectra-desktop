package repository

import (
	"context"
	"testing"

	"spectra-desktop/internal/domain"
)

func sampleTests() []domain.EndpointTest {
	return []domain.EndpointTest{
		{Name: "status", Kind: "status", Op: "eq", Expected: "200"},
		{Name: "id", Kind: "json", JSONPath: "data.id", Op: "exists", Expected: ""},
	}
}

func TestTestRepository_ReplaceAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewTestRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "tst")
	if err := repo.Replace(ctx, p.ID, "GET /users", sampleTests()); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, err := repo.List(ctx, p.ID, "GET /users")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].Name != "status" {
		t.Fatalf("unexpected: %+v", list)
	}
	if list[0].SortOrder != 0 || list[1].SortOrder != 1 {
		t.Fatalf("sort order: %+v", list)
	}
	if list[0].ID == "" {
		t.Fatalf("expected generated ID")
	}
}

func TestTestRepository_Replace_EmptyClears(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewTestRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "tst")
	if err := repo.Replace(ctx, p.ID, "k", sampleTests()); err != nil {
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

func TestTestRepository_ScopedByEndpointKey(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewTestRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "tst")
	if err := repo.Replace(ctx, p.ID, "a", sampleTests()); err != nil {
		t.Fatalf("a: %v", err)
	}
	if err := repo.Replace(ctx, p.ID, "b", sampleTests()[:1]); err != nil {
		t.Fatalf("b: %v", err)
	}
	a, _ := repo.List(ctx, p.ID, "a")
	b, _ := repo.List(ctx, p.ID, "b")
	if len(a) != 2 || len(b) != 1 {
		t.Fatalf("scoping wrong a=%d b=%d", len(a), len(b))
	}
}

func TestTestRepository_DeleteByEndpoint(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewTestRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "tst")
	if err := repo.Replace(ctx, p.ID, "k", sampleTests()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.DeleteByEndpoint(ctx, p.ID, "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, "k")
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

func TestTestRepository_PreservesProvidedID(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewTestRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "tst")
	tests := []domain.EndpointTest{{ID: "fixed", Name: "n", Kind: "status", Op: "eq", Expected: "200"}}
	if err := repo.Replace(ctx, p.ID, "k", tests); err != nil {
		t.Fatalf("replace: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, "k")
	if len(list) != 1 || list[0].ID != "fixed" {
		t.Fatalf("expected fixed id, got %+v", list)
	}
}
