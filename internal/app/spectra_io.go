package app

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"spectra-desktop/internal/exporter/spectra"
	"spectra-desktop/internal/secrets"
	"spectra-desktop/internal/storage"
)

type SpectraExportRequest struct {
	ProjectID      string `json:"projectId"`
	IncludeSecrets bool   `json:"includeSecrets,omitempty"`
	IncludeHistory bool   `json:"includeHistory,omitempty"`
	Passphrase     string `json:"passphrase,omitempty"`
}

type SpectraImportRequest struct {
	Passphrase string `json:"passphrase,omitempty"`
}

type SpectraImportResult struct {
	NewProjectID    string `json:"newProjectId"`
	ProjectName     string `json:"projectName"`
	NeedsPassphrase bool   `json:"needsPassphrase"`
}

type DatabaseRestoreRequest struct {
	Passphrase string `json:"passphrase,omitempty"`
}

type DatabaseRestoreResult struct {
	Path            string `json:"path,omitempty"`
	NeedsPassphrase bool   `json:"needsPassphrase"`
}

func (a *App) ExportProjectArchive(input SpectraExportRequest) (string, error) {
	if input.ProjectID == "" {
		return "", fmt.Errorf("projectId required")
	}
	project, err := a.projects.GetByID(a.ctx, input.ProjectID)
	if err != nil || project == nil {
		return "", fmt.Errorf("project not found")
	}
	defaultName := suggestExportFilename(project.Name)
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Export project archive",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spectra archive (*.spectra)", Pattern: "*.spectra"},
		},
	})
	if err != nil {
		return "", err
	}
	if savePath == "" {
		return "", nil
	}
	if !strings.HasSuffix(savePath, ".spectra") {
		savePath += ".spectra"
	}

	dbPath, err := storage.DefaultPath()
	if err != nil {
		return "", err
	}

	opts := spectra.ExportOptions{
		IncludeHistory: input.IncludeHistory,
		IncludeSecrets: input.IncludeSecrets,
	}
	log.Printf("spectra export: project=%q src=%q dst=%q encrypted=%t", input.ProjectID, dbPath, savePath, input.Passphrase != "")
	if err := spectra.Export(a.ctx, dbPath, input.ProjectID, savePath, opts); err != nil {
		log.Printf("spectra export failed: %v", err)
		return "", err
	}
	if input.Passphrase != "" {
		if err := secrets.WrapFileInPlace(savePath, input.Passphrase); err != nil {
			log.Printf("spectra export encrypt failed: %v", err)
			return "", err
		}
	}
	log.Printf("spectra export: ok %s", savePath)
	return savePath, nil
}

func (a *App) ImportProjectArchive() (*SpectraImportResult, error) {
	openPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import project archive",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spectra archive (*.spectra)", Pattern: "*.spectra"},
		},
	})
	if err != nil {
		return nil, err
	}
	if openPath == "" {
		return nil, nil
	}
	encrypted, err := secrets.IsPassphraseEnvelopeFile(openPath)
	if err != nil {
		return nil, err
	}
	if encrypted {
		stashPendingImport(openPath)
		return &SpectraImportResult{NeedsPassphrase: true}, nil
	}
	return runProjectImport(a, openPath)
}

func (a *App) FinishProjectImport(input SpectraImportRequest) (*SpectraImportResult, error) {
	openPath, ok := consumePendingImport()
	if !ok {
		return nil, fmt.Errorf("no pending import")
	}
	if input.Passphrase == "" {
		return nil, fmt.Errorf("passphrase required")
	}
	tmp := openPath + ".decrypted"
	defer removeFile(tmp)
	if err := secrets.UnwrapFileToPath(openPath, tmp, input.Passphrase); err != nil {
		return nil, err
	}
	return runProjectImport(a, tmp)
}

