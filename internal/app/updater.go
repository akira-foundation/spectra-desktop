package app

import (
	"fmt"
	"os"
	gort "runtime"
	"time"

	"spectra-desktop/internal/updater"
	"spectra-desktop/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CheckForUpdates queries the release manifest and returns update info when a
// newer version is available. Returns nil when up to date or in dev mode.
func (a *App) CheckForUpdates() (*updater.UpdateInfo, error) {
	return updater.Check(a.ctx, version.Version)
}

// InstallUpdate downloads, verifies, swaps the .app bundle, then relaunches.
// Emits "update:progress" events while downloading.
func (a *App) InstallUpdate() error {
	info, err := updater.Check(a.ctx, version.Version)
	if err != nil {
		return err
	}
	if info == nil {
		return nil
	}

	progress := func(downloaded, total int64) {
		runtime.EventsEmit(a.ctx, "update:progress", map[string]any{
			"downloaded": downloaded,
			"total":      total,
		})
	}

	if err := updater.Install(a.ctx, info, progress); err != nil {
		runtime.EventsEmit(a.ctx, "update:error", err.Error())
		return err
	}

	runtime.EventsEmit(a.ctx, "update:installed", info.Version)
	// Give the frontend a moment to receive the event, then exit so the
	// new bundle takes over.
	go func() {
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}

// AppVersion returns the running version, useful for the UI.
func (a *App) AppVersion() string {
	return version.Version
}

// AppPlatform returns a human-readable os/arch string, e.g. "darwin · arm64".
func (a *App) AppPlatform() string {
	return fmt.Sprintf("%s · %s", gort.GOOS, gort.GOARCH)
}

// AppChannel reports the build channel: "development" for unbuilt/dev binaries,
// otherwise "stable" (CI sets the version via ldflags only for release builds).
func (a *App) AppChannel() string {
	if version.IsDev() {
		return "development"
	}
	return "stable"
}
