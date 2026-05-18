package spectra

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"

	"spectra-desktop/internal/storage"
)

type seedFixture struct {
	ProjectID  string
	EndpointID string
	EnvID      string
	HistoryID  string
}

func newMigratedDB(t *testing.T) (string, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "src.db")
	dsn := fmt.Sprintf("file:%s?cache=shared&_journal=WAL&_busy_timeout=5000&_foreign_keys=on", path)
	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	raw.SetMaxOpenConns(1)
	if err := raw.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	db := bun.NewDB(raw, sqlitedialect.New())
	if err := storage.RunMigrationsOnDB(context.Background(), db); err != nil {
		_ = raw.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	return path, raw
}

func seedProject(t *testing.T, db *sql.DB) seedFixture {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	fx := seedFixture{
		ProjectID:  uuid.NewString(),
		EndpointID: uuid.NewString(),
		EnvID:      uuid.NewString(),
		HistoryID:  uuid.NewString(),
	}
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, name, path, framework, framework_version, status, created_at, updated_at, base_url, active_environment_id)
		 VALUES (?, ?, ?, ?, '', 'connected', ?, ?, ?, '')`,
		fx.ProjectID, "demo", filepath.Join(t.TempDir(), "proj"), "laravel", now, now, "http://demo.test",
	); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO endpoints (id, project_id, method, path, scanned_at, created_at, updated_at)
		 VALUES (?, ?, 'GET', '/users', ?, ?, ?)`,
		fx.EndpointID, fx.ProjectID, now, now, now,
	); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO environments (id, project_id, name, vars_json, sort_order, created_at, updated_at)
		 VALUES (?, ?, 'local', '{"FOO":"bar"}', 0, ?, ?)`,
		fx.EnvID, fx.ProjectID, now, now,
	); err != nil {
		t.Fatalf("seed env: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO project_auth (project_id, scheme, token, user_json, cookies_json, headers_json, captured_at, updated_at)
		 VALUES (?, 'bearer', 'super-secret-token', '{"id":1}', '[]', '{}', ?, ?)`,
		fx.ProjectID, now, now,
	); err != nil {
		t.Fatalf("seed auth: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO request_history (id, project_id, endpoint_id, method, url, request_headers, response_headers, created_at)
		 VALUES (?, ?, ?, 'GET', 'http://demo.test/users', ?, ?, ?)`,
		fx.HistoryID, fx.ProjectID, fx.EndpointID,
		`{"Authorization":"Bearer abc","X-Trace":"1"}`,
		`{"Set-Cookie":"sid=xyz","X-Other":"v"}`,
		now,
	); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	return fx
}

func openArchive(t *testing.T, path string) (*Manifest, map[string][]byte) {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer zr.Close()

	entries := map[string][]byte{}
	var manifest *Manifest
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open entry %s: %v", f.Name, err)
		}
		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read entry %s: %v", f.Name, err)
		}
		entries[f.Name] = body
		if f.Name == manifestFile {
			var m Manifest
			if err := json.Unmarshal(body, &m); err != nil {
				t.Fatalf("decode manifest: %v", err)
			}
			manifest = &m
		}
	}
	if manifest == nil {
		t.Fatal("archive missing manifest")
	}
	return manifest, entries
}

func queryString(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var v sql.NullString
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return v.String
}

func queryInt(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var v int
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return v
}

func openRawSQLite(t *testing.T, path string) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?_journal=WAL", path)
	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = raw.Close() })
	return raw
}

func writeFile(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
