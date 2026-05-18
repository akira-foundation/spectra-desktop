package spectra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"spectra-desktop/internal/storage"
)

func buildExportDB(ctx context.Context, sourceDBPath, targetPath, projectID string, opts ExportOptions) error {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(0)", targetPath)
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

func stripSecretColumns(ctx context.Context, db *sql.DB, spec tableSpec) error {
	for _, col := range spec.secretCols {
		stmt := fmt.Sprintf("UPDATE main.%s SET %s = ''", spec.name, col)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
