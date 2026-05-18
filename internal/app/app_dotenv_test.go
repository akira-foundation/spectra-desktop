package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractDotenvKey_FindsMatchingKey(t *testing.T) {
	content := "FOO=bar\nAPP_URL=https://app.test\nOTHER=1\n"
	if got := extractDotenvKey(content, "APP_URL"); got != "https://app.test" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractDotenvKey_StripsQuotes(t *testing.T) {
	if got := extractDotenvKey(`APP_URL="https://x.test"`, "APP_URL"); got != "https://x.test" {
		t.Fatalf("got %q", got)
	}
	if got := extractDotenvKey(`APP_URL='https://x.test'`, "APP_URL"); got != "https://x.test" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractDotenvKey_SkipsCommentsAndBlanks(t *testing.T) {
	content := "# comment\n\n  # also\nAPP_URL=https://x.test\n"
	if got := extractDotenvKey(content, "APP_URL"); got != "https://x.test" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractDotenvKey_MissingKeyReturnsEmpty(t *testing.T) {
	if got := extractDotenvKey("FOO=bar", "APP_URL"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractDotenvKey_EmptyValueIgnored(t *testing.T) {
	if got := extractDotenvKey("APP_URL=\n", "APP_URL"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestReadDotenvAppURL_PrefersEnvOverExample(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_URL=https://real.test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("APP_URL=https://example.test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readDotenvAppURL(dir); got != "https://real.test" {
		t.Fatalf("got %q", got)
	}
}

func TestReadDotenvAppURL_FallsThroughToExample(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("APP_URL=https://example.test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readDotenvAppURL(dir); got != "https://example.test" {
		t.Fatalf("got %q", got)
	}
}

func TestReadDotenvAppURL_NoFilesReturnsEmpty(t *testing.T) {
	if got := readDotenvAppURL(t.TempDir()); got != "" {
		t.Fatalf("got %q", got)
	}
}
