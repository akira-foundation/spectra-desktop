package repository

import (
	"context"
	"testing"
	"time"
)

func TestCapturedValuesRepository_UpsertAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCapturedValuesRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cv")
	now := time.Now().UTC()
	if err := repo.Upsert(ctx, p.ID, "token", "abc", "GET /login", now); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	rows, err := repo.ListByProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != "abc" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestCapturedValuesRepository_Upsert_Overrides(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCapturedValuesRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cv")
	now := time.Now().UTC()
	if err := repo.Upsert(ctx, p.ID, "k", "v1", "e1", now); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := repo.Upsert(ctx, p.ID, "k", "v2", "e2", now); err != nil {
		t.Fatalf("second: %v", err)
	}
	rows, _ := repo.ListByProject(ctx, p.ID)
	if len(rows) != 1 || rows[0].Value != "v2" || rows[0].EndpointKey != "e2" {
		t.Fatalf("upsert failed: %+v", rows)
	}
}

func TestCapturedValuesRepository_DeleteByProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCapturedValuesRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cv")
	now := time.Now().UTC()
	if err := repo.Upsert(ctx, p.ID, "k", "v", "e", now); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := repo.DeleteByProject(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	rows, _ := repo.ListByProject(ctx, p.ID)
	if len(rows) != 0 {
		t.Fatalf("expected wiped, got %d", len(rows))
	}
}

func TestCapturedValuesRepository_DeleteByEndpoint_KeepSet(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCapturedValuesRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "cv")
	now := time.Now().UTC()
	if err := repo.Upsert(ctx, p.ID, "keep", "v", "e1", now); err != nil {
		t.Fatalf("upsert keep: %v", err)
	}
	if err := repo.Upsert(ctx, p.ID, "drop", "v", "e1", now); err != nil {
		t.Fatalf("upsert drop: %v", err)
	}
	if err := repo.Upsert(ctx, p.ID, "other", "v", "e2", now); err != nil {
		t.Fatalf("upsert other: %v", err)
	}
	if err := repo.DeleteByEndpoint(ctx, p.ID, "e1", map[string]bool{"keep": true}); err != nil {
		t.Fatalf("delete by endpoint: %v", err)
	}
	rows, _ := repo.ListByProject(ctx, p.ID)
	got := map[string]bool{}
	for _, r := range rows {
		got[r.Name] = true
	}
	if !got["keep"] || !got["other"] || got["drop"] {
		t.Fatalf("unexpected remaining: %+v", got)
	}
}

func TestCapturedValuesRepository_ListByProject_Scoped(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCapturedValuesRepository(s.DB)
	ctx := context.Background()

	p1 := seedProject(t, projects, "p1")
	p2 := seedProject(t, projects, "p2")
	now := time.Now().UTC()
	if err := repo.Upsert(ctx, p1.ID, "k", "a", "e", now); err != nil {
		t.Fatalf("p1: %v", err)
	}
	if err := repo.Upsert(ctx, p2.ID, "k", "b", "e", now); err != nil {
		t.Fatalf("p2: %v", err)
	}
	rows, _ := repo.ListByProject(ctx, p1.ID)
	if len(rows) != 1 || rows[0].Value != "a" {
		t.Fatalf("scoping broken: %+v", rows)
	}
}
