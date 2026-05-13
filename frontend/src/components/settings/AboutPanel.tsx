import { Info, Download, RefreshCw, CheckCircle2, AlertCircle } from 'lucide-react'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { Button } from '@/components/ui/button'
import { useUpdater } from '@/hooks/useUpdater'

export function AboutPanel() {
  const { update, currentVersion, checking, installing, progress, error, check, install } = useUpdater()

  const hasUpdate = update !== null

  return (
    <div>
      <SettingsHeader icon={Info} title="About" description="What is running locally." />

      <SettingsCard>
        <SettingsRow label="Version" control={<Mono value={currentVersion || '—'} />} />
        <SettingsRow label="Build" control={<Mono value="local" />} />
        <SettingsRow label="License" control={<Mono value="Beta · all features unlocked" />} />
        <SettingsRow
          label="Updates"
          control={<UpdateControl
            hasUpdate={hasUpdate}
            checking={checking}
            installing={installing}
            progress={progress}
            update={update}
            error={error}
            onCheck={check}
            onInstall={install}
          />}
        />
      </SettingsCard>
    </div>
  )
}

interface UpdateControlProps {
  hasUpdate: boolean
  checking: boolean
  installing: boolean
  progress: { downloaded: number; total: number } | null
  update: { version: string; notes: string } | null
  error: string | null
  onCheck: () => void
  onInstall: () => void
}

function UpdateControl({
  hasUpdate,
  checking,
  installing,
  progress,
  update,
  error,
  onCheck,
  onInstall,
}: UpdateControlProps) {
  if (installing) {
    const pct = progress && progress.total > 0
      ? Math.round((progress.downloaded / progress.total) * 100)
      : null
    return (
      <div className="flex items-center gap-2 text-[12px] text-muted-foreground">
        <Download className="h-3.5 w-3.5 animate-pulse" />
        <span>Installing v{update?.version}{pct !== null ? ` · ${pct}%` : '…'}</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center gap-2">
        <AlertCircle className="h-3.5 w-3.5 text-destructive" />
        <span className="text-[12px] text-destructive truncate max-w-[240px]" title={error}>{error}</span>
        <Button size="sm" variant="ghost" onClick={onCheck}>Retry</Button>
      </div>
    )
  }

  if (hasUpdate && update) {
    return (
      <div className="flex items-center gap-2">
        <Mono value={`v${update.version} available`} />
        <Button size="sm" onClick={onInstall}>Install</Button>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      {checking ? (
        <>
          <RefreshCw className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
          <span className="text-[12px] text-muted-foreground">Checking…</span>
        </>
      ) : (
        <>
          <CheckCircle2 className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="text-[12px] text-muted-foreground">Up to date</span>
        </>
      )}
      <Button size="sm" variant="ghost" onClick={onCheck} disabled={checking}>Check now</Button>
    </div>
  )
}

function Mono({ value }: { value: string }) {
  return <span className="font-mono text-[12px] text-muted-foreground">{value}</span>
}
