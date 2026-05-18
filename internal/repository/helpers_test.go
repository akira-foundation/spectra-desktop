package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/storage"
)

func newStorage(t *testing.T) *storage.Storage {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s := storage.New()
	if err := s.Open(path); err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Migrate(context.Background()); err != nil {
		_ = s.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func seedProject(t *testing.T, repo *ProjectRepository, name string) *domain.Project {
	t.Helper()
	p, err := repo.Save(context.Background(), domain.ProjectInput{
		ID:        uuid.NewString(),
		Name:      name,
		Path:      "/tmp/" + uuid.NewString(),
		Framework: "laravel",
	})
	if err != nil {
		t.Fatalf("seed project %q: %v", name, err)
	}
	return p
}

func seedMetrics(t *testing.T, repo *HistoryRepository, entries []domain.HistoryEntry) {
	t.Helper()
	for _, e := range entries {
		if e.ID == "" {
			e.ID = uuid.NewString()
		}
		if err := repo.Save(context.Background(), e); err != nil {
			t.Fatalf("seed history: %v", err)
		}
	}
}
