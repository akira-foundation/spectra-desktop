package updater

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestURL_DefaultLooksValid(t *testing.T) {
	if !strings.HasPrefix(ManifestURL, "https://") {
		t.Fatalf("ManifestURL must be https; got %q", ManifestURL)
	}
	if !strings.HasSuffix(ManifestURL, ".json") {
		t.Fatalf("ManifestURL must point at a .json manifest; got %q", ManifestURL)
	}
}

func TestMinisignPublicKey_DefaultNonEmpty(t *testing.T) {
	if strings.TrimSpace(MinisignPublicKey) == "" {
		t.Fatal("MinisignPublicKey must not be empty in default build")
	}
	if strings.HasPrefix(MinisignPublicKey, "REPLACE_") {
		t.Fatalf("MinisignPublicKey is still a placeholder: %q", MinisignPublicKey)
	}
}

func TestConfigVars_RoundtripOverride(t *testing.T) {
	tmp := t.TempDir()
	fakePath := filepath.Join(tmp, "latest.json")

	prevURL := ManifestURL
	prevKey := MinisignPublicKey
	t.Cleanup(func() {
		ManifestURL = prevURL
		MinisignPublicKey = prevKey
	})

	ManifestURL = "file://" + fakePath
	MinisignPublicKey = "test-override"

	if ManifestURL != "file://"+fakePath {
		t.Fatalf("ManifestURL override not applied: %q", ManifestURL)
	}
	if MinisignPublicKey != "test-override" {
		t.Fatalf("MinisignPublicKey override not applied: %q", MinisignPublicKey)
	}
}
