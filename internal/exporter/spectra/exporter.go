package spectra

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const (
	manifestFile  = "manifest.json"
	dataFile      = "data.db"
	jsonShardDir  = "shards"
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
