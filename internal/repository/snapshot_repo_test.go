package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func makeSnapshot(projectID string, at time.Time) domain.EndpointSnapshot {
	return domain.EndpointSnapshot{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
		Hash:          uuid.NewString(),
		PayloadJSON:   `{"a":1}`,
		EndpointCount: 1,
		ScannedAt:     at,
	}
}

func TestSnapshotRepository_SaveAndListAndGet(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	now := time.Now().UTC()
	a := makeSnapshot(p.ID, now.Add(-2*time.Hour))
	b := makeSnapshot(p.ID, now)
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("save b: %v", err)
	}
	list, err := repo.List(ctx, p.ID, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].ID != b.ID {
		t.Fatalf("expected newest first, got %+v", list)
	}
	if list[0].PayloadJSON != "" {
		t.Fatalf("list should not include payload, got %q", list[0].PayloadJSON)
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got == nil || got.ID != a.ID || got.PayloadJSON != `{"a":1}` {
		t.Fatalf("GetByID should include payload: %+v", got)
	}
}

func TestSnapshotRepository_GetByID_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewSnapshotRepository(s.DB)
	got, err := repo.GetByID(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestSnapshotRepository_Latest(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	now := time.Now().UTC()
	older := makeSnapshot(p.ID, now.Add(-time.Hour))
	newer := makeSnapshot(p.ID, now)
	if err := repo.Save(ctx, older); err != nil {
		t.Fatalf("older: %v", err)
	}
	if err := repo.Save(ctx, newer); err != nil {
		t.Fatalf("newer: %v", err)
	}
	got, err := repo.Latest(ctx, p.ID)
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got == nil || got.ID != newer.ID {
		t.Fatalf("expected newer, got %+v", got)
	}
}

func TestSnapshotRepository_Latest_EmptyReturnsNil(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()
	p := seedProject(t, projects, "snap")
	got, err := repo.Latest(ctx, p.ID)
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestSnapshotRepository_Predecessor(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	now := time.Now().UTC()
	a := makeSnapshot(p.ID, now.Add(-2*time.Hour))
	b := makeSnapshot(p.ID, now.Add(-time.Hour))
	c := makeSnapshot(p.ID, now)
	for _, sn := range []domain.EndpointSnapshot{a, b, c} {
		if err := repo.Save(ctx, sn); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	pred, err := repo.Predecessor(ctx, p.ID, c.ScannedAt)
	if err != nil {
		t.Fatalf("predecessor: %v", err)
	}
	if pred == nil || pred.ID != b.ID {
		t.Fatalf("expected b, got %+v", pred)
	}

	first, err := repo.Predecessor(ctx, p.ID, a.ScannedAt)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if first != nil {
		t.Fatalf("expected nil for oldest, got %+v", first)
	}
}

func TestSnapshotRepository_TrimOldest(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		sn := makeSnapshot(p.ID, now.Add(time.Duration(i)*time.Minute))
		if err := repo.Save(ctx, sn); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	if err := repo.TrimOldest(ctx, p.ID, 2); err != nil {
		t.Fatalf("trim: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if len(list) != 2 {
		t.Fatalf("expected 2 after trim, got %d", len(list))
	}
}

func TestSnapshotRepository_TrimOldest_KeepZeroNoOp(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	if err := repo.Save(ctx, makeSnapshot(p.ID, time.Now().UTC())); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.TrimOldest(ctx, p.ID, 0); err != nil {
		t.Fatalf("trim: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if len(list) != 1 {
		t.Fatalf("expected no-op, got %d", len(list))
	}
}

func TestSnapshotRepository_List_DefaultLimit(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewSnapshotRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "snap")
	if err := repo.Save(ctx, makeSnapshot(p.ID, time.Now().UTC())); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := repo.List(ctx, p.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}
}
