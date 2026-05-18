package updater

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestInstall_NonDarwinReturnsError(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skipf("darwin platform: cannot test non-darwin gate")
	}
	err := Install(context.Background(), &UpdateInfo{URL: "https://example.test/x.zip"}, nil)
	if err == nil {
		t.Fatal("expected gating error on non-darwin")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected not-supported message, got %v", err)
	}
}

func TestInstall_NilInfoOnDarwinReturnsError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skipf("darwin-only test on %s", runtime.GOOS)
	}
	err := Install(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for nil info")
	}
}

func TestUnzip_ExtractsFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.zip")
	writeZip(t, src, map[string]string{
		"a.txt":       "hello",
		"sub/b.txt":   "world",
		"sub/c/d.txt": "nested",
	})

	dest := filepath.Join(dir, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := unzip(src, dest); err != nil {
		t.Fatalf("unzip: %v", err)
	}

	cases := map[string]string{
		"a.txt":       "hello",
		"sub/b.txt":   "world",
		"sub/c/d.txt": "nested",
	}
	for rel, want := range cases {
		got, err := os.ReadFile(filepath.Join(dest, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, want %q", rel, got, want)
		}
	}
}

func TestUnzip_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "evil.zip")

	f, err := os.Create(src)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("../escape.txt")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	_, _ = w.Write([]byte("bad"))
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	f.Close()

	dest := filepath.Join(dir, "out")
	_ = os.MkdirAll(dest, 0o755)
	err = unzip(src, dest)
	if err == nil {
		t.Fatal("expected path-traversal rejection")
	}
	if !strings.Contains(err.Error(), "escapes dest") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnzip_MissingSourceErrors(t *testing.T) {
	dir := t.TempDir()
	if err := unzip(filepath.Join(dir, "nope.zip"), dir); err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestFindAppBundle_FindsDotApp(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "Spectra.app")
	if err := os.MkdirAll(bundle, 0o755); err != nil {
		t.Fatalf("mkdir bundle: %v", err)
	}
	got, err := findAppBundle(dir)
	if err != nil {
		t.Fatalf("findAppBundle: %v", err)
	}
	if got != bundle {
		t.Fatalf("got %q, want %q", got, bundle)
	}
}

func TestFindAppBundle_NoBundleErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "NotAnApp"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := findAppBundle(dir); err == nil {
		t.Fatal("expected error when no .app bundle present")
	}
}

func TestFindAppBundle_MissingDirErrors(t *testing.T) {
	if _, err := findAppBundle(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestSwapBundle_MovesAndBackups(t *testing.T) {
	dir := t.TempDir()
	currentApp := filepath.Join(dir, "Spectra.app")
	newApp := filepath.Join(dir, "Spectra-new.app")

	if err := os.MkdirAll(currentApp, 0o755); err != nil {
		t.Fatalf("mkdir current: %v", err)
	}
	if err := os.WriteFile(filepath.Join(currentApp, "marker.txt"), []byte("old"), 0o600); err != nil {
		t.Fatalf("write old marker: %v", err)
	}
	if err := os.MkdirAll(newApp, 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newApp, "marker.txt"), []byte("new"), 0o600); err != nil {
		t.Fatalf("write new marker: %v", err)
	}

	if err := swapBundle(currentApp, newApp); err != nil {
		t.Fatalf("swapBundle: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(currentApp, "marker.txt"))
	if err != nil {
		t.Fatalf("read installed marker: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("installed bundle marker = %q, want %q", got, "new")
	}

	backup := currentApp + ".old"
	got, err = os.ReadFile(filepath.Join(backup, "marker.txt"))
	if err != nil {
		t.Fatalf("read backup marker: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("backup marker = %q, want %q", got, "old")
	}
}

func TestSwapBundle_OverwritesExistingBackup(t *testing.T) {
	dir := t.TempDir()
	currentApp := filepath.Join(dir, "Spectra.app")
	newApp := filepath.Join(dir, "Spectra-new.app")
	if err := os.MkdirAll(currentApp, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(newApp, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	staleBackup := currentApp + ".old"
	if err := os.MkdirAll(staleBackup, 0o755); err != nil {
		t.Fatalf("mkdir stale: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleBackup, "stale.txt"), []byte("stale"), 0o600); err != nil {
		t.Fatalf("write stale: %v", err)
	}

	if err := swapBundle(currentApp, newApp); err != nil {
		t.Fatalf("swapBundle: %v", err)
	}
	if _, err := os.Stat(filepath.Join(staleBackup, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale backup to be removed, stat err = %v", err)
	}
}

func TestCurrentAppBundlePath_ReturnsErrorWhenNotInBundle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("path walk assumptions differ on windows")
	}
	if strings.Contains(os.Args[0], ".app/") {
		t.Skipf("test runner appears to be inside a .app bundle")
	}
	_, err := currentAppBundlePath()
	if err == nil {
		t.Fatal("expected error: test binary should not be inside a .app bundle")
	}
}
