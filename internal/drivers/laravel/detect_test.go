package laravel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_EmptyPath(t *testing.T) {
	got := detect("")
	if got.Detected || got.Confidence != 0 || len(got.Markers) != 0 {
		t.Fatalf("want zero result, got %+v", got)
	}
}

func TestDetect_NoMarkers(t *testing.T) {
	dir := t.TempDir()
	got := detect(dir)
	if got.Detected || len(got.Markers) != 0 {
		t.Fatalf("want empty result, got %+v", got)
	}
}

func TestDetect_SingleMarkerBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	mustTouch(t, filepath.Join(dir, "artisan"))
	got := detect(dir)
	if got.Detected {
		t.Fatalf("single marker should not satisfy threshold: %+v", got)
	}
	if len(got.Markers) != 1 || got.Markers[0] != "artisan" {
		t.Fatalf("want marker [artisan], got %+v", got.Markers)
	}
}

func TestDetect_AboveThreshold(t *testing.T) {
	dir := t.TempDir()
	mustTouch(t, filepath.Join(dir, "artisan"))
	mustTouch(t, filepath.Join(dir, "composer.json"))
	mustMkdir(t, filepath.Join(dir, "routes"))
	mustTouch(t, filepath.Join(dir, "routes", "web.php"))
	got := detect(dir)
	if !got.Detected {
		t.Fatalf("want detected, got %+v", got)
	}
	if got.Confidence < 0.4 {
		t.Fatalf("want confidence >= 0.4, got %v", got.Confidence)
	}
	if len(got.Markers) != 3 {
		t.Fatalf("want 3 markers, got %+v", got.Markers)
	}
}

func TestDetect_AllMarkers(t *testing.T) {
	dir := t.TempDir()
	for _, m := range detectionMarkers {
		full := filepath.Join(dir, m)
		mustMkdir(t, filepath.Dir(full))
		mustTouch(t, full)
	}
	got := detect(dir)
	if !got.Detected {
		t.Fatalf("want detected")
	}
	if got.Confidence != 1.0 {
		t.Fatalf("want confidence 1.0, got %v", got.Confidence)
	}
}

func mustTouch(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	_ = f.Close()
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
