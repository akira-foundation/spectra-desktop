package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_StagePendingRestore_WritesPendingFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "live.db")
	s := openMigrated(t, srcPath)

	archive := filepath.Join(dir, "archive.db")
	if err := s.BackupTo(context.Background(), archive); err != nil {
		t.Fatalf("backup as archive: %v", err)
	}

	if err := s.StagePendingRestore(archive); err != nil {
		t.Fatalf("stage: %v", err)
	}

	pending := srcPath + pendingRestoreSuffix
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("expected pending file: %v", err)
	}
	if fileHash(t, pending) != fileHash(t, archive) {
		t.Fatal("pending file content differs from archive")
	}
}

func TestStorage_StagePendingRestore_RejectsInvalidArchive(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "live.db")
	s := openMigrated(t, srcPath)

	bogus := filepath.Join(dir, "bogus.bin")
	if err := os.WriteFile(bogus, []byte("not a sqlite file"), 0o644); err != nil {
		t.Fatalf("write bogus: %v", err)
	}

	if err := s.StagePendingRestore(bogus); err == nil {
		t.Fatal("expected validation failure for bogus archive")
	}

	if _, err := os.Stat(srcPath + pendingRestoreSuffix); !os.IsNotExist(err) {
		t.Fatalf("expected no pending file, stat err=%v", err)
	}
}

func TestApplyPendingRestoreIfAny_NoopWithoutPending(t *testing.T) {
	isolateHome(t)
	applied, err := ApplyPendingRestoreIfAny()
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if applied {
		t.Fatal("expected no apply when no pending file")
	}
}

func TestApplyPendingRestoreIfAny_ReplacesTargetAndClearsMarker(t *testing.T) {
	isolateHome(t)

	target, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}

	s := openMigrated(t, target)

	archive := filepath.Join(t.TempDir(), "snap.db")
	if err := s.BackupTo(context.Background(), archive); err != nil {
		t.Fatalf("backup: %v", err)
	}
	archiveHash := fileHash(t, archive)

	if err := s.StagePendingRestore(archive); err != nil {
		t.Fatalf("stage: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("close before apply: %v", err)
	}

	for _, suffix := range []string{walSuffix, shmSuffix} {
		path := target + suffix
		if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", path, err)
		}
	}

	applied, err := ApplyPendingRestoreIfAny()
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !applied {
		t.Fatal("expected applied=true")
	}

	if _, err := os.Stat(target + pendingRestoreSuffix); !os.IsNotExist(err) {
		t.Fatalf("expected pending cleared, stat err=%v", err)
	}
	for _, suffix := range []string{walSuffix, shmSuffix} {
		if _, err := os.Stat(target + suffix); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed, stat err=%v", target+suffix, err)
		}
	}

	if fileHash(t, target) != archiveHash {
		t.Fatal("restored target hash does not match archive")
	}

	if err := ValidateDatabaseFile(target); err != nil {
		t.Fatalf("restored db failed validation: %v", err)
	}
}
