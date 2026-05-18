package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	}
	return home
}

func openMigrated(t *testing.T, path string) *Storage {
	t.Helper()
	s := New()
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

func fileHash(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("hash %s: %v", path, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
