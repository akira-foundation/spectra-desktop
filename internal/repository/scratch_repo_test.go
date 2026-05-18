package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"spectra-desktop/internal/repository/model"
)

func newScratch(projectID, name string) *model.ScratchRequest {
	return &model.ScratchRequest{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Name:        name,
		Method:      "GET",
		URL:         "https://example.test",
		HeadersJSON: "{}",
		Body:        "",
	}
}

func TestScratchRepository_SaveAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewScratchRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "scr")
	r1 := newScratch(p.ID, "a")
	r1.SortOrder = 1
	r2 := newScratch(p.ID, "b")
	r2.SortOrder = 0

	if err := repo.Save(ctx, r1); err != nil {
		t.Fatalf("save 1: %v", err)
	}
	if err := repo.Save(ctx, r2); err != nil {
		t.Fatalf("save 2: %v", err)
	}

	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].Name != "b" {
		t.Fatalf("unexpected order: %+v", list)
	}
}

func TestScratchRepository_Save_Upsert(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewScratchRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "scr")
	r := newScratch(p.ID, "original")
	if err := repo.Save(ctx, r); err != nil {
		t.Fatalf("save: %v", err)
	}
	r.Name = "renamed"
	r.Method = "POST"
	if err := repo.Save(ctx, r); err != nil {
		t.Fatalf("save again: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 1 || list[0].Name != "renamed" || list[0].Method != "POST" {
		t.Fatalf("upsert failed: %+v", list)
	}
}

func TestScratchRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewScratchRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "scr")
	r := newScratch(p.ID, "x")
	if err := repo.Save(ctx, r); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Delete(ctx, r.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 0 {
		t.Fatalf("expected wiped, got %d", len(list))
	}
}

func TestScratchRepository_DeleteByProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewScratchRepository(s.DB)
	ctx := context.Background()

	p1 := seedProject(t, projects, "p1")
	p2 := seedProject(t, projects, "p2")
	if err := repo.Save(ctx, newScratch(p1.ID, "a")); err != nil {
		t.Fatalf("save p1: %v", err)
	}
	if err := repo.Save(ctx, newScratch(p2.ID, "b")); err != nil {
		t.Fatalf("save p2: %v", err)
	}
	if err := repo.DeleteByProject(ctx, p1.ID); err != nil {
		t.Fatalf("delete by project: %v", err)
	}
	left1, _ := repo.List(ctx, p1.ID)
	left2, _ := repo.List(ctx, p2.ID)
	if len(left1) != 0 || len(left2) != 1 {
		t.Fatalf("scoping wrong: p1=%d p2=%d", len(left1), len(left2))
	}
}
