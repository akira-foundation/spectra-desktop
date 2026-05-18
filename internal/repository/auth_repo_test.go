package repository

import (
	"context"
	"testing"
	"time"

	"spectra-desktop/internal/domain"
)

func TestAuthRepository_SaveAndGet(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAuthRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "auth")
	in := domain.ProjectAuth{
		ProjectID: p.ID,
		Scheme:    "bearer",
		Token:     "tok",
		TokenPath: "data.token",
		UserJSON:  `{"id":1}`,
	}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := repo.Get(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.Token != "tok" || got.UserJSON != `{"id":1}` {
		t.Fatalf("unexpected: %+v", got)
	}
	if got.CapturedAt.IsZero() {
		t.Fatalf("expected CapturedAt set")
	}
}

func TestAuthRepository_SaveUpsert(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAuthRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "auth")
	in := domain.ProjectAuth{ProjectID: p.ID, Token: "one"}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatalf("save: %v", err)
	}
	in.Token = "two"
	if err := repo.Save(ctx, in); err != nil {
		t.Fatalf("save again: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID)
	if got == nil || got.Token != "two" {
		t.Fatalf("expected two, got %+v", got)
	}
}

func TestAuthRepository_Get_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewAuthRepository(s.DB)
	got, err := repo.Get(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestAuthRepository_Clear(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAuthRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "auth")
	if err := repo.Save(ctx, domain.ProjectAuth{ProjectID: p.ID, Token: "t"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Clear(ctx, p.ID); err != nil {
		t.Fatalf("clear: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID)
	if got != nil {
		t.Fatalf("expected cleared, got %+v", got)
	}
}

func TestAuthRepository_SavePreservesCapturedAt(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewAuthRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "auth")
	captured := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	in := domain.ProjectAuth{ProjectID: p.ID, Token: "t", CapturedAt: captured}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, _ := repo.Get(ctx, p.ID)
	if got == nil || !got.CapturedAt.Equal(captured) {
		t.Fatalf("expected captured preserved, got %+v", got)
	}
}
