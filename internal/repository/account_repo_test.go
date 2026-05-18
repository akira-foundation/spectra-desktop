package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func newAccount(projectID, label string) domain.ProjectAccount {
	return domain.ProjectAccount{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Label:     label,
		Kind:      domain.AccountKindBearer,
		TokenEnc:  "enc-token",
		SortOrder: 0,
	}
}

func TestAccountRepository_SaveAndList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "Primary")
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save: %v", err)
	}
	b := newAccount(p.ID, "Secondary")
	b.SortOrder = 1
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("save b: %v", err)
	}

	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	if list[0].Label != "Primary" {
		t.Fatalf("unexpected sort: %+v", list[0].Label)
	}
}

func TestAccountRepository_Save_Upsert(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "Primary")
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save: %v", err)
	}
	a.Label = "Renamed"
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save again: %v", err)
	}
	got, err := repo.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.Label != "Renamed" {
		t.Fatalf("expected Renamed, got %+v", got)
	}
}

func TestAccountRepository_Get_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewAccountRepository(s.DB)
	got, err := repo.Get(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestAccountRepository_GetDefault_PrefersFlag(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "A")
	b := newAccount(p.ID, "B")
	b.IsDefault = true
	b.SortOrder = 5
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("save b: %v", err)
	}

	got, err := repo.GetDefault(ctx, p.ID)
	if err != nil {
		t.Fatalf("default: %v", err)
	}
	if got == nil || got.ID != b.ID {
		t.Fatalf("expected b as default, got %+v", got)
	}
}

func TestAccountRepository_GetDefault_FallsBackToFirst(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "A")
	a.SortOrder = 2
	b := newAccount(p.ID, "B")
	b.SortOrder = 1
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("save b: %v", err)
	}
	got, err := repo.GetDefault(ctx, p.ID)
	if err != nil {
		t.Fatalf("default: %v", err)
	}
	if got == nil || got.ID != b.ID {
		t.Fatalf("expected b (lowest sort), got %+v", got)
	}
}

func TestAccountRepository_GetDefault_EmptyReturnsNil(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()
	p := seedProject(t, projects, "acc")

	got, err := repo.GetDefault(ctx, p.ID)
	if err != nil {
		t.Fatalf("default: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestAccountRepository_SetDefault_ClearsOthers(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "A")
	a.IsDefault = true
	b := newAccount(p.ID, "B")
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("save b: %v", err)
	}

	if err := repo.SetDefault(ctx, p.ID, b.ID); err != nil {
		t.Fatalf("set default: %v", err)
	}

	gotA, _ := repo.Get(ctx, a.ID)
	gotB, _ := repo.Get(ctx, b.ID)
	if gotA == nil || gotA.IsDefault {
		t.Fatalf("expected a cleared, got %+v", gotA)
	}
	if gotB == nil || !gotB.IsDefault {
		t.Fatalf("expected b default, got %+v", gotB)
	}
}

func TestAccountRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "acc")
	a := newAccount(p.ID, "A")
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Delete(ctx, a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.Get(ctx, a.ID)
	if got != nil {
		t.Fatalf("expected gone, got %+v", got)
	}
}

func TestAccountRepository_ListScopedByProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAccountRepository(s.DB)
	ctx := context.Background()

	p1 := seedProject(t, projects, "p1")
	p2 := seedProject(t, projects, "p2")
	if err := repo.Save(ctx, newAccount(p1.ID, "A")); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Save(ctx, newAccount(p2.ID, "B")); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := repo.List(ctx, p1.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Label != "A" {
		t.Fatalf("expected only project p1 rows, got %+v", list)
	}
}
