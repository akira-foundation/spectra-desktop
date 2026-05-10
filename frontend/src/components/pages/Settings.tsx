import { useState } from 'react'
import { Palette, Folder, Info, FileDown, FileUp, Package, Database, RotateCcw } from 'lucide-react'
import toast from 'react-hot-toast'
import { useTheme } from '@/hooks/useTheme'
import { useProjectStore } from '@/store/projectStore'
import { spectraIOService } from '@/services/spectraIOService'
import { Button } from '@/components/ui/button'
import { PassphraseDialog } from '@/components/settings/PassphraseDialog'
import { cn } from '@/lib/utils'

type PendingDialog =
  | { kind: 'export' }
  | { kind: 'import' }
  | { kind: 'backup' }
  | { kind: 'restore' }
  | null

export function Settings() {
  const theme = useTheme((s) => s.theme)
  const setTheme = useTheme((s) => s.setTheme)
  const projects = useProjectStore((s) => s.projects)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const loadFromStorage = useProjectStore((s) => s.loadFromStorage)
  const setActiveProject = useProjectStore((s) => s.setActiveProject)
  const activeProject = projects.find((p) => p.id === activeProjectId) ?? null
  const [busy, setBusy] = useState(false)
  const [dialog, setDialog] = useState<PendingDialog>(null)

  const exportArchive = async () => {
    if (!activeProjectId) {
      toast.error('No active project')
      return
    }
    setDialog({ kind: 'export' })
  }

  const runExport = async (passphrase: string) => {
    if (!activeProjectId) return
    const path = await spectraIOService.export({
      projectId: activeProjectId,
      passphrase: passphrase || undefined,
    })
    setDialog(null)
    if (path) toast.success(`Exported: ${path}`)
  }

  const importArchive = async () => {
    setBusy(true)
    try {
      const result = await spectraIOService.import()
      if (!result) return
      if (result.needsPassphrase) {
        setDialog({ kind: 'import' })
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
    setDialog(null)
    await loadFromStorage()
    await setActiveProject(result.newProjectId)
    toast.success(`Imported "${result.projectName}"`)
  }

  const backupDatabase = async () => {
    setDialog({ kind: 'backup' })
  }

  const runBackup = async (passphrase: string) => {
    const path = await spectraIOService.backupDatabase(passphrase || undefined)
    setDialog(null)
    if (path) toast.success(`Backup saved: ${path}`)
  }

  const restoreDatabase = async () => {
    if (!confirm('Restoring will replace your current database on next launch. Continue?')) return
    setBusy(true)
    try {
      const result = await spectraIOService.restoreDatabase()
      if (!result) return
      if (result.needsPassphrase) {
        setDialog({ kind: 'restore' })
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
    setDialog(null)
    toast.success('Backup staged. Relaunching…')
    setTimeout(() => void spectraIOService.relaunch(), 800)
  }

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-2xl mx-auto p-6 space-y-4">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Settings</h1>
          <p className="text-muted-foreground text-[12.5px] mt-1">
            Spectra preferences
          </p>
        </div>

        <Section title="Appearance" icon={Palette}>
          <div className="px-3.5 py-2.5 flex items-center justify-between text-[12.5px]">
            <span className="text-foreground/80">Theme</span>
            <div className="inline-flex items-center gap-1 rounded-md border border-border/40 p-0.5">
              {(['light', 'dark', 'system'] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setTheme(t)}
                  className={cn(
                    'px-2 py-0.5 text-[11px] rounded capitalize',
                    theme === t ? 'bg-accent text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>
        </Section>

        <Section title="Project archives" icon={Package}>
          <div className="px-3.5 py-3 space-y-3 text-[12.5px]">
            <p className="text-muted-foreground text-[11.5px]">
              Export the active project as a portable .spectra file (SQLite + JSON shards) or
              import one from a teammate. Secrets are stripped from exports.
            </p>
            <div className="flex items-center gap-2">
              <Button size="sm" onClick={exportArchive} disabled={busy || !activeProject}>
                <FileDown className="h-3.5 w-3.5" />
                Export {activeProject ? `"${activeProject.name}"` : 'project'}
              </Button>
              <Button size="sm" variant="outline" onClick={importArchive} disabled={busy}>
                <FileUp className="h-3.5 w-3.5" />
                Import .spectra
              </Button>
            </div>
          </div>
        </Section>

        <Section title="Database backup" icon={Database}>
          <div className="px-3.5 py-3 space-y-3 text-[12.5px]">
            <p className="text-muted-foreground text-[11.5px]">
              Snapshot the entire SQLite database (every project, account, mock, history)
              to a single file. Restore replaces the current database on next launch.
            </p>
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" onClick={backupDatabase} disabled={busy}>
                <Database className="h-3.5 w-3.5" />
                Backup database
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={restoreDatabase}
                disabled={busy}
                className="text-amber-600 hover:text-amber-700 dark:text-amber-400"
              >
                <RotateCcw className="h-3.5 w-3.5" />
                Restore database…
              </Button>
            </div>
          </div>
        </Section>

        <Section title="Projects" icon={Folder}>
          {projects.length === 0 ? (
            <p className="px-3.5 py-3 text-[11.5px] italic text-muted-foreground">
              No projects yet.
            </p>
          ) : (
            <ul className="divide-y divide-border/40">
              {projects.map((p) => (
                <li
                  key={p.id}
                  className="px-3.5 py-2 flex items-center justify-between gap-3 text-[12px]"
                >
                  <div className="min-w-0">
                    <p className="font-medium truncate capitalize">{p.name}</p>
                    <p className="text-[10.5px] text-muted-foreground font-mono truncate">
                      {p.path}
                    </p>
                  </div>
                  <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground shrink-0">
                    {p.framework}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Section>

        <Section title="About" icon={Info}>
          <div className="px-3.5 py-2.5 space-y-1 text-[12px]">
            <Row label="Version" value="0.1.0" />
            <Row label="Build" value="local" />
            <Row label="License" value="Beta · all features unlocked" />
          </div>
        </Section>
      </div>

      <PassphraseDialog
        open={dialog?.kind === 'export'}
        title="Encrypt project archive?"
        description="Leave empty for an unencrypted archive. With a passphrase, the file is wrapped in AES-256-GCM (PBKDF2)."
        confirmLabel="Export"
        requireConfirmation
        onClose={() => setDialog(null)}
        onSubmit={runExport}
      />
      <PassphraseDialog
        open={dialog?.kind === 'import'}
        title="Encrypted archive"
        description="This .spectra file is encrypted. Enter the passphrase to import it."
        confirmLabel="Import"
        onClose={() => setDialog(null)}
        onSubmit={finishImport}
      />
      <PassphraseDialog
        open={dialog?.kind === 'backup'}
        title="Encrypt backup?"
        description="Backups contain runtime history with real headers. A passphrase strongly recommended."
        confirmLabel="Backup"
        requireConfirmation
        onClose={() => setDialog(null)}
        onSubmit={runBackup}
      />
      <PassphraseDialog
        open={dialog?.kind === 'restore'}
        title="Encrypted backup"
        description="This backup is encrypted. Enter the passphrase to continue."
        confirmLabel="Restore"
        onClose={() => setDialog(null)}
        onSubmit={finishRestore}
      />
    </div>
  )
}

interface SectionProps {
  title: string
  icon: React.ComponentType<{ className?: string }>
  children: React.ReactNode
}

function Section({ title, icon: Icon, children }: SectionProps) {
  return (
    <section className="border border-border/60 rounded-lg overflow-hidden bg-card/40">
      <header className="px-3.5 py-2 bg-card/60 border-b border-border/50 flex items-center gap-2">
        <Icon className="w-3.5 h-3.5 text-muted-foreground" />
        <h2 className="font-semibold text-[11.5px] uppercase tracking-wider text-muted-foreground">
          {title}
        </h2>
      </header>
      {children}
    </section>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-foreground/80">{label}</span>
      <span className="text-muted-foreground font-mono">{value}</span>
    </div>
  )
}
