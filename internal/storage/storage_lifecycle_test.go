package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestDefaultPath_UsesIsolatedHome(t *testing.T) {
	home := isolateHome(t)

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if !strings.HasPrefix(got, home) {
		t.Fatalf("expected DefaultPath inside %s, got %s", home, got)
	}
	if filepath.Base(got) != dbFile {
		t.Fatalf("expected basename %s, got %s", dbFile, filepath.Base(got))
	}
	if !strings.Contains(got, appFolder()) {
		t.Fatalf("expected path to contain app folder %q: %s", appFolder(), got)
	}
}

func TestStorage_OpenAndCloseLifecycle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lifecycle.db")

	s := New()
	if err := s.Open(path); err != nil {
		t.Fatalf("open: %v", err)
	}
	if s.DB == nil || s.sql == nil {
		t.Fatal("expected DB and sql handles populated")
	}
	if s.dbPath != path {
		t.Fatalf("expected dbPath %s, got %s", path, s.dbPath)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file created: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestStorage_OpenDefaultPathWhenEmpty(t *testing.T) {
	isolateHome(t)
	s := New()
	if err := s.Open(""); err != nil {
		t.Fatalf("open default: %v", err)
	}
	defer s.Close()

	want, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if s.dbPath != want {
		t.Fatalf("expected dbPath %s, got %s", want, s.dbPath)
	}
}

func TestStorage_MigrateIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "idem.db")

	s1 := openMigrated(t, path)
	if err := s1.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}

	s2 := openMigrated(t, path)

	row := s2.sql.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='projects'")
	var n int
	if err := row.Scan(&n); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected projects table after reopen, got count=%d", n)
	}

	if err := s2.Migrate(context.Background()); err != nil {
		t.Fatalf("third migrate: %v", err)
	}
}

func TestStorage_ConcurrentOpenSameFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.db")

	if err := openMigrated(t, path).Close(); err != nil {
		t.Fatalf("seed close: %v", err)
	}

	const n = 4
	errs := make([]error, n)
	storages := make([]*Storage, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			s := New()
			if err := s.Open(path); err != nil {
				errs[i] = err
				return
			}
			storages[i] = s
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent open %d: %v", i, err)
		}
	}
	for _, s := range storages {
		if s != nil {
			_ = s.Close()
		}
	}
}
