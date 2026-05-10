package spectra

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"spectra-desktop/internal/storage"
)

const (
	manifestFile = "manifest.json"
	dataFile     = "data.db"
	jsonShardDir = "shards"
	formatVersion = 1
)

type ExportOptions struct {
	IncludeHistory bool
	IncludeSecrets bool
}

type Manifest struct {
	FormatVersion  int       `json:"formatVersion"`
	ExportedAt     time.Time `json:"exportedAt"`
	ProjectID      string    `json:"projectId"`
	ProjectName    string    `json:"projectName"`
	Framework      string    `json:"framework"`
	BaseURL        string    `json:"baseUrl"`
	IncludeHistory bool      `json:"includeHistory"`
	IncludeSecrets bool      `json:"includeSecrets"`
	Tables         []string  `json:"tables"`
}

type tableSpec struct {
	name        string
	whereClause string
	secretCols  []string
}

func tablesForExport(opts ExportOptions) []tableSpec {
	specs := []tableSpec{
		{name: "projects", whereClause: "id = ?"},
		{name: "endpoints", whereClause: "project_id = ?"},
		{name: "project_auth", whereClause: "project_id = ?", secretCols: []string{"token", "user_json", "cookies_json", "headers_json"}},
		{name: "project_accounts", whereClause: "project_id = ?", secretCols: []string{"token_enc", "password_enc", "api_key_enc", "refresh_token_enc", "totp_secret_enc", "oauth_config_json", "user_json", "cookies_json", "headers_json"}},
		{name: "environments", whereClause: "project_id = ?"},
		{name: "captured_values", whereClause: "project_id = ?"},
		{name: "collections", whereClause: "project_id = ?"},
		{name: "collection_items", whereClause: "collection_id IN (SELECT id FROM src.collections WHERE project_id = ?)"},
		{name: "collection_runs", whereClause: "collection_id IN (SELECT id FROM src.collections WHERE project_id = ?)"},
		{name: "endpoint_datasets", whereClause: "project_id = ?"},
		{name: "endpoint_snapshots", whereClause: "project_id = ?"},
		{name: "endpoint_tests", whereClause: "project_id = ?"},
		{name: "endpoint_captures", whereClause: "project_id = ?"},
		{name: "mock_overrides", whereClause: "project_id = ?"},
		{name: "scratch_requests", whereClause: "project_id = ?"},
	}
	if opts.IncludeHistory {
		specs = append(specs, tableSpec{name: "request_history", whereClause: "project_id = ?"})
	}
	return specs
}

func Export(ctx context.Context, sourceDBPath, projectID, outPath string, opts ExportOptions) error {
	tmpDir, err := os.MkdirTemp("", "spectra-export-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceCopy := filepath.Join(tmpDir, "source.db")
	if err := copyFileWithSidecars(sourceDBPath, sourceCopy); err != nil {
		return fmt.Errorf("snapshot source db: %w", err)
	}

	dataPath := filepath.Join(tmpDir, dataFile)
	if err := buildExportDB(ctx, sourceCopy, dataPath, projectID, opts); err != nil {
		return err
	}

	manifest, err := buildManifest(ctx, dataPath, projectID, opts)
	if err != nil {
		return err
	}

	shards, err := dumpJSONShards(ctx, dataPath, manifest.Tables)
	if err != nil {
		return err
	}

	return writeZipArchive(outPath, dataPath, manifest, shards)
}

func buildExportDB(ctx context.Context, sourceDBPath, targetPath, projectID string, opts ExportOptions) error {
	dsn := fmt.Sprintf("file:%s?_journal=WAL&_foreign_keys=off", targetPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open target db: %w", err)
	}
	defer rawDB.Close()
	rawDB.SetMaxOpenConns(1)

	target := bun.NewDB(rawDB, sqlitedialect.New())
	if err := storage.RunMigrationsOnDB(ctx, target); err != nil {
		return fmt.Errorf("migrate target db: %w", err)
	}

	if _, err := rawDB.ExecContext(ctx, fmt.Sprintf("ATTACH DATABASE '%s' AS src", sourceDBPath)); err != nil {
		return fmt.Errorf("attach source: %w", err)
	}

	for _, spec := range tablesForExport(opts) {
		if err := copyTableRows(ctx, rawDB, spec, projectID); err != nil {
			return fmt.Errorf("copy %s: %w", spec.name, err)
		}
		if !opts.IncludeSecrets && len(spec.secretCols) > 0 {
			if err := stripSecretColumns(ctx, rawDB, spec); err != nil {
				return fmt.Errorf("strip %s: %w", spec.name, err)
			}
		}
	}

	if opts.IncludeHistory {
		if err := scrubSensitiveHistoryColumns(ctx, rawDB); err != nil {
			return fmt.Errorf("scrub history: %w", err)
		}
	}

	if _, err := rawDB.ExecContext(ctx, "DETACH DATABASE src"); err != nil {
		return fmt.Errorf("detach source: %w", err)
	}
	return nil
}

func copyTableRows(ctx context.Context, db *sql.DB, spec tableSpec, projectID string) error {
	stmt := fmt.Sprintf("INSERT INTO main.%s SELECT * FROM src.%s WHERE %s", spec.name, spec.name, spec.whereClause)
	_, err := db.ExecContext(ctx, stmt, projectID)
	return err
}

func scrubSensitiveHistoryColumns(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id, request_headers, response_headers FROM main.request_history")
	if err != nil {
		return err
	}
	defer rows.Close()
	type entry struct {
		id              string
		requestHeaders  string
		responseHeaders string
	}
	var batch []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.requestHeaders, &e.responseHeaders); err != nil {
			return err
		}
		batch = append(batch, e)
	}
	rows.Close()
	for _, e := range batch {
		stmt := `UPDATE main.request_history SET request_headers = ?, response_headers = ? WHERE id = ?`
		if _, err := db.ExecContext(ctx, stmt,
			redactSensitiveHeadersJSON(e.requestHeaders),
			redactSensitiveHeadersJSON(e.responseHeaders),
			e.id); err != nil {
			return err
		}
	}
	return nil
}

