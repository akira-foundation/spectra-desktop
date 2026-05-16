package billing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	machineIDFile = "machine_id"
	bundleID      = "foundation.akira.spectra"
)

type MachineIdentity struct {
	ID string
}

func GetOrCreateMachineIdentity(appConfigDir string) (*MachineIdentity, error) {
	path := identityPath(appConfigDir)
	if existing, err := loadIdentity(path); err == nil && existing != nil {
		return existing, nil
	}
	identity, err := createIdentity()
	if err != nil {
		return nil, err
	}
	if err := storeIdentity(path, identity); err != nil {
		return nil, err
	}
	return identity, nil
}

func identityPath(appConfigDir string) string {
	return filepath.Join(appConfigDir, machineIDFile)
}

func loadIdentity(path string) (*MachineIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	plain, err := decryptIdentity(data)
	if err != nil {
		return nil, err
	}
	id := string(plain)
	if id == "" {
		return nil, errors.New("machine_id: empty after decrypt")
	}
	return &MachineIdentity{ID: id}, nil
}

func createIdentity() (*MachineIdentity, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, fmt.Errorf("machine_id: random: %w", err)
	}
	return &MachineIdentity{ID: "sp_" + hex.EncodeToString(buf)}, nil
}

func storeIdentity(path string, identity *MachineIdentity) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	enc, err := encryptIdentity([]byte(identity.ID))
	if err != nil {
		return err
	}
	return os.WriteFile(path, enc, 0o600)
}

func deriveMachineKey() [32]byte {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "unknown"
	}
	h := sha256.New()
	h.Write([]byte(bundleID))
	h.Write([]byte{':'})
	h.Write([]byte(user))
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key
}

func encryptIdentity(plain []byte) ([]byte, error) {
	key := deriveMachineKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)
	return sealed, nil
}

func decryptIdentity(data []byte) ([]byte, error) {
	key := deriveMachineKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("machine_id: ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
