import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Database, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useUIStore } from '@/store/uiStore'
import { spectraIOService } from '@/services/spectraIOService'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { PassphraseDialog } from './PassphraseDialog'

type Stage = { kind: 'backup' } | { kind: 'restore' } | null

export function DatabasePanel() {
  const [stage, setStage] = useState<Stage>(null)
  const [busy, setBusy] = useState(false)
  const pendingAction = useUIStore((s) => s.pendingArchiveAction)
  const clearPending = useUIStore((s) => s.setPendingArchiveAction)

  const startBackup = () => setStage({ kind: 'backup' })

  const runBackup = async (passphrase: string) => {
    const path = await spectraIOService.backupDatabase(passphrase || undefined)
    setStage(null)
    if (path) toast.success(`Backup saved: ${path}`)
  }

  const startRestore = async () => {
    if (!confirm('Restoring will replace your current database on next launch. Continue?')) return
    setBusy(true)
    try {
      const result = await spectraIOService.restoreDatabase()
      if (!result) return
      if (result.needsPassphrase) {
        setStage({ kind: 'restore' })
        return
      }
      toast.success('Backup staged. Relaunching…')
      setTimeout(() => void spectraIOService.relaunch(), 800)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const finishRestore = async (passphrase: string) => {
    await spectraIOService.finishRestore(passphrase)
    setStage(null)
    toast.success('Backup staged. Relaunching…')
    setTimeout(() => void spectraIOService.relaunch(), 800)
  }

  useEffect(() => {
    if (pendingAction === 'backup') {
      clearPending(null)
      startBackup()
    } else if (pendingAction === 'restore') {
      clearPending(null)
      void startRestore()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pendingAction])

  return (
    <div>
      <SettingsHeader
        icon={Database}
        title="Database backup"
        description="Snapshot the entire SQLite database — every project, account, history, and mock."
      />

      <SettingsCard>
        <SettingsRow
          label="Backup database"
          description="Encrypt with a passphrase if the file leaves your machine — backups contain runtime history with real headers."
          control={
            <Button size="sm" onClick={startBackup} disabled={busy}>
              <Database className="h-3.5 w-3.5" />
              Backup
            </Button>
          }
        />
        <SettingsRow
          label="Restore from backup"
          description="Replaces the current database on next launch. Spectra relaunches automatically."
          control={
            <Button
              size="sm"
              variant="outline"
              onClick={startRestore}
              disabled={busy}
              className="text-amber-600 hover:text-amber-700 dark:text-amber-400"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Restore
            </Button>
          }
        />
      </SettingsCard>

      <PassphraseDialog
        open={stage?.kind === 'backup'}
        title="Encrypt backup?"
        description="Strongly recommended. Without a passphrase, the file is the raw SQLite database."
        confirmLabel="Backup"
        requireConfirmation
        onClose={() => setStage(null)}
        onSubmit={runBackup}
      />
      <PassphraseDialog
        open={stage?.kind === 'restore'}
        title="Encrypted backup"
        description="This backup is encrypted. Enter the passphrase to continue."
        confirmLabel="Restore"
        onClose={() => setStage(null)}
        onSubmit={finishRestore}
      />
    </div>
  )
}