func redactSensitiveHeadersJSON(headersJSON string) string {
	if headersJSON == "" {
		return headersJSON
	}
	var asMultiValue map[string][]string
	if err := json.Unmarshal([]byte(headersJSON), &asMultiValue); err == nil {
		for key := range asMultiValue {
			if isSensitiveHeader(key) {
				asMultiValue[key] = []string{"[redacted]"}
			}
		}
		raw, _ := json.Marshal(asMultiValue)
		return string(raw)
	}
	var asSingle map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &asSingle); err == nil {
		for key := range asSingle {
			if isSensitiveHeader(key) {
				asSingle[key] = "[redacted]"
			}
		}
		raw, _ := json.Marshal(asSingle)
		return string(raw)
	}
	return headersJSON
}

func isSensitiveHeader(name string) bool {
	switch lowerASCII(name) {
	case "authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token", "x-otp", "proxy-authorization":
		return true
	}
	return false
}

func lowerASCII(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}

func stripSecretColumns(ctx context.Context, db *sql.DB, spec tableSpec) error {
	for _, col := range spec.secretCols {
		stmt := fmt.Sprintf("UPDATE main.%s SET %s = ''", spec.name, col)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func buildManifest(ctx context.Context, dbPath, projectID string, opts ExportOptions) (*Manifest, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_journal=WAL", dbPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer rawDB.Close()

	row := rawDB.QueryRowContext(ctx, "SELECT name, framework, base_url FROM projects WHERE id = ?", projectID)
	var name, framework, baseURL string
	if err := row.Scan(&name, &framework, &baseURL); err != nil {
		return nil, fmt.Errorf("read project metadata: %w", err)
	}

	tables := []string{}
	for _, spec := range tablesForExport(opts) {
		tables = append(tables, spec.name)
	}

	return &Manifest{
		FormatVersion:  formatVersion,
		ExportedAt:     time.Now().UTC(),
		ProjectID:      projectID,
		ProjectName:    name,
		Framework:      framework,
		BaseURL:        baseURL,
		IncludeHistory: opts.IncludeHistory,
		IncludeSecrets: opts.IncludeSecrets,
		Tables:         tables,
	}, nil
}

type shardPayload struct {
	tableName string
	jsonBytes []byte
}

func dumpJSONShards(ctx context.Context, dbPath string, tables []string) ([]shardPayload, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_journal=WAL", dbPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer rawDB.Close()

	shards := make([]shardPayload, 0, len(tables))
	for _, table := range tables {
		rows, err := rawDB.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", table, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			return nil, err
		}
		records := []map[string]any{}
		for rows.Next() {
			values := make([]any, len(columns))
			scanTargets := make([]any, len(columns))
			for i := range values {
				scanTargets[i] = &values[i]
			}
			if err := rows.Scan(scanTargets...); err != nil {
				rows.Close()
				return nil, err
			}
			record := map[string]any{}
			for i, col := range columns {
				record[col] = normalizeScannedValue(values[i])
			}
			records = append(records, record)
		}
		rows.Close()
		raw, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return nil, err
		}
		shards = append(shards, shardPayload{tableName: table, jsonBytes: raw})
	}
	return shards, nil
}

func normalizeScannedValue(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return x
	}
}

func writeZipArchive(outPath, dataPath string, manifest *Manifest, shards []shardPayload) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := zip.NewWriter(out)
	defer writer.Close()

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := writeZipEntry(writer, manifestFile, manifestBytes); err != nil {
		return err
	}

	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	if err := writeZipEntry(writer, dataFile, dataBytes); err != nil {
		return err
	}

	for _, shard := range shards {
		path := filepath.Join(jsonShardDir, shard.tableName+".json")
		if err := writeZipEntry(writer, path, shard.jsonBytes); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(writer *zip.Writer, name string, body []byte) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, bytes.NewReader(body))
	return err
}

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