func runProjectImport(a *App, archivePath string) (*SpectraImportResult, error) {
	dbPath, err := storage.DefaultPath()
	if err != nil {
		return nil, err
	}
	log.Printf("spectra import: src=%q dst=%q", archivePath, dbPath)
	result, err := spectra.Import(a.ctx, archivePath, dbPath)
	if err != nil {
		log.Printf("spectra import failed: %v", err)
		return nil, err
	}
	log.Printf("spectra import: new project %s", result.NewProjectID)
	return &SpectraImportResult{
		NewProjectID: result.NewProjectID,
		ProjectName:  result.ProjectName,
	}, nil
}

type DatabaseBackupRequest struct {
	Passphrase string `json:"passphrase,omitempty"`
}

func (a *App) BackupDatabase(input DatabaseBackupRequest) (string, error) {
	defaultName := fmt.Sprintf("spectra-backup-%s.spectra-db", time.Now().UTC().Format("20060102-150405"))
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Backup Spectra database",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spectra database backup (*.spectra-db)", Pattern: "*.spectra-db"},
		},
	})
	if err != nil {
		return "", err
	}
	if savePath == "" {
		return "", nil
	}
	if !strings.HasSuffix(savePath, ".spectra-db") {
		savePath += ".spectra-db"
	}
	if err := a.storage.BackupTo(a.ctx, savePath); err != nil {
		log.Printf("backup database failed: %v", err)
		return "", err
	}
	if input.Passphrase != "" {
		if err := secrets.WrapFileInPlace(savePath, input.Passphrase); err != nil {
			log.Printf("backup encrypt failed: %v", err)
			return "", err
		}
	}
	log.Printf("backup database: ok %s encrypted=%t", savePath, input.Passphrase != "")
	return savePath, nil
}

func (a *App) RestoreDatabaseFromBackup() (*DatabaseRestoreResult, error) {
	openPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Restore Spectra database",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spectra database backup (*.spectra-db)", Pattern: "*.spectra-db"},
		},
	})
	if err != nil {
		return nil, err
	}
	if openPath == "" {
		return nil, nil
	}
	encrypted, err := secrets.IsPassphraseEnvelopeFile(openPath)
	if err != nil {
		return nil, err
	}
	if encrypted {
		stashPendingRestore(openPath)
		return &DatabaseRestoreResult{NeedsPassphrase: true}, nil
	}
	if err := a.storage.StagePendingRestore(openPath); err != nil {
		log.Printf("stage restore failed: %v", err)
		return nil, err
	}
	log.Printf("restore staged from %s; will apply on next launch", openPath)
	return &DatabaseRestoreResult{Path: openPath}, nil
}

func (a *App) FinishDatabaseRestore(input DatabaseRestoreRequest) (*DatabaseRestoreResult, error) {
	openPath, ok := consumePendingRestore()
	if !ok {
		return nil, fmt.Errorf("no pending restore")
	}
	if input.Passphrase == "" {
		return nil, fmt.Errorf("passphrase required")
	}
	tmp := openPath + ".decrypted"
	defer removeFile(tmp)
	if err := secrets.UnwrapFileToPath(openPath, tmp, input.Passphrase); err != nil {
		return nil, err
	}
	if err := a.storage.StagePendingRestore(tmp); err != nil {
		log.Printf("stage restore failed: %v", err)
		return nil, err
	}
	return &DatabaseRestoreResult{Path: openPath}, nil
}

func (a *App) RelaunchApplication() {
	runtime.Quit(a.ctx)
}

var pendingImportPath, pendingRestorePath string

func stashPendingImport(path string) { pendingImportPath = path }
func consumePendingImport() (string, bool) {
	if pendingImportPath == "" {
		return "", false
	}
	p := pendingImportPath
	pendingImportPath = ""
	return p, true
}
func stashPendingRestore(path string) { pendingRestorePath = path }
func consumePendingRestore() (string, bool) {
	if pendingRestorePath == "" {
		return "", false
	}
	p := pendingRestorePath
	pendingRestorePath = ""
	return p, true
}
func removeFile(path string) { _ = os.Remove(path) }

func suggestExportFilename(projectName string) string {
	clean := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '-' || r == '_' {
			return r
		}
		return '-'
	}, strings.ToLower(projectName))
	timestamp := time.Now().UTC().Format("20060102")
	return fmt.Sprintf("%s-%s.spectra", clean, timestamp)
}
