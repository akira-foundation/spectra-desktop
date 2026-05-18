package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStorage_BackupTo_ProducesReadableCopy(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "src.db")
	s := openMigrated(t, srcPath)

	dstPath := filepath.Join(dir, "backups", "snap.db")
	if err := s.BackupTo(context.Background(), dstPath); err != nil {
		t.Fatalf("BackupTo: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file is empty")
	}

	if err := ValidateDatabaseFile(dstPath); err != nil {
		t.Fatalf("backup failed Spectra validation: %v", err)
	}

	srcHash := fileHash(t, srcPath)
	dstHash := fileHash(t, dstPath)
	if srcHash != dstHash {
		t.Fatalf("backup hash %s differs from source %s", dstHash, srcHash)
	}
}

func TestStorage_BackupTo_FailsWhenNotOpen(t *testing.T) {
	s := New()
	err := s.BackupTo(context.Background(), filepath.Join(t.TempDir(), "x.db"))
	if err == nil {
		t.Fatal("expected error backing up closed storage")
	}
}

func TestValidateDatabaseFile_RejectsNonExistent(t *testing.T) {
	err := ValidateDatabaseFile(filepath.Join(t.TempDir(), "nope.db"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateDatabaseFile_RejectsForeignSQLite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foreign.db")

	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("open foreign: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE other (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create other: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close foreign: %v", err)
	}

	err = ValidateDatabaseFile(path)
	if err == nil {
		t.Fatal("expected schema rejection for foreign sqlite db")
	}
	if !strings.Contains(err.Error(), "Spectra schema") {
		t.Fatalf("expected Spectra schema error, got %v", err)
	}
}
