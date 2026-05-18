package repository

import (
	"context"
	"testing"

	"spectra-desktop/internal/domain"
)

func TestEnvironmentRepository_SaveCreatesNew(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEnvironmentRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "env")
	env, err := repo.Save(ctx, domain.EnvironmentInput{
		ProjectID: p.ID,
		Name:      "local",
		Vars:      map[string]string{"BASE": "http://local"},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if env.ID == "" || env.Vars["BASE"] != "http://local" {
		t.Fatalf("unexpected: %+v", env)
	}
}

func TestEnvironmentRepository_SaveUpdates(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEnvironmentRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "env")
	env, _ := repo.Save(ctx, domain.EnvironmentInput{ProjectID: p.ID, Name: "a", Vars: map[string]string{"K": "1"}})

	updated, err := repo.Save(ctx, domain.EnvironmentInput{
		ID: env.ID, ProjectID: p.ID, Name: "renamed", Vars: map[string]string{"K": "2"}, SortOrder: 3,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "renamed" || updated.Vars["K"] != "2" || updated.SortOrder != 3 {
		t.Fatalf("update failed: %+v", updated)
	}
}

func TestEnvironmentRepository_List(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEnvironmentRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "env")
	if _, err := repo.Save(ctx, domain.EnvironmentInput{ProjectID: p.ID, Name: "b", SortOrder: 1}); err != nil {
		t.Fatalf("b: %v", err)
	}
	if _, err := repo.Save(ctx, domain.EnvironmentInput{ProjectID: p.ID, Name: "a", SortOrder: 0}); err != nil {
		t.Fatalf("a: %v", err)
	}
	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].Name != "a" {
		t.Fatalf("unexpected order: %+v", list)
	}
}

func TestEnvironmentRepository_GetByID_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewEnvironmentRepository(s.DB)
	got, err := repo.GetByID(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestEnvironmentRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEnvironmentRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "env")
	env, _ := repo.Save(ctx, domain.EnvironmentInput{ProjectID: p.ID, Name: "x"})
	if err := repo.Delete(ctx, env.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.GetByID(ctx, env.ID)
	if got != nil {
		t.Fatalf("expected gone, got %+v", got)
	}
}

func TestEnvironmentRepository_VarsRoundTripEmpty(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEnvironmentRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "env")
	env, _ := repo.Save(ctx, domain.EnvironmentInput{ProjectID: p.ID, Name: "x", Vars: map[string]string{}})
	got, _ := repo.GetByID(ctx, env.ID)
	if got == nil {
		t.Fatalf("expected environment, got nil")
	}
	if len(got.Vars) != 0 {
		t.Fatalf("expected empty vars, got %+v", got.Vars)
	}
}
