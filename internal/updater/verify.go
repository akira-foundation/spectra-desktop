package updater

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"aead.dev/minisign"
)

// verify checks zipPath against sigPath using the embedded public key.
func verify(zipPath, sigPath string) error {
	if strings.TrimSpace(MinisignPublicKey) == "" || strings.HasPrefix(MinisignPublicKey, "REPLACE_") {
		return errors.New("updater public key not configured")
	}

	var pub minisign.PublicKey
	if err := pub.UnmarshalText([]byte(MinisignPublicKey)); err != nil {
		return fmt.Errorf("parse minisign public key: %w", err)
	}

	data, err := os.ReadFile(zipPath)
	if err != nil {
		return err
	}
	sig, err := os.ReadFile(sigPath)
	if err != nil {
		return err
	}

	if !minisign.Verify(pub, data, sig) {
		return errors.New("update signature invalid")
	}
	return nil
}
