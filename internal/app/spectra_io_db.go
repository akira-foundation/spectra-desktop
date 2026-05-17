package app

import (
	"fmt"
	"log"
	"spectra-desktop/internal/secrets"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DatabaseRestoreRequest struct {
	Passphrase string `json:"passphrase,omitempty"`
}

type DatabaseRestoreResult struct {
	Path            string `json:"path,omitempty"`
	NeedsPassphrase bool   `json:"needsPassphrase"`
}
type DatabaseBackupRequest struct {
	Passphrase string `json:"passphrase,omitempty"`
}

func (a *App) BackupDatabase(input DatabaseBackupRequest) (string, error) {
	if a.billingGate != nil {
		if err := a.billingGate.Require(a.ctx, "database_backup_restore"); err != nil {
			a.emitUpsell("database_backup_restore", err)
			return "", err
		}
	}
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
	if a.billingGate != nil {
		if err := a.billingGate.Require(a.ctx, "database_backup_restore"); err != nil {
			a.emitUpsell("database_backup_restore", err)
			return nil, err
		}
	}
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
