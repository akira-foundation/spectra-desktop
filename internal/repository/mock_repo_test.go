package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func makeMock(projectID, endpointID string) domain.MockOverride {
	return domain.MockOverride{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		EndpointID:  endpointID,
		Enabled:     true,
		Status:      200,
		LatencyMs:   25,
		Body:        `{"ok":true}`,
		HeadersJSON: `{"Content-Type":"application/json"}`,
		Source:      domain.MockSourceCustom,
	}
}

func TestMockRepository_SaveAndGet(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewMockRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "m")
	override := makeMock(p.ID, "ep-1")
	if err := repo.Save(ctx, override); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.Get(ctx, p.ID, "ep-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.ID != override.ID || got.Source != domain.MockSourceCustom {
		t.Fatalf("round-trip: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("expected populated timestamps")
	}
}

func TestMockRepository_Get_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewMockRepository(s.DB)
	got, err := repo.Get(context.Background(), "p", "ep")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestMockRepository_Save_UpsertsOnConflict(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewMockRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "m")
	override := makeMock(p.ID, "ep-1")
	if err := repo.Save(ctx, override); err != nil {
		t.Fatalf("first save: %v", err)
	}

	override.Status = 418
	override.Body = `{"teapot":true}`
	if err := repo.Save(ctx, override); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, _ := repo.Get(ctx, p.ID, "ep-1")
	if got.Status != 418 || got.Body != `{"teapot":true}` {
		t.Fatalf("upsert failed: %+v", got)
	}
}

func TestMockRepository_List_OrderedByUpdatedAtDesc(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewMockRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "m")
	first := makeMock(p.ID, "ep-1")
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}
	second := makeMock(p.ID, "ep-2")
	if err := repo.Save(ctx, second); err != nil {
		t.Fatalf("save second: %v", err)
	}

	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestMockRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewMockRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "m")
	override := makeMock(p.ID, "ep-1")
	if err := repo.Save(ctx, override); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Delete(ctx, override.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID, "ep-1")
	if got != nil {
		t.Fatalf("expected deleted, got %+v", got)
	}
}

func TestMockRepository_DeleteByProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewMockRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "m")
	for i, ep := range []string{"a", "b", "c"} {
		_ = i
		if err := repo.Save(ctx, makeMock(p.ID, ep)); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if err := repo.DeleteByProject(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ := repo.List(ctx, p.ID)
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

func TestMockRepository_Delete_MissingNoError(t *testing.T) {
	s := newStorage(t)
	repo := NewMockRepository(s.DB)
	if err := repo.Delete(context.Background(), "nope"); err != nil {
		t.Fatalf("delete missing: %v", err)
	}
}
