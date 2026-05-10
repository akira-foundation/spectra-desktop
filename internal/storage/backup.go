package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	pendingRestoreSuffix = ".pending"
	walSuffix            = "-wal"
	shmSuffix            = "-shm"
)

func (s *Storage) BackupTo(ctx context.Context, destPath string) error {
	if s.sql == nil {
		return errors.New("storage: db not open")
	}
	if _, err := s.sql.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("checkpoint wal: %w", err)
	}
	srcPath, err := s.path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return copySingleFile(srcPath, destPath)
}

func (s *Storage) StagePendingRestore(srcArchivePath string) error {
	if err := ValidateDatabaseFile(srcArchivePath); err != nil {
		return err
	}
	target, err := s.path()
	if err != nil {
		return err
	}
	pending := target + pendingRestoreSuffix
	if err := os.MkdirAll(filepath.Dir(pending), 0o755); err != nil {
		return err
	}
	return copySingleFile(srcArchivePath, pending)
}

func ValidateDatabaseFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?mode=ro", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open candidate: %w", err)
	}
	defer db.Close()

	row := db.QueryRow("PRAGMA integrity_check")
	var result string
	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("integrity check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("database failed integrity check: %s", result)
	}

	row = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='projects'")
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("inspect schema: %w", err)
	}
	if count == 0 {
		return errors.New("file is a SQLite database but does not contain Spectra schema")
	}
	return nil
}

func ApplyPendingRestoreIfAny() (bool, error) {
	target, err := DefaultPath()
	if err != nil {
		return false, err
	}
	pending := target + pendingRestoreSuffix
	if _, err := os.Stat(pending); err != nil {
		return false, nil
	}
	for _, suffix := range []string{walSuffix, shmSuffix} {
		_ = os.Remove(target + suffix)
	}
	if err := os.Rename(pending, target); err != nil {
		return false, fmt.Errorf("apply pending restore: %w", err)
	}
	return true, nil
}

func (s *Storage) path() (string, error) {
	if s.dbPath != "" {
		return s.dbPath, nil
	}
	return DefaultPath()
}

func copySingleFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
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
