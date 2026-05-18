package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func seedHistory(t *testing.T, repo *HistoryRepository, projectID, endpointID string, status int, age time.Duration) domain.HistoryEntry {
	t.Helper()
	entry := domain.HistoryEntry{
		ID:             uuid.NewString(),
		ProjectID:      projectID,
		EndpointID:     endpointID,
		Method:         "GET",
		URL:            "http://localhost/users",
		ResponseStatus: status,
		DurationMs:     50,
		CreatedAt:      time.Now().UTC().Add(-age),
	}
	if err := repo.Save(context.Background(), entry); err != nil {
		t.Fatalf("save history: %v", err)
	}
	return entry
}

func TestHistoryRepository_SaveAndGetByID(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	entry := domain.HistoryEntry{
		ProjectID:      p.ID,
		EndpointID:     "ep-1",
		Method:         "POST",
		URL:            "http://x/y",
		ResponseStatus: 201,
	}
	if err := repo.Save(ctx, entry); err != nil {
		t.Fatalf("save: %v", err)
	}

	list, err := repo.List(ctx, p.ID, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}
	got, err := repo.GetByID(ctx, list[0].ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.Method != "POST" || got.URL != "http://x/y" {
		t.Fatalf("round-trip: %+v", got)
	}
}

func TestHistoryRepository_GetByID_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewHistoryRepository(s.DB)
	got, err := repo.GetByID(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestHistoryRepository_List_OrderedDescAndLimited(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	seedHistory(t, repo, p.ID, "ep-1", 200, 3*time.Hour)
	mid := seedHistory(t, repo, p.ID, "ep-1", 200, 2*time.Hour)
	newest := seedHistory(t, repo, p.ID, "ep-1", 200, 1*time.Hour)

	list, err := repo.List(ctx, p.ID, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[0].ID != newest.ID || list[1].ID != mid.ID {
		t.Fatalf("ordering wrong: %s, %s", list[0].ID, list[1].ID)
	}
}

func TestHistoryRepository_List_DefaultLimit(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	for i := 0; i < 3; i++ {
		seedHistory(t, repo, p.ID, "ep", 200, time.Duration(i)*time.Hour)
	}
	list, err := repo.List(ctx, p.ID, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}
}

func TestHistoryRepository_LatestSuccessByEndpoint(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	seedHistory(t, repo, p.ID, "ep-1", 500, 3*time.Hour)
	older := seedHistory(t, repo, p.ID, "ep-1", 200, 2*time.Hour)
	seedHistory(t, repo, p.ID, "ep-2", 200, 1*time.Hour)

	got, err := repo.LatestSuccessByEndpoint(ctx, p.ID, "ep-1")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got == nil || got.ID != older.ID {
		t.Fatalf("expected %s, got %+v", older.ID, got)
	}

	none, err := repo.LatestSuccessByEndpoint(ctx, p.ID, "ep-none")
	if err != nil {
		t.Fatalf("latest none: %v", err)
	}
	if none != nil {
		t.Fatalf("expected nil, got %+v", none)
	}
}

func TestHistoryRepository_Clear(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	seedHistory(t, repo, p.ID, "ep-1", 200, time.Minute)

	if err := repo.Clear(ctx, p.ID); err != nil {
		t.Fatalf("clear: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if len(list) != 0 {
		t.Fatalf("expected cleared, got %d", len(list))
	}
}

func TestHistoryRepository_TrimOldest(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	for i := 0; i < 5; i++ {
		seedHistory(t, repo, p.ID, "ep", 200, time.Duration(5-i)*time.Hour)
	}

	if err := repo.TrimOldest(ctx, p.ID, 2); err != nil {
		t.Fatalf("trim: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if len(list) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(list))
	}
}

func TestHistoryRepository_TrimOldest_ZeroIsNoop(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	seedHistory(t, repo, p.ID, "ep", 200, time.Minute)

	if err := repo.TrimOldest(ctx, p.ID, 0); err != nil {
		t.Fatalf("trim 0: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if len(list) != 1 {
		t.Fatalf("expected unchanged, got %d", len(list))
	}
}

func TestHistoryRepository_Save_PopulatesIDAndCreatedAt(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewHistoryRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "h")
	if err := repo.Save(ctx, domain.HistoryEntry{ProjectID: p.ID, Method: "GET", URL: "/x"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, _ := repo.List(ctx, p.ID, 10)
	if list[0].ID == "" {
		t.Fatal("expected generated id")
	}
	if list[0].CreatedAt.IsZero() {
		t.Fatal("expected populated created_at")
	}
}
