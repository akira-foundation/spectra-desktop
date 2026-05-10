package secrets

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	envelopeMagic     = "SPECTRA1"
	envelopeSaltLen   = 16
	envelopeNonceLen  = 12
	envelopeKeyLen    = 32
	envelopeKDFIters  = 100_000
)

func IsPassphraseEnvelope(data []byte) bool {
	return len(data) >= len(envelopeMagic) && bytes.HasPrefix(data, []byte(envelopeMagic))
}

func IsPassphraseEnvelopeFile(path string) (bool, error) {
	header := make([]byte, len(envelopeMagic))
	if err := readPrefix(path, header); err != nil {
		return false, err
	}
	return bytes.HasPrefix(header, []byte(envelopeMagic)), nil
}

func WrapWithPassphrase(plaintext []byte, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("envelope: passphrase required")
	}
	salt := make([]byte, envelopeSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	nonce := make([]byte, envelopeNonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(passphrase), salt, envelopeKDFIters, envelopeKeyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	out := bytes.NewBuffer(nil)
	out.WriteString(envelopeMagic)
	binary.Write(out, binary.BigEndian, uint32(envelopeKDFIters))
	out.Write(salt)
	out.Write(nonce)
	out.Write(ciphertext)
	return out.Bytes(), nil
}

func UnwrapWithPassphrase(envelope []byte, passphrase string) ([]byte, error) {
	if !IsPassphraseEnvelope(envelope) {
		return nil, errors.New("envelope: missing magic header")
	}
	headerLen := len(envelopeMagic) + 4 + envelopeSaltLen + envelopeNonceLen
	if len(envelope) < headerLen {
		return nil, errors.New("envelope: truncated header")
	}
	cursor := envelope[len(envelopeMagic):]
	iters := binary.BigEndian.Uint32(cursor[:4])
	cursor = cursor[4:]
	salt := cursor[:envelopeSaltLen]
	cursor = cursor[envelopeSaltLen:]
	nonce := cursor[:envelopeNonceLen]
	ciphertext := cursor[envelopeNonceLen:]

	key := pbkdf2.Key([]byte(passphrase), salt, int(iters), envelopeKeyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("envelope: invalid passphrase or corrupt data")
	}
	return plaintext, nil
}

func readPrefix(path string, dst []byte) error {
	f, err := openReadOnly(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.ReadFull(f, dst)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}
	return nil
}
