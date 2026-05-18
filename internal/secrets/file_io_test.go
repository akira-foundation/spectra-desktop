package secrets

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapFileInPlace_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	original := []byte("important data here")
	if err := os.WriteFile(src, original, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "pw"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	wrapped, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !IsPassphraseEnvelope(wrapped) {
		t.Fatal("not an envelope after wrap")
	}
	if bytes.Equal(wrapped, original) {
		t.Fatal("file content unchanged after wrap")
	}

	dst := filepath.Join(dir, "data.out")
	if err := UnwrapFileToPath(src, dst, "pw"); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("unwrapped content mismatch")
	}
}

func TestWrapFileInPlace_ZeroByteFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty.bin")
	if err := os.WriteFile(src, nil, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "pw"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	dst := filepath.Join(dir, "empty.out")
	if err := UnwrapFileToPath(src, dst, "pw"); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty file, got %d bytes", len(got))
	}
}

func TestWrapFileInPlace_LargeFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "big.bin")
	payload := randomBytes(t, 1<<20)
	if err := os.WriteFile(src, payload, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "pw"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	dst := filepath.Join(dir, "big.out")
	if err := UnwrapFileToPath(src, dst, "pw"); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("large file content mismatch")
	}
}

func TestWrapFileInPlace_MissingSource(t *testing.T) {
	err := WrapFileInPlace(filepath.Join(t.TempDir(), "nope"), "pw")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestUnwrapFileToPath_MissingSource(t *testing.T) {
	err := UnwrapFileToPath(filepath.Join(t.TempDir(), "nope"), filepath.Join(t.TempDir(), "out"), "pw")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestUnwrapFileToPath_WrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "right"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	dst := filepath.Join(dir, "data.out")
	err := UnwrapFileToPath(src, dst, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
	if !strings.Contains(err.Error(), "invalid passphrase") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnwrapFileToPath_CorruptSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "pw"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	wrapped, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	wrapped[len(wrapped)-1] ^= 0xff
	if err := os.WriteFile(src, wrapped, 0o600); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	dst := filepath.Join(dir, "data.out")
	if err := UnwrapFileToPath(src, dst, "pw"); err == nil {
		t.Fatal("expected error for corrupt source")
	}
}

func TestUnwrapFileToPath_DestUnwritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, "pw"); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	readOnly := filepath.Join(dir, "ro")
	if err := os.Mkdir(readOnly, 0o500); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dst := filepath.Join(readOnly, "out")
	if err := UnwrapFileToPath(src, dst, "pw"); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestWrapFileInPlace_EmptyPassphraseRejected(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WrapFileInPlace(src, ""); err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}
