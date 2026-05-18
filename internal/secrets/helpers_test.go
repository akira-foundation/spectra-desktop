package secrets

import (
	"crypto/rand"
	"io"
	"testing"
)

func newTestVault(t *testing.T) *Vault {
	t.Helper()
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("read key: %v", err)
	}
	return &Vault{key: key}
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return buf
}
