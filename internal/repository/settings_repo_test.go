package repository

import (
	"context"
	"testing"
)

func TestSettingsRepository_Get_MissingReturnsEmpty(t *testing.T) {
	s := newStorage(t)
	repo := NewSettingsRepository(s.DB)
	got, err := repo.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestSettingsRepository_SetAndGet(t *testing.T) {
	s := newStorage(t)
	repo := NewSettingsRepository(s.DB)
	ctx := context.Background()

	if err := repo.Set(ctx, "theme", "dark"); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := repo.Get(ctx, "theme")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "dark" {
		t.Fatalf("expected dark, got %q", got)
	}
}

func TestSettingsRepository_Set_Upsert(t *testing.T) {
	s := newStorage(t)
	repo := NewSettingsRepository(s.DB)
	ctx := context.Background()

	if err := repo.Set(ctx, "k", "v1"); err != nil {
		t.Fatalf("set 1: %v", err)
	}
	if err := repo.Set(ctx, "k", "v2"); err != nil {
		t.Fatalf("set 2: %v", err)
	}
	got, _ := repo.Get(ctx, "k")
	if got != "v2" {
		t.Fatalf("expected v2, got %q", got)
	}
}

func TestSettingsRepository_Delete(t *testing.T) {
	s := newStorage(t)
	repo := NewSettingsRepository(s.DB)
	ctx := context.Background()
	if err := repo.Set(ctx, "k", "v"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := repo.Delete(ctx, "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.Get(ctx, "k")
	if got != "" {
		t.Fatalf("expected empty after delete, got %q", got)
	}
}

func TestSettingsRepository_Delete_Missing(t *testing.T) {
	s := newStorage(t)
	repo := NewSettingsRepository(s.DB)
	if err := repo.Delete(context.Background(), "missing"); err != nil {
		t.Fatalf("delete missing: %v", err)
	}
}
