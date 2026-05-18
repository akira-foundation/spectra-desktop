package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akira-io/desktopkit/paths"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"spectra-desktop/internal/version"

	_ "modernc.org/sqlite"
)

const dbFile = "spectra.db"

func appFolder() string {
	if version.IsDev() {
		return "Spectra-dev"
	}
	return "Spectra"
}

type Storage struct {
	DB     *bun.DB
	sql    *sql.DB
	dbPath string
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Open(path string) error {
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return err
		}
		path = resolved
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?cache=shared&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	if err := sqldb.Ping(); err != nil {
		return fmt.Errorf("ping sqlite: %w", err)
	}
	sqldb.SetMaxOpenConns(1)

	s.sql = sqldb
	s.DB = bun.NewDB(sqldb, sqlitedialect.New())
	s.dbPath = path
	return nil
}

func (s *Storage) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	if s.sql != nil {
		return s.sql.Close()
	}
	return nil
}

func (s *Storage) Migrate(ctx context.Context) error {
	return runMigrations(ctx, s.DB)
}

func DefaultPath() (string, error) {
	cfg, err := paths.For(appFolder()).Config()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, dbFile), nil
}
