package repository

import (
	"context"
	"testing"
)

func TestDatasetRepository_Get_MissingReturnsEmptyArray(t *testing.T) {
	s := newStorage(t)
	repo := NewDatasetRepository(s.DB)
	got, err := repo.Get(context.Background(), "p", "k")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "[]" {
		t.Fatalf("expected []: got %q", got)
	}
}

func TestDatasetRepository_SaveAndGet(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewDatasetRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ds")
	if err := repo.Save(ctx, p.ID, "k", `[{"id":1}]`); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := repo.Get(ctx, p.ID, "k")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != `[{"id":1}]` {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestDatasetRepository_Save_Upsert(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewDatasetRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ds")
	if err := repo.Save(ctx, p.ID, "k", "[1]"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Save(ctx, p.ID, "k", "[2]"); err != nil {
		t.Fatalf("save again: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID, "k")
	if got != "[2]" {
		t.Fatalf("expected upsert, got %q", got)
	}
}

func TestDatasetRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewDatasetRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ds")
	if err := repo.Save(ctx, p.ID, "k", "[1]"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Delete(ctx, p.ID, "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID, "k")
	if got != "[]" {
		t.Fatalf("expected reset, got %q", got)
	}
}

func TestDatasetRepository_GetEmptyStringReturnsArray(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewDatasetRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "ds")
	if err := repo.Save(ctx, p.ID, "k", ""); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID, "k")
	if got != "[]" {
		t.Fatalf("expected fallback to []: got %q", got)
	}
}
