package spectra

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func buildManifest(ctx context.Context, dbPath, projectID string, opts ExportOptions) (*Manifest, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=journal_mode(WAL)", dbPath)
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
