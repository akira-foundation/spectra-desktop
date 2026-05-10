import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Package, FileDown, FileUp } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useProjectStore } from '@/store/projectStore'
import { useUIStore } from '@/store/uiStore'
import { spectraIOService } from '@/services/spectraIOService'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { PassphraseDialog } from './PassphraseDialog'

type Stage = { kind: 'export' } | { kind: 'import' } | null

export function ArchivesPanel() {
  const projects = useProjectStore((s) => s.projects)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const loadFromStorage = useProjectStore((s) => s.loadFromStorage)
  const setActiveProject = useProjectStore((s) => s.setActiveProject)
  const activeProject = projects.find((p) => p.id === activeProjectId) ?? null
  const [stage, setStage] = useState<Stage>(null)
  const [busy, setBusy] = useState(false)
  const pendingAction = useUIStore((s) => s.pendingArchiveAction)
  const clearPending = useUIStore((s) => s.setPendingArchiveAction)

  const startExport = () => {
    if (!activeProjectId) return toast.error('No active project')
    setStage({ kind: 'export' })
  }

  const runExport = async (passphrase: string) => {
    if (!activeProjectId) return
    const path = await spectraIOService.export({
      projectId: activeProjectId,
      passphrase: passphrase || undefined,
    })
    setStage(null)
    if (path) toast.success(`Exported: ${path}`)
  }

  const startImport = async () => {
    setBusy(true)
    try {
      const result = await spectraIOService.import()
      if (!result) return
      if (result.needsPassphrase) {
        setStage({ kind: 'import' })
        return
      }
      await loadFromStorage()
      await setActiveProject(result.newProjectId)
      toast.success(`Imported "${result.projectName}"`)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const finishImport = async (passphrase: string) => {
    const result = await spectraIOService.finishImport(passphrase)
    setStage(null)
    await loadFromStorage()
    await setActiveProject(result.newProjectId)
    toast.success(`Imported "${result.projectName}"`)
  }

  useEffect(() => {
    if (pendingAction === 'export') {
      clearPending(null)
      startExport()
    } else if (pendingAction === 'import') {
      clearPending(null)
      void startImport()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pendingAction])

  return (
    <div>
      <SettingsHeader
        icon={Package}
        title="Project archives"
        description="Share a single project as a portable .spectra file. Secrets are stripped."
      />

      <SettingsCard>
        <SettingsRow
          label={`Export ${activeProject ? `"${activeProject.name}"` : 'project'}`}
          description="Bundles endpoints, accounts (no secrets), collections, env, mocks and tests."
          control={
            <Button size="sm" onClick={startExport} disabled={busy || !activeProject}>
              <FileDown className="h-3.5 w-3.5" />
              Export
            </Button>
          }
        />
        <SettingsRow
          label="Import .spectra"
          description="Creates a new project. Encrypted archives prompt for the passphrase."
          control={
            <Button size="sm" variant="outline" onClick={startImport} disabled={busy}>
              <FileUp className="h-3.5 w-3.5" />
              Import
            </Button>
          }
        />
      </SettingsCard>

      <PassphraseDialog
        open={stage?.kind === 'export'}
        title="Encrypt project archive?"
        description="Leave empty for an unencrypted archive. With a passphrase, the file is wrapped in AES-256-GCM (PBKDF2 100k)."
        confirmLabel="Export"
        requireConfirmation
        onClose={() => setStage(null)}
        onSubmit={runExport}
      />
      <PassphraseDialog
        open={stage?.kind === 'import'}
        title="Encrypted archive"
        description="This .spectra file is encrypted. Enter the passphrase to import it."
        confirmLabel="Import"
        onClose={() => setStage(null)}
        onSubmit={finishImport}
      />
    </div>
  )
}
