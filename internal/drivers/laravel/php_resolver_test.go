package laravel

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func resetPHPResolver(t *testing.T) {
	t.Helper()
	phpOverrideMu.Lock()
	phpOverride = ""
	phpOverrideMu.Unlock()
	phpResolveOnce = sync.Once{}
	phpResolved = ""
	phpResolveErr = nil
}

func TestIsExecutableFile_TrueForExecutable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isExecutableFile(p) {
		t.Fatal("want true")
	}
}

func TestIsExecutableFile_FalseForNonExecutable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if isExecutableFile(p) {
		t.Fatal("want false for 0644")
	}
}

func TestIsExecutableFile_FalseForDirOrMissing(t *testing.T) {
	if isExecutableFile(t.TempDir()) {
		t.Fatal("dir should be false")
	}
	if isExecutableFile(filepath.Join(t.TempDir(), "nope")) {
		t.Fatal("missing should be false")
	}
}

func TestSetPHPBinaryOverride_StoresAndTrims(t *testing.T) {
	resetPHPResolver(t)
	t.Cleanup(func() { resetPHPResolver(t) })
	SetPHPBinaryOverride("  /a/b  ")
	if currentPHPOverride() != "/a/b" {
		t.Fatalf("got %q", currentPHPOverride())
	}
}

func TestResolvePHPBinary_OverrideWins(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix-only fixture")
	}
	resetPHPResolver(t)
	t.Cleanup(func() { resetPHPResolver(t) })
	dir := t.TempDir()
	stub := filepath.Join(dir, "phpstub")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	SetPHPBinaryOverride(stub)
	got, err := ResolvePHPBinaryPath()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != stub {
		t.Fatalf("want %s, got %s", stub, got)
	}
}

func TestResolvePHPBinary_OverrideIgnoredWhenNotExecutable(t *testing.T) {
	resetPHPResolver(t)
	t.Cleanup(func() { resetPHPResolver(t) })
	skipIfSystemPHPPresent(t)
	SetPHPBinaryOverride("/definitely/not/here-php-xyz")
	t.Setenv("PATH", "/definitely-empty")
	t.Setenv("SHELL", "/definitely/missing/sh")
	_, err := ResolvePHPBinaryPath()
	if !errors.Is(err, ErrPHPNotFound) {
		t.Fatalf("want ErrPHPNotFound, got %v", err)
	}
}

func TestResolvePHPBinary_AllFail(t *testing.T) {
	resetPHPResolver(t)
	t.Cleanup(func() { resetPHPResolver(t) })
	skipIfSystemPHPPresent(t)
	t.Setenv("PATH", "/definitely-empty")
	t.Setenv("SHELL", "/definitely/missing/shell")
	_, err := ResolvePHPBinaryPath()
	if !errors.Is(err, ErrPHPNotFound) {
		t.Fatalf("want ErrPHPNotFound, got %v", err)
	}
}

func skipIfSystemPHPPresent(t *testing.T) {
	t.Helper()
	for _, p := range systemPHPCandidates() {
		if isExecutableFile(p) {
			t.Skipf("system php present at %s; cannot exercise all-fail branch", p)
		}
	}
}

func TestLookupPHPViaLoginShell_NoShellReturnsEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix-only")
	}
	t.Setenv("SHELL", "/definitely/missing/shell")
	if got := lookupPHPViaLoginShell(); got != "" {
		t.Fatalf("want empty, got %s", got)
	}
}

func TestLookupPHPViaLoginShell_DeterministicShellNoPhp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix-only")
	}
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("/bin/sh not present")
	}
	t.Setenv("SHELL", "/bin/sh")
	t.Setenv("PATH", "/definitely-empty")
	got := lookupPHPViaLoginShell()
	if got != "" {
		t.Fatalf("want empty (no php on empty PATH), got %s", got)
	}
}

func TestSystemPHPCandidates_NotEmpty(t *testing.T) {
	got := systemPHPCandidates()
	if len(got) == 0 {
		t.Fatal("want candidates")
	}
}

func TestSplitPath(t *testing.T) {
	if splitPath("") != nil {
		t.Fatal("empty => nil")
	}
	sep := string(os.PathListSeparator)
	got := splitPath("/a" + sep + "/b")
	if !equalStringSlice(got, []string{"/a", "/b"}) {
		t.Fatalf("got %v", got)
	}
}

func TestMergeUniquePathSegments_DropsDuplicatesAndEmpties(t *testing.T) {
	sep := string(os.PathListSeparator)
	existing := "/a" + sep + "/b"
	extras := []string{"/a", "/c"}
	got := mergeUniquePathSegments(existing, extras)
	parts := strings.Split(got, sep)
	if !equalStringSlice(parts, []string{"/a", "/b", "/c"}) {
		t.Fatalf("got %v", parts)
	}
}

func TestEnrichedEnv_SetsPATH(t *testing.T) {
	t.Setenv("PATH", "/x")
	t.Setenv("SHELL", "/definitely/missing")
	env := enrichedEnv()
	found := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			found = true
			if !strings.Contains(kv, "/x") {
				t.Fatalf("PATH missing /x: %s", kv)
			}
		}
	}
	if !found {
		t.Fatal("PATH entry missing")
	}
}

func TestLoginShellPath_MissingShellReturnsEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix-only")
	}
	t.Setenv("SHELL", "/definitely/missing/shell")
	if got := loginShellPath(); got != "" {
		t.Fatalf("want empty, got %s", got)
	}
}
