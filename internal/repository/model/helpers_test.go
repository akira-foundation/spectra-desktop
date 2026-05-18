package model_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
	"spectra-desktop/internal/storage"

	_ "modernc.org/sqlite"
)

func newDB(t *testing.T) *bun.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	dsn := fmt.Sprintf("file:%s?cache=shared&_journal=WAL&_busy_timeout=5000&_foreign_keys=on", path)
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	if err := storage.RunMigrationsOnDB(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertProject(t *testing.T, db *bun.DB) string {
	t.Helper()
	id := uuid.NewString()
	now := time.Now().UTC()
	p := model.Project{
		ID:        id,
		Name:      "p",
		Path:      "/" + id,
		Framework: "laravel",
		Status:    string(domain.ProjectStatusConnected),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&p).Exec(context.Background()); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return id
}
