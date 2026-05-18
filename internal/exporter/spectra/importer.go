package spectra

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ImportResult struct {
	NewProjectID string   `json:"newProjectId"`
	ProjectName  string   `json:"projectName"`
	Tables       []string `json:"tables"`
}

func Import(ctx context.Context, archivePath, targetDBPath string) (*ImportResult, error) {
	tmpDir, err := os.MkdirTemp("", "spectra-import-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	manifest, dataPath, err := unzipArchive(archivePath, tmpDir)
	if err != nil {
		return nil, err
	}

	if manifest.FormatVersion > formatVersion {
		return nil, fmt.Errorf("archive format version %d is newer than supported %d", manifest.FormatVersion, formatVersion)
	}

	newProjectID := uuid.NewString()
	if err := remapProjectIDInExportDB(ctx, dataPath, manifest.ProjectID, newProjectID); err != nil {
		return nil, err
	}

	if err := mergeIntoMainDB(ctx, targetDBPath, dataPath, manifest.Tables); err != nil {
		return nil, err
	}

	return &ImportResult{
		NewProjectID: newProjectID,
		ProjectName:  manifest.ProjectName,
		Tables:       manifest.Tables,
	}, nil
}

func unzipArchive(archivePath, destDir string) (*Manifest, string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, "", fmt.Errorf("open archive: %w", err)
	}
	defer reader.Close()

	var manifest *Manifest
	dataPath := ""

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "..") || strings.Contains(file.Name, "..") {
			continue
		}
		switch file.Name {
		case manifestFile:
			payload, err := readZipFile(file)
			if err != nil {
				return nil, "", err
			}
			var parsed Manifest
			if err := json.Unmarshal(payload, &parsed); err != nil {
				return nil, "", fmt.Errorf("parse manifest: %w", err)
			}
			manifest = &parsed
		case dataFile:
			path := filepath.Join(destDir, dataFile)
			if err := extractZipFileTo(file, path); err != nil {
				return nil, "", err
			}
			dataPath = path
		}
	}

	if manifest == nil {
		return nil, "", errors.New("manifest.json missing in archive")
	}
	if dataPath == "" {
		return nil, "", errors.New("data.db missing in archive")
	}
	return manifest, dataPath, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func extractZipFileTo(file *zip.File, dest string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func remapProjectIDInExportDB(ctx context.Context, dbPath, oldID, newID string) error {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(0)", dbPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer rawDB.Close()
	rawDB.SetMaxOpenConns(1)

	rows, err := rawDB.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return err
	}
	tableNames := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		tableNames = append(tableNames, name)
	}
	rows.Close()

	for _, table := range tableNames {
		hasColumn, err := tableHasColumn(ctx, rawDB, table, "project_id")
		if err != nil {
			return err
		}
		if !hasColumn {
			continue
		}
		stmt := fmt.Sprintf("UPDATE %s SET project_id = ? WHERE project_id = ?", table)
		if _, err := rawDB.ExecContext(ctx, stmt, newID, oldID); err != nil {
			return fmt.Errorf("remap %s: %w", table, err)
		}
	}

	if _, err := rawDB.ExecContext(ctx, "UPDATE projects SET id = ? WHERE id = ?", newID, oldID); err != nil {
		return fmt.Errorf("remap projects: %w", err)
	}

	return nil
}

func tableHasColumn(ctx context.Context, db *sql.DB, table, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, nil
}

func mergeIntoMainDB(ctx context.Context, mainDBPath, importDBPath string, tables []string) error {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(0)", mainDBPath)
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer rawDB.Close()
	rawDB.SetMaxOpenConns(1)

	if _, err := rawDB.ExecContext(ctx, fmt.Sprintf("ATTACH DATABASE '%s' AS imp", importDBPath)); err != nil {
		return fmt.Errorf("attach import db: %w", err)
	}

	for _, table := range tables {
		stmt := fmt.Sprintf("INSERT INTO main.%s SELECT * FROM imp.%s", table, table)
		if _, err := rawDB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("merge %s: %w", table, err)
		}
	}

	if _, err := rawDB.ExecContext(ctx, "DETACH DATABASE imp"); err != nil {
		return fmt.Errorf("detach import: %w", err)
	}
	return nil
}
