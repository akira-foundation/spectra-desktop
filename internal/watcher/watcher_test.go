package watcher

import (
	"testing"
)

func TestNew_ReturnsNonNil(t *testing.T) {
	w := New()
	if w == nil {
		t.Fatal("New() returned nil")
	}
}

func TestWatcher_StartReturnsNil(t *testing.T) {
	w := New()
	if err := w.Start(t.TempDir()); err != nil {
		t.Fatalf("Start: unexpected error: %v", err)
	}
}

func TestWatcher_StartEmptyPath(t *testing.T) {
	w := New()
	if err := w.Start(""); err != nil {
		t.Fatalf("Start(\"\"): unexpected error: %v", err)
	}
}

func TestWatcher_StartMissingDirectory(t *testing.T) {
	w := New()
	if err := w.Start("/nonexistent/path/does/not/exist"); err != nil {
		t.Fatalf("Start(missing): unexpected error: %v", err)
	}
}

func TestWatcher_StopReturnsNil(t *testing.T) {
	w := New()
	if err := w.Stop(); err != nil {
		t.Fatalf("Stop: unexpected error: %v", err)
	}
}

func TestWatcher_StopIdempotent(t *testing.T) {
	w := New()
	for i := 0; i < 3; i++ {
		if err := w.Stop(); err != nil {
			t.Fatalf("Stop iter %d: unexpected error: %v", i, err)
		}
	}
}

func TestWatcher_DoubleStart(t *testing.T) {
	w := New()
	dir := t.TempDir()
	if err := w.Start(dir); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	if err := w.Start(dir); err != nil {
		t.Fatalf("second Start: %v", err)
	}
}

func TestWatcher_StartStopCycle(t *testing.T) {
	w := New()
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		if err := w.Start(dir); err != nil {
			t.Fatalf("Start iter %d: %v", i, err)
		}
		if err := w.Stop(); err != nil {
			t.Fatalf("Stop iter %d: %v", i, err)
		}
	}
}

func TestEvent_FieldsAssignable(t *testing.T) {
	e := Event{Path: "/tmp/foo", Op: "CREATE"}
	if e.Path != "/tmp/foo" {
		t.Fatalf("Path: got %q", e.Path)
	}
	if e.Op != "CREATE" {
		t.Fatalf("Op: got %q", e.Op)
	}
}

func TestEvent_ZeroValue(t *testing.T) {
	var e Event
	if e.Path != "" || e.Op != "" {
		t.Fatalf("zero Event not empty: %+v", e)
	}
}
