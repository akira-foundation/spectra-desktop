package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/akira-io/desktopkit/paths"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "spectra-desktop"
	keyringUser    = "encryption-key-v1"
	envelopePrefix = "v1:"
)

var (
	once       sync.Once
	cachedKey  []byte
	cachedErr  error
	keyfileEnv = "SPECTRA_KEY_FILE"
)

type Vault struct {
	key []byte
}

func Default() (*Vault, error) {
	once.Do(func() {
		cachedKey, cachedErr = loadOrCreateKey()
	})
	if cachedErr != nil {
		return nil, cachedErr
	}
	return &Vault{key: cachedKey}, nil
}

func NewVault(key []byte) (*Vault, error) {
	if len(key) != 32 {
		return nil, errors.New("secrets: key must be 32 bytes")
	}
	dup := make([]byte, 32)
	copy(dup, key)
	return &Vault{key: dup}, nil
}

func MustDefault() *Vault {
	v, err := Default()
	if err != nil {
		panic(fmt.Errorf("secrets: init vault: %w", err))
	}
	return v
}

func (v *Vault) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return envelopePrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func (v *Vault) Decrypt(envelope string) (string, error) {
	if envelope == "" {
		return "", nil
	}
	if !strings.HasPrefix(envelope, envelopePrefix) {
		return envelope, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(envelope, envelopePrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("secrets: ciphertext too short")
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func loadOrCreateKey() ([]byte, error) {
	if existing, err := keyring.Get(keyringService, keyringUser); err == nil {
		decoded, decodeErr := base64.StdEncoding.DecodeString(existing)
		if decodeErr == nil && len(decoded) == 32 {
			return decoded, nil
		}
	}

	if path := keyFilePath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			decoded, decodeErr := base64.StdEncoding.DecodeString(string(data))
			if decodeErr == nil && len(decoded) == 32 {
				return decoded, nil
			}
		}
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(key)

	if err := keyring.Set(keyringService, keyringUser, encoded); err == nil {
		return key, nil
	}

	if path := keyFilePath(); path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err == nil {
			_ = os.WriteFile(path, []byte(encoded), 0o600)
		}
	}
	return key, nil
}

func keyFilePath() string {
	if override := os.Getenv(keyfileEnv); override != "" {
		return override
	}
	dir, err := paths.For("spectra").Config()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "vault.key")
}
