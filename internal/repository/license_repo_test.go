package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func TestLicenseRepository_Get_BaselineSeededRow(t *testing.T) {
	s := newStorage(t)
	repo := NewLicenseRepository(s.DB)
	got, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatalf("expected seeded singleton row, got nil")
	}
	if got.ID != licenseRowID || got.Status != "inactive" {
		t.Fatalf("expected inactive baseline, got %+v", got)
	}
}

func TestLicenseRepository_SaveAndGet(t *testing.T) {
	s := newStorage(t)
	repo := NewLicenseRepository(s.DB)
	ctx := context.Background()

	in := domain.License{
		CustomerEmail: "u@example.com",
		Plan:          "pro",
		Status:        "active",
		FeaturesJSON:  `{"x":1}`,
	}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.ID != licenseRowID || got.Plan != "pro" || got.CustomerEmail != "u@example.com" {
		t.Fatalf("unexpected: %+v", got)
	}
	if got.UpdatedAt.IsZero() {
		t.Fatalf("expected updated_at set")
	}
}

func TestLicenseRepository_SaveIsSingleton(t *testing.T) {
	s := newStorage(t)
	repo := NewLicenseRepository(s.DB)
	ctx := context.Background()

	if err := repo.Save(ctx, domain.License{Plan: "a", Status: "active"}); err != nil {
		t.Fatalf("save 1: %v", err)
	}
	if err := repo.Save(ctx, domain.License{Plan: "b", Status: "active"}); err != nil {
		t.Fatalf("save 2: %v", err)
	}
	got, _ := repo.Get(ctx)
	if got == nil || got.Plan != "b" {
		t.Fatalf("expected last write wins, got %+v", got)
	}
}

func TestLicenseRepository_Clear(t *testing.T) {
	s := newStorage(t)
	repo := NewLicenseRepository(s.DB)
	ctx := context.Background()
	if err := repo.Save(ctx, domain.License{Plan: "pro", Status: "active"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Clear(ctx); err != nil {
		t.Fatalf("clear: %v", err)
	}
	got, _ := repo.Get(ctx)
	if got == nil {
		t.Fatalf("clear should keep singleton row, got nil")
	}
	if got.Status != "inactive" || got.Plan != "" {
		t.Fatalf("expected inactive: %+v", got)
	}
}

func TestUsageBufferRepository_AppendAndPending(t *testing.T) {
	s := newStorage(t)
	repo := NewUsageBufferRepository(s.DB)
	ctx := context.Background()

	now := time.Now().UTC()
	e1 := domain.UsageBufferEntry{ID: uuid.NewString(), Feature: "f", Amount: 1, OccurredAt: now.Add(-1 * time.Hour)}
	e2 := domain.UsageBufferEntry{ID: uuid.NewString(), Feature: "f", Amount: 2, OccurredAt: now}
	if err := repo.Append(ctx, e1); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if err := repo.Append(ctx, e2); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	rows, err := repo.PendingBatch(ctx, 10)
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2, got %d", len(rows))
	}
	if rows[0].ID != e1.ID {
		t.Fatalf("expected oldest first, got %s", rows[0].ID)
	}
}

func TestUsageBufferRepository_PendingBatch_DefaultLimit(t *testing.T) {
	s := newStorage(t)
	repo := NewUsageBufferRepository(s.DB)
	ctx := context.Background()

	now := time.Now().UTC()
	if err := repo.Append(ctx, domain.UsageBufferEntry{ID: uuid.NewString(), Feature: "f", Amount: 1, OccurredAt: now}); err != nil {
		t.Fatalf("append: %v", err)
	}
	rows, err := repo.PendingBatch(ctx, 0)
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1, got %d", len(rows))
	}
}

func TestUsageBufferRepository_MarkFlushed(t *testing.T) {
	s := newStorage(t)
	repo := NewUsageBufferRepository(s.DB)
	ctx := context.Background()

	now := time.Now().UTC()
	e := domain.UsageBufferEntry{ID: uuid.NewString(), Feature: "f", Amount: 1, OccurredAt: now}
	if err := repo.Append(ctx, e); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := repo.MarkFlushed(ctx, []string{e.ID}); err != nil {
		t.Fatalf("flush: %v", err)
	}
	rows, _ := repo.PendingBatch(ctx, 10)
	if len(rows) != 0 {
		t.Fatalf("expected none pending, got %d", len(rows))
	}
}

func TestUsageBufferRepository_MarkFlushed_Empty(t *testing.T) {
	s := newStorage(t)
	repo := NewUsageBufferRepository(s.DB)
	if err := repo.MarkFlushed(context.Background(), nil); err != nil {
		t.Fatalf("empty flush: %v", err)
	}
}
