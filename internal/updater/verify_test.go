package updater

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aead.dev/minisign"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func generateKey(t *testing.T) (minisign.PublicKey, minisign.PrivateKey) {
	t.Helper()
	pub, priv, err := minisign.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	return pub, priv
}

func installKey(t *testing.T, pub minisign.PublicKey) {
	t.Helper()
	text, err := pub.MarshalText()
	if err != nil {
		t.Fatalf("marshal pub: %v", err)
	}
	prev := MinisignPublicKey
	MinisignPublicKey = string(text)
	t.Cleanup(func() { MinisignPublicKey = prev })
}

func TestVerify_GoodSignaturePasses(t *testing.T) {
	pub, priv := generateKey(t)
	installKey(t, pub)

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "artifact.zip")
	sigPath := filepath.Join(dir, "artifact.zip.sig")

	payload := []byte("this is the artifact body")
	sig := minisign.Sign(priv, payload)
	writeFile(t, zipPath, payload)
	writeFile(t, sigPath, sig)

	if err := verify(zipPath, sigPath); err != nil {
		t.Fatalf("verify good signature: %v", err)
	}
}

func TestVerify_TamperedFileRejects(t *testing.T) {
	pub, priv := generateKey(t)
	installKey(t, pub)

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "artifact.zip")
	sigPath := filepath.Join(dir, "artifact.zip.sig")

	payload := []byte("original payload")
	sig := minisign.Sign(priv, payload)

	writeFile(t, zipPath, []byte("tampered payload"))
	writeFile(t, sigPath, sig)

	err := verify(zipPath, sigPath)
	if err == nil {
		t.Fatal("expected verification to fail for tampered file")
	}
	if !strings.Contains(err.Error(), "signature invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerify_WrongKeyRejects(t *testing.T) {
	_, signer := generateKey(t)
	otherPub, _ := generateKey(t)
	installKey(t, otherPub)

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "artifact.zip")
	sigPath := filepath.Join(dir, "artifact.zip.sig")

	payload := []byte("payload")
	sig := minisign.Sign(signer, payload)
	writeFile(t, zipPath, payload)
	writeFile(t, sigPath, sig)

	if err := verify(zipPath, sigPath); err == nil {
		t.Fatal("expected verification to fail under wrong key")
	}
}

func TestVerify_UnconfiguredKey(t *testing.T) {
	prev := MinisignPublicKey
	MinisignPublicKey = ""
	t.Cleanup(func() { MinisignPublicKey = prev })

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "a.zip")
	sigPath := filepath.Join(dir, "a.zip.sig")
	writeFile(t, zipPath, []byte("x"))
	writeFile(t, sigPath, []byte("x"))

	err := verify(zipPath, sigPath)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected not-configured error, got %v", err)
	}
}

func TestVerify_PlaceholderKey(t *testing.T) {
	prev := MinisignPublicKey
	MinisignPublicKey = "REPLACE_ME"
	t.Cleanup(func() { MinisignPublicKey = prev })

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "a.zip")
	sigPath := filepath.Join(dir, "a.zip.sig")
	writeFile(t, zipPath, []byte("x"))
	writeFile(t, sigPath, []byte("x"))

	if err := verify(zipPath, sigPath); err == nil {
		t.Fatal("expected error for placeholder key")
	}
}

func TestVerify_MalformedKey(t *testing.T) {
	prev := MinisignPublicKey
	MinisignPublicKey = "not-a-valid-minisign-key"
	t.Cleanup(func() { MinisignPublicKey = prev })

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "a.zip")
	sigPath := filepath.Join(dir, "a.zip.sig")
	writeFile(t, zipPath, []byte("x"))
	writeFile(t, sigPath, []byte("x"))

	err := verify(zipPath, sigPath)
	if err == nil || !strings.Contains(err.Error(), "minisign public key") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestVerify_MissingFiles(t *testing.T) {
	pub, _ := generateKey(t)
	installKey(t, pub)

	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.zip")
	sigPath := filepath.Join(dir, "x.sig")
	writeFile(t, sigPath, []byte("x"))

	if err := verify(missing, sigPath); err == nil {
		t.Fatal("expected error for missing zip")
	}

	zipPath := filepath.Join(dir, "x.zip")
	writeFile(t, zipPath, []byte("x"))
	if err := verify(zipPath, filepath.Join(dir, "missing.sig")); err == nil {
		t.Fatal("expected error for missing sig")
	}
}
