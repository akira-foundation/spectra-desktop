package app

import (
	"strings"
	"testing"
	"time"
)

func TestSuggestExportFilename_LowercasesAndKeepsAllowed(t *testing.T) {
	got := suggestExportFilename("My-Project_42")
	stamp := time.Now().UTC().Format("20060102")
	want := "my-project_42-" + stamp + ".spectra"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSuggestExportFilename_ReplacesUnicodeAndSpaces(t *testing.T) {
	got := suggestExportFilename("Hello World Água")
	stamp := time.Now().UTC().Format("20060102")
	if !strings.HasSuffix(got, "-"+stamp+".spectra") {
		t.Fatalf("missing timestamp suffix: %s", got)
	}
	prefix := strings.TrimSuffix(got, "-"+stamp+".spectra")
	for _, r := range prefix {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			t.Fatalf("unexpected char %q in %q", r, prefix)
		}
	}
}

func TestSuggestExportFilename_EmptyName(t *testing.T) {
	got := suggestExportFilename("")
	stamp := time.Now().UTC().Format("20060102")
	want := "-" + stamp + ".spectra"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestPendingImportStash_RoundTrip(t *testing.T) {
	if _, ok := consumePendingImport(); ok {
		t.Fatal("expected empty pending import at start")
	}
	stashPendingImport("/tmp/x.spectra")
	path, ok := consumePendingImport()
	if !ok || path != "/tmp/x.spectra" {
		t.Fatalf("got %q ok=%v", path, ok)
	}
	if _, ok := consumePendingImport(); ok {
		t.Fatal("expected empty after consume")
	}
}

func TestPendingRestoreStash_RoundTrip(t *testing.T) {
	if _, ok := consumePendingRestore(); ok {
		t.Fatal("expected empty pending restore at start")
	}
	stashPendingRestore("/tmp/r.spectra")
	path, ok := consumePendingRestore()
	if !ok || path != "/tmp/r.spectra" {
		t.Fatalf("got %q ok=%v", path, ok)
	}
}

func TestRemoveFile_MissingIsNoop(t *testing.T) {
	removeFile("/nonexistent/path/should/not/panic")
}
