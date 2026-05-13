package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ProgressFunc receives bytes downloaded and total bytes (-1 if unknown).
type ProgressFunc func(downloaded, total int64)

// download fetches url into a fresh temp file and returns its path.
func download(ctx context.Context, url, suffix string, progress ProgressFunc) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: http %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "spectra-update-*"+suffix)
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 64*1024)
	for {
		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := tmp.Write(buf[:n]); wErr != nil {
				_ = os.Remove(tmp.Name())
				return "", wErr
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, total)
			}
		}
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			_ = os.Remove(tmp.Name())
			return "", rErr
		}
	}

	return tmp.Name(), nil
}

// stageAssets downloads the update zip and its signature into a temp dir.
func stageAssets(ctx context.Context, info *UpdateInfo, progress ProgressFunc) (zipPath, sigPath string, err error) {
	zipPath, err = download(ctx, info.URL, filepath.Ext(info.URL), progress)
	if err != nil {
		return "", "", err
	}
	// The signature lives next to the artifact as <artifact>.sig in DO Spaces.
	sigURL := info.URL + ".sig"
	sigPath, err = download(ctx, sigURL, ".sig", nil)
	if err != nil {
		_ = os.Remove(zipPath)
		return "", "", err
	}
	return zipPath, sigPath, nil
}
