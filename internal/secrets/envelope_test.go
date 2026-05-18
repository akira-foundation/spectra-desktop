package secrets

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapWithPassphrase_RoundTrip(t *testing.T) {
	plain := []byte("hello spectra")
	wrapped, err := WrapWithPassphrase(plain, "correct horse")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if !IsPassphraseEnvelope(wrapped) {
		t.Fatal("wrapped output missing magic")
	}
	got, err := UnwrapWithPassphrase(wrapped, "correct horse")
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plain)
	}
}

func TestWrapWithPassphrase_EmptyPassphraseRejected(t *testing.T) {
	if _, err := WrapWithPassphrase([]byte("x"), ""); err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}

func TestWrapWithPassphrase_ZeroByteInput(t *testing.T) {
	wrapped, err := WrapWithPassphrase(nil, "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got, err := UnwrapWithPassphrase(wrapped, "pw")
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty plaintext, got %d bytes", len(got))
	}
}

func TestWrapWithPassphrase_LargeInput(t *testing.T) {
	plain := randomBytes(t, 1<<20)
	wrapped, err := WrapWithPassphrase(plain, "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got, err := UnwrapWithPassphrase(wrapped, "pw")
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("1MB roundtrip mismatch")
	}
}

func TestWrapWithPassphrase_BinaryContent(t *testing.T) {
	plain := []byte{0x00, 0xff, 0x7f, 0x80, 0x01, 0x02, 0xfe}
	wrapped, err := WrapWithPassphrase(plain, "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got, err := UnwrapWithPassphrase(wrapped, "pw")
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("binary roundtrip mismatch")
	}
}

func TestUnwrapWithPassphrase_WrongPassphrase(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("secret"), "right")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	_, err = UnwrapWithPassphrase(wrapped, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
	if !strings.Contains(err.Error(), "invalid passphrase") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnwrapWithPassphrase_TamperedCiphertext(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("secret payload"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	wrapped[len(wrapped)-1] ^= 0x01
	if _, err := UnwrapWithPassphrase(wrapped, "pw"); err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestUnwrapWithPassphrase_TamperedNonce(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("secret"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	nonceOffset := len(envelopeMagic) + 4 + envelopeSaltLen
	wrapped[nonceOffset] ^= 0xff
	if _, err := UnwrapWithPassphrase(wrapped, "pw"); err == nil {
		t.Fatal("expected error for tampered nonce")
	}
}

func TestUnwrapWithPassphrase_TamperedSalt(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("secret"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	saltOffset := len(envelopeMagic) + 4
	wrapped[saltOffset] ^= 0xff
	if _, err := UnwrapWithPassphrase(wrapped, "pw"); err == nil {
		t.Fatal("expected error for tampered salt")
	}
}

func TestUnwrapWithPassphrase_MissingMagic(t *testing.T) {
	_, err := UnwrapWithPassphrase([]byte("not an envelope at all"), "pw")
	if err == nil {
		t.Fatal("expected error for missing magic")
	}
	if !strings.Contains(err.Error(), "missing magic") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnwrapWithPassphrase_TruncatedHeader(t *testing.T) {
	short := []byte(envelopeMagic + "\x00\x00")
	if _, err := UnwrapWithPassphrase(short, "pw"); err == nil {
		t.Fatal("expected truncated header error")
	}
}

func TestIsPassphraseEnvelope_Detection(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("x"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if !IsPassphraseEnvelope(wrapped) {
		t.Fatal("expected true for envelope")
	}
	if IsPassphraseEnvelope([]byte("plain text")) {
		t.Fatal("expected false for non-envelope")
	}
	if IsPassphraseEnvelope(nil) {
		t.Fatal("expected false for nil")
	}
	if IsPassphraseEnvelope([]byte("SPEC")) {
		t.Fatal("expected false for short prefix")
	}
}

func TestIsPassphraseEnvelopeFile_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wrapped.bin")
	wrapped, err := WrapWithPassphrase([]byte("payload"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if err := os.WriteFile(path, wrapped, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	ok, err := IsPassphraseEnvelopeFile(path)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !ok {
		t.Fatal("expected true")
	}
}

func TestIsPassphraseEnvelopeFile_PlainFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(path, []byte("hello world this is plain content"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	ok, err := IsPassphraseEnvelopeFile(path)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if ok {
		t.Fatal("expected false")
	}
}

func TestIsPassphraseEnvelopeFile_MissingFile(t *testing.T) {
	if _, err := IsPassphraseEnvelopeFile(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIsPassphraseEnvelopeFile_ShortFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "short.bin")
	if err := os.WriteFile(path, []byte("AB"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	ok, err := IsPassphraseEnvelopeFile(path)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if ok {
		t.Fatal("expected false for short file")
	}
}

func TestWrapWithPassphrase_NonceAndSaltUnique(t *testing.T) {
	plain := []byte("same input")
	const iterations = 8
	nonces := make(map[string]struct{}, iterations)
	salts := make(map[string]struct{}, iterations)
	ciphertexts := make(map[string]struct{}, iterations)
	for i := 0; i < iterations; i++ {
		wrapped, err := WrapWithPassphrase(plain, "pw")
		if err != nil {
			t.Fatalf("wrap: %v", err)
		}
		saltOffset := len(envelopeMagic) + 4
		nonceOffset := saltOffset + envelopeSaltLen
		salt := string(wrapped[saltOffset:nonceOffset])
		nonce := string(wrapped[nonceOffset : nonceOffset+envelopeNonceLen])
		if _, dup := nonces[nonce]; dup {
			t.Fatal("duplicate nonce across encryptions")
		}
		if _, dup := salts[salt]; dup {
			t.Fatal("duplicate salt across encryptions")
		}
		if _, dup := ciphertexts[string(wrapped)]; dup {
			t.Fatal("duplicate ciphertext across encryptions")
		}
		nonces[nonce] = struct{}{}
		salts[salt] = struct{}{}
		ciphertexts[string(wrapped)] = struct{}{}
	}
}

func TestWrapWithPassphrase_HeaderLayout(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("data"), "pw")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if !bytes.HasPrefix(wrapped, []byte(envelopeMagic)) {
		t.Fatal("magic prefix missing")
	}
	iters := binary.BigEndian.Uint32(wrapped[len(envelopeMagic) : len(envelopeMagic)+4])
	if iters != uint32(envelopeKDFIters) {
		t.Fatalf("kdf iters: got %d want %d", iters, envelopeKDFIters)
	}
	headerLen := len(envelopeMagic) + 4 + envelopeSaltLen + envelopeNonceLen
	if len(wrapped) < headerLen+1 {
		t.Fatal("envelope shorter than header")
	}
}

func TestPBKDF2_SamePassphraseSameSaltDerivesSameKey(t *testing.T) {
	wrapped, err := WrapWithPassphrase([]byte("payload"), "fixed-passphrase")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got1, err := UnwrapWithPassphrase(wrapped, "fixed-passphrase")
	if err != nil {
		t.Fatalf("unwrap1: %v", err)
	}
	got2, err := UnwrapWithPassphrase(wrapped, "fixed-passphrase")
	if err != nil {
		t.Fatalf("unwrap2: %v", err)
	}
	if !bytes.Equal(got1, got2) {
		t.Fatal("deterministic decrypt should match")
	}
}
