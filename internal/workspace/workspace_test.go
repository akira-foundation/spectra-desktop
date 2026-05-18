package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewService_ReturnsEmpty(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Fatal("NewService returned nil")
	}
	if s.Current() != nil {
		t.Fatalf("expected nil current, got %#v", s.Current())
	}
}

func TestServiceOpen_Success(t *testing.T) {
	dir := t.TempDir()
	s := NewService()

	ws, err := s.Open(dir)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if ws == nil {
		t.Fatal("expected workspace, got nil")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if ws.Path != absDir {
		t.Errorf("Path = %q, want %q", ws.Path, absDir)
	}
	if ws.Name != filepath.Base(absDir) {
		t.Errorf("Name = %q, want %q", ws.Name, filepath.Base(absDir))
	}
	if s.Current() != ws {
		t.Errorf("Current did not return opened workspace")
	}
}

func TestServiceOpen_EmptyPath(t *testing.T) {
	s := NewService()
	ws, err := s.Open("")
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("err = %v, want ErrInvalidPath", err)
	}
	if ws != nil {
		t.Errorf("expected nil workspace, got %#v", ws)
	}
}

func TestServiceOpen_MissingPath(t *testing.T) {
	s := NewService()
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	ws, err := s.Open(missing)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
	if ws != nil {
		t.Errorf("expected nil workspace, got %#v", ws)
	}
}

func TestServiceOpen_FileNotDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	s := NewService()
	ws, err := s.Open(file)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("err = %v, want ErrInvalidPath", err)
	}
	if ws != nil {
		t.Errorf("expected nil workspace, got %#v", ws)
	}
}

func TestServiceOpen_RelativePathResolvesToAbsolute(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	sub := "child"
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	s := NewService()
	ws, err := s.Open(sub)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !filepath.IsAbs(ws.Path) {
		t.Errorf("Path %q is not absolute", ws.Path)
	}
	if filepath.Base(ws.Path) != sub {
		t.Errorf("Name basename = %q, want %q", filepath.Base(ws.Path), sub)
	}
	if ws.Name != sub {
		t.Errorf("Name = %q, want %q", ws.Name, sub)
	}
}

func TestServiceOpen_ReplacesCurrent(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	s := NewService()

	if _, err := s.Open(first); err != nil {
		t.Fatalf("first Open: %v", err)
	}
	ws, err := s.Open(second)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}

	absSecond, _ := filepath.Abs(second)
	if s.Current().Path != absSecond {
		t.Errorf("Current.Path = %q, want %q", s.Current().Path, absSecond)
	}
	if ws != s.Current() {
		t.Errorf("Current did not match returned workspace")
	}
}

func TestServiceClose_ClearsCurrent(t *testing.T) {
	dir := t.TempDir()
	s := NewService()

	if _, err := s.Open(dir); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if s.Current() == nil {
		t.Fatal("expected current after Open")
	}

	s.Close()
	if s.Current() != nil {
		t.Errorf("Current = %#v after Close, want nil", s.Current())
	}
}

func TestServiceClose_OnFreshService(t *testing.T) {
	s := NewService()
	s.Close()
	if s.Current() != nil {
		t.Errorf("Current = %#v, want nil", s.Current())
	}
}

func TestServiceCurrent_ParallelReadsSafe(t *testing.T) {
	dir := t.TempDir()
	s := NewService()
	if _, err := s.Open(dir); err != nil {
		t.Fatalf("Open: %v", err)
	}

	const readers = 32
	var wg sync.WaitGroup
	wg.Add(readers)
	for range readers {
		go func() {
			defer wg.Done()
			for range 100 {
				if ws := s.Current(); ws == nil || ws.Path == "" {
					t.Errorf("unexpected empty workspace from Current")
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestErrInvalidPath_Message(t *testing.T) {
	if ErrInvalidPath.Error() != "invalid project path" {
		t.Errorf("message = %q", ErrInvalidPath.Error())
	}
}
