import {
  ExportProjectArchive,
  ImportProjectArchive,
  FinishProjectImport,
  BackupDatabase,
  RestoreDatabaseFromBackup,
  FinishDatabaseRestore,
  RelaunchApplication,
} from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export interface ExportOptions {
  projectId: string
  includeSecrets?: boolean
  includeHistory?: boolean
  passphrase?: string
}

export interface ImportResult {
  newProjectId: string
  projectName: string
  needsPassphrase: boolean
}

export interface RestoreResult {
  path?: string
  needsPassphrase: boolean
}

export const spectraIOService = {
  async export(opts: ExportOptions): Promise<string | null> {
    const path = await ExportProjectArchive({
      projectId: opts.projectId,
      includeSecrets: !!opts.includeSecrets,
      includeHistory: !!opts.includeHistory,
      passphrase: opts.passphrase ?? '',
    } as unknown as app.SpectraExportRequest)
    return path || null
  },
  async import(): Promise<ImportResult | null> {
    const result = await ImportProjectArchive()
    return (result as unknown as ImportResult) ?? null
  },
  async finishImport(passphrase: string): Promise<ImportResult> {
    const result = await FinishProjectImport({ passphrase } as unknown as app.SpectraImportRequest)
    return result as unknown as ImportResult
  },
  async backupDatabase(passphrase?: string): Promise<string | null> {
    const path = await BackupDatabase({ passphrase: passphrase ?? '' } as unknown as app.DatabaseBackupRequest)
    return path || null
  },
  async restoreDatabase(): Promise<RestoreResult | null> {
    const result = await RestoreDatabaseFromBackup()
    return (result as unknown as RestoreResult) ?? null
  },
  async finishRestore(passphrase: string): Promise<RestoreResult> {
    const result = await FinishDatabaseRestore({ passphrase } as unknown as app.DatabaseRestoreRequest)
    return result as unknown as RestoreResult
  },
  async relaunch(): Promise<void> {
    await RelaunchApplication()
  },
}
