package secrets

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
)

func TestVault_EncryptDecryptRoundTrip(t *testing.T) {
	v := newTestVault(t)
	enc, err := v.Encrypt("hello world")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !strings.HasPrefix(enc, envelopePrefix) {
		t.Fatalf("missing prefix: %q", enc)
	}
	got, err := v.Decrypt(enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != "hello world" {
		t.Fatalf("roundtrip mismatch: %q", got)
	}
}

func TestVault_EncryptEmptyReturnsEmpty(t *testing.T) {
	v := newTestVault(t)
	enc, err := v.Encrypt("")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if enc != "" {
		t.Fatalf("expected empty, got %q", enc)
	}
}

func TestVault_DecryptEmptyReturnsEmpty(t *testing.T) {
	v := newTestVault(t)
	got, err := v.Decrypt("")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestVault_DecryptPlaintextPassthrough(t *testing.T) {
	v := newTestVault(t)
	got, err := v.Decrypt("not encrypted")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != "not encrypted" {
		t.Fatalf("passthrough failed: %q", got)
	}
}

func TestVault_DecryptWrongKeyFails(t *testing.T) {
	v1 := newTestVault(t)
	v2 := newTestVault(t)
	enc, err := v1.Encrypt("classified")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := v2.Decrypt(enc); err == nil {
		t.Fatal("expected failure decrypting with wrong key")
	}
}

func TestVault_DecryptTamperedCiphertext(t *testing.T) {
	v := newTestVault(t)
	enc, err := v.Encrypt("payload")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(enc, envelopePrefix))
	if err != nil {
		t.Fatalf("b64: %v", err)
	}
	raw[len(raw)-1] ^= 0xff
	tampered := envelopePrefix + base64.StdEncoding.EncodeToString(raw)
	if _, err := v.Decrypt(tampered); err == nil {
		t.Fatal("expected failure on tampered ciphertext")
	}
}

func TestVault_DecryptShortCiphertext(t *testing.T) {
	v := newTestVault(t)
	short := envelopePrefix + base64.StdEncoding.EncodeToString([]byte{0x01})
	_, err := v.Decrypt(short)
	if err == nil {
		t.Fatal("expected error for short ciphertext")
	}
	if !strings.Contains(err.Error(), "too short") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVault_DecryptInvalidBase64(t *testing.T) {
	v := newTestVault(t)
	if _, err := v.Decrypt(envelopePrefix + "!!!not base64!!!"); err == nil {
		t.Fatal("expected base64 error")
	}
}

func TestVault_EncryptProducesUniqueOutputs(t *testing.T) {
	v := newTestVault(t)
	seen := make(map[string]struct{}, 16)
	for i := 0; i < 16; i++ {
		enc, err := v.Encrypt("same plaintext")
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		if _, dup := seen[enc]; dup {
			t.Fatal("nonce reuse: duplicate ciphertext for identical plaintext")
		}
		seen[enc] = struct{}{}
	}
}

func TestVault_EncryptBinaryString(t *testing.T) {
	v := newTestVault(t)
	plain := string([]byte{0x00, 0x01, 0xff, 0x7f, 0x80})
	enc, err := v.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	got, err := v.Decrypt(enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != plain {
		t.Fatal("binary content mismatch")
	}
}

func TestKeyFilePath_HonorsEnvOverride(t *testing.T) {
	override := filepath.Join(t.TempDir(), "custom.key")
	t.Setenv(keyfileEnv, override)
	if got := keyFilePath(); got != override {
		t.Fatalf("expected %q got %q", override, got)
	}
}

func TestKeyFilePath_FallsBackToPathsConfig(t *testing.T) {
	t.Setenv(keyfileEnv, "")
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", t.TempDir())
	path := keyFilePath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if filepath.Base(path) != "vault.key" {
		t.Fatalf("expected vault.key basename, got %q", path)
	}
}
