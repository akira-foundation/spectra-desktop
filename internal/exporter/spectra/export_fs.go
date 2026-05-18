package spectra

import (
	"io"
	"os"
	"path/filepath"
)

func copyFileWithSidecars(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		sidecar := src + suffix
		if _, err := os.Stat(sidecar); err == nil {
			_ = copyFile(sidecar, dst+suffix)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
