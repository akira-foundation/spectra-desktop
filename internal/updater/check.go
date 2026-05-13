package updater

import (
	"context"
	"errors"
	"fmt"

	"spectra-desktop/internal/version"
)

// UpdateInfo describes an available update to surface to the UI.
type UpdateInfo struct {
	Version        string `json:"version"`
	CurrentVersion string `json:"currentVersion"`
	Notes          string `json:"notes"`
	URL            string `json:"-"`
}

// Check returns a non-nil UpdateInfo when a newer version is available for
// the current platform. Returns nil when up to date.
func Check(ctx context.Context, currentVersion string) (*UpdateInfo, error) {
	if version.IsDev() {
		return nil, nil
	}
	if currentVersion == "" {
		return nil, nil
	}

	m, err := fetchManifest(ctx)
	if err != nil {
		return nil, err
	}

	key := PlatformKey()
	plat, ok := m.Platforms[key]
	if !ok {
		return nil, fmt.Errorf("no artifact for platform %s", key)
	}
	if plat.URL == "" {
		return nil, errors.New("manifest platform url empty")
	}

	if compareSemver(m.Version, currentVersion) <= 0 {
		return nil, nil
	}

	return &UpdateInfo{
		Version:        m.Version,
		CurrentVersion: currentVersion,
		Notes:          m.Notes,
		URL:            plat.URL,
	}, nil
}
