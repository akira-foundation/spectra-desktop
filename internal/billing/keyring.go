package billing

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"crypto/ed25519"

	billingsdk "github.com/akira-io/billing-sdk-go"
)

type Keyring struct {
	keys map[string]ed25519.PublicKey
}

func NewKeyringFromEnv() (*Keyring, error) {
	raw := strings.TrimSpace(LicensePubKey)
	if raw == "" {
		return nil, errors.New("billing: license public key not configured")
	}

	keys := make(map[string]ed25519.PublicKey)
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		keyID, pkB64 := splitKeyEntry(entry)
		bytes, err := base64.StdEncoding.DecodeString(pkB64)
		if err != nil {
			return nil, fmt.Errorf("billing: decode pubkey %q: %w", keyID, err)
		}
		if len(bytes) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("billing: pubkey %q must be %d bytes", keyID, ed25519.PublicKeySize)
		}
		keys[keyID] = ed25519.PublicKey(bytes)
	}
	if len(keys) == 0 {
		return nil, errors.New("billing: keyring parsed to zero keys")
	}
	return &Keyring{keys: keys}, nil
}

func (k *Keyring) Verify(signed billingsdk.SignedLicense) (*billingsdk.LicenseSnapshotPayload, error) {
	pub, ok := k.keys[signed.KeyID]
	if !ok {
		if signed.KeyID == "" {
			if fallback, exists := k.keys["default"]; exists {
				pub = fallback
				ok = true
			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("billing: unknown signing key_id %q", signed.KeyID)
	}

	pkB64 := base64.StdEncoding.EncodeToString(pub)
	verified, err := billingsdk.VerifyLicense(signed, pkB64)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, errors.New("billing: license signature invalid")
	}

	decoded, err := billingsdk.DecodeLicense(signed)
	if err != nil {
		return nil, err
	}
	return &decoded.Payload, nil
}

func splitKeyEntry(entry string) (string, string) {
	if idx := strings.Index(entry, ":"); idx >= 0 {
		return strings.TrimSpace(entry[:idx]), strings.TrimSpace(entry[idx+1:])
	}
	return "default", strings.TrimSpace(entry)
}
