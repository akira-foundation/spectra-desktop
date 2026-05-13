import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Info, Download, RefreshCw, CheckCircle2, AlertCircle } from 'lucide-react'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { Button } from '@/components/ui/button'
import { useUpdatesStore, type UpdatePhase } from '@/store/updatesStore'
import { AppPlatform, AppChannel } from '../../../wailsjs/go/app/App'

export function AboutPanel() {
  const phase = useUpdatesStore((s) => s.phase)
  const info = useUpdatesStore((s) => s.info)
  const currentVersion = useUpdatesStore((s) => s.currentVersion)
  const progress = useUpdatesStore((s) => s.progress)
  const error = useUpdatesStore((s) => s.error)
  const check = useUpdatesStore((s) => s.check)
  const install = useUpdatesStore((s) => s.install)

  const runCheck = async () => {
    await check()
    const next = useUpdatesStore.getState()
    if (next.phase === 'error' && next.error) {
      toast.error(next.error)
    } else if (next.info) {
      toast.success(`Update available: v${next.info.version}`)
    } else {
      toast.success('You’re up to date')
    }
  }

  const [platform, setPlatform] = useState('')
  const [channel, setChannel] = useState('')

  useEffect(() => {
    void AppPlatform().then(setPlatform)
    void AppChannel().then(setChannel)
  }, [])

  return (
    <div>
      <SettingsHeader
        icon={Info}
        title="About Spectra"
        description="Everything Spectra knows about the build running on this machine."
      />

      <SettingsCard>
        <SettingsRow
          label="Version"
          description={channel ? capitalize(channel) + ' channel' : undefined}
          control={<Mono value={currentVersion || '—'} />}
        />
        <SettingsRow label="Platform" control={<Mono value={platform || '—'} />} />
        <SettingsRow
          label="Updates"
          description="Use the Updates page for channel and cadence settings."
          control={
            <UpdateControl
              phase={phase}
              version={info?.version ?? null}
              progress={progress}
              error={error}
              onCheck={runCheck}
              onInstall={install}
            />
          }
        />
      </SettingsCard>
    </div>
  )
}

interface UpdateControlProps {
  phase: UpdatePhase
  version: string | null
  progress: { downloaded: number; total: number } | null
  error: string | null
  onCheck: () => void
  onInstall: () => void
}

function UpdateControl({ phase, version, progress, error, onCheck, onInstall }: UpdateControlProps) {
  if (phase === 'downloading' || phase === 'ready') {
    const pct =
      progress && progress.total > 0 ? Math.round((progress.downloaded / progress.total) * 100) : null
    return (
      <div className="flex items-center gap-2 text-[12px] text-muted-foreground">
        <Download className="h-3.5 w-3.5 animate-pulse" />
        <span>
          Installing v{version}
          {pct !== null ? ` · ${pct}%` : '…'}
        </span>
      </div>
    )
  }

  if (phase === 'error' && error) {
    return (
      <div className="flex items-center gap-2">
        <AlertCircle className="h-3.5 w-3.5 text-destructive" />
        <span className="text-[12px] text-destructive truncate max-w-[240px]" title={error}>
          {error}
        </span>
        <Button size="sm" variant="ghost" onClick={onCheck}>
          Retry
        </Button>
      </div>
    )
  }

  if (phase === 'available' && version) {
    return (
      <div className="flex items-center gap-2">
        <Mono value={`v${version} available`} />
        <Button size="sm" onClick={onInstall}>
          Install
        </Button>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      {phase === 'checking' ? (
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
      <Button size="sm" variant="ghost" onClick={onCheck} disabled={phase === 'checking'}>
        Check now
      </Button>
    </div>
  )
}

function Mono({ value }: { value: string }) {
  return <span className="font-mono text-[12px] text-muted-foreground">{value}</span>
}

function capitalize(s: string): string {
  if (!s) return s
  return s[0].toUpperCase() + s.slice(1)
}
