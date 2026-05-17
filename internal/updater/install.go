package updater

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akira-io/desktopkit/osinfo"
	"github.com/akira-io/desktopkit/shell"
)

func Install(ctx context.Context, info *UpdateInfo, progress ProgressFunc) error {
	platform := osinfo.Current()
	if !platform.IsDarwin() {
		return fmt.Errorf("auto-update not supported on %s", platform.String())
	}
	if info == nil {
		return errors.New("nil update info")
	}

	zipPath, sigPath, err := stageAssets(ctx, info, progress)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)
	defer os.Remove(sigPath)

	if err := verify(zipPath, sigPath); err != nil {
		return err
	}

	currentApp, err := currentAppBundlePath()
	if err != nil {
		return err
	}

	stagingDir, err := os.MkdirTemp("", "spectra-update-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stagingDir)

	if err := unzip(zipPath, stagingDir); err != nil {
		return fmt.Errorf("unzip update: %w", err)
	}

	newApp, err := findAppBundle(stagingDir)
	if err != nil {
		return err
	}

	if err := swapBundle(currentApp, newApp); err != nil {
		return err
	}

	return relaunch(currentApp)
}

// currentAppBundlePath walks up from the running executable to the .app bundle root.
func currentAppBundlePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	for {
		if strings.HasSuffix(dir, ".app") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("running executable is not inside a .app bundle")
		}
		dir = parent
	}
}

func findAppBundle(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".app") {
			return filepath.Join(root, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no .app bundle found in %s", root)
}

// swapBundle moves the running .app aside and installs the new one in its place.
// Leaves the old bundle as <name>.app.old next to it for rollback.
func swapBundle(currentApp, newApp string) error {
	backup := currentApp + ".old"
	_ = os.RemoveAll(backup)

	if err := os.Rename(currentApp, backup); err != nil {
		return fmt.Errorf("backup current app: %w", err)
	}

	if err := os.Rename(newApp, currentApp); err != nil {
		// Best-effort rollback
		_ = os.Rename(backup, currentApp)
		return fmt.Errorf("install new app: %w", err)
	}

	return nil
}

func relaunch(appPath string) error {
	resolved, err := shell.NewCandidates().
		WithName("open").
		WithCandidate("/usr/bin/open").
		Resolve()
	if err != nil {
		return fmt.Errorf("relaunch: locate open: %w", err)
	}
	cmd := exec.Command(resolved.AbsolutePath(), "-n", appPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("relaunch: %w", err)
	}
	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry escapes dest: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		// Preserve symlinks (frameworks rely on them).
		if f.Mode()&os.ModeSymlink != 0 {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}
			if err := os.Symlink(string(data), target); err != nil {
				return err
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return err
		}
		rc.Close()
		out.Close()
	}
	return nil
}
