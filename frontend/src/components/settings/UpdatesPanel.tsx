import {
  Download,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  ChevronDown,
} from 'lucide-react'
import toast from 'react-hot-toast'
import { Button } from '@/components/ui/button'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { useUpdatesStore, type UpdatePhase } from '@/store/updatesStore'
import { useState } from 'react'
import { cn } from '@/lib/utils'

type Channel = 'stable' | 'beta'
type Cadence = 'launch' | 'daily' | 'weekly' | 'manual'

export function UpdatesPanel() {
  const phase = useUpdatesStore((s) => s.phase)
  const info = useUpdatesStore((s) => s.info)
  const currentVersion = useUpdatesStore((s) => s.currentVersion)
  const progress = useUpdatesStore((s) => s.progress)
  const error = useUpdatesStore((s) => s.error)
  const lastCheckedAt = useUpdatesStore((s) => s.lastCheckedAt)
  const check = useUpdatesStore((s) => s.check)
  const install = useUpdatesStore((s) => s.install)
  const dismiss = useUpdatesStore((s) => s.dismiss)

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

  const [channel, setChannel] = useState<Channel>('stable')
  const [cadence, setCadence] = useState<Cadence>('launch')
  const [autoDownload, setAutoDownload] = useState(true)
  const [autoInstall, setAutoInstall] = useState(false)
  const [betaPrompt, setBetaPrompt] = useState(true)

  const percent =
    progress && progress.total > 0
      ? Math.min(100, Math.round((progress.downloaded / progress.total) * 100))
      : null
  const downloadedMB = progress ? (progress.downloaded / (1024 * 1024)).toFixed(1) : '0.0'
  const totalMB = progress && progress.total > 0 ? (progress.total / (1024 * 1024)).toFixed(1) : '?'

  return (
    <div>
      <SettingsHeader
        icon={Download}
        title="Updates"
        description="Stay current with new releases. Spectra checks the public update feed."
      />

      <div className="space-y-4">
        <UpdateBanner
          phase={phase}
          version={info?.version ?? null}
          notes={info?.notes ?? ''}
          percent={percent}
          downloadedMB={downloadedMB}
          totalMB={totalMB}
          error={error}
          onCheck={runCheck}
          onInstall={install}
          onDismiss={dismiss}
        />

        <SettingsCard>
          <SettingsRow
            label="Channel"
            description="Stable is signed and notarized. Beta gets early features and may break."
            control={
              <SegmentedToggle
                value={channel}
                options={[
                  { value: 'stable', label: 'Stable' },
                  { value: 'beta', label: 'Beta' },
                ]}
                onChange={(v) => setChannel(v as Channel)}
              />
            }
          />
          <SettingsRow
            label="Check for updates"
            description="When Spectra reaches out to the release feed."
            control={
              <SelectButton
                value={cadence}
                options={[
                  { value: 'launch', label: 'At launch' },
                  { value: 'daily', label: 'Daily' },
                  { value: 'weekly', label: 'Weekly' },
                  { value: 'manual', label: 'Manual only' },
                ]}
                onChange={(v) => setCadence(v as Cadence)}
              />
            }
          />
          <SettingsRow
            label="Download in background"
            description="Pull the installer as soon as it's available so the next launch is instant."
            control={<Toggle value={autoDownload} onChange={setAutoDownload} />}
          />
          <SettingsRow
            label="Install after relaunch"
            description="Apply downloaded updates automatically next time you quit Spectra."
            control={<Toggle value={autoInstall} onChange={setAutoInstall} />}
          />
          <SettingsRow
            label="Notify before switching to beta"
            description="Ask first if a Beta build supersedes the current Stable install."
            control={<Toggle value={betaPrompt} onChange={setBetaPrompt} />}
          />
        </SettingsCard>

        <SettingsCard>
          <SettingsRow label="Current version" control={<Mono value={currentVersion || '—'} />} />
          <SettingsRow
            label="Last checked"
            control={<Mono value={lastCheckedAt ? formatRelative(lastCheckedAt) : 'Never'} />}
          />
          <SettingsRow
            label="Latest available"
            control={<Mono value={info ? info.version : currentVersion || '—'} />}
          />
        </SettingsCard>
      </div>
    </div>
  )
}

interface UpdateBannerProps {
  phase: UpdatePhase
  version: string | null
  notes: string
  percent: number | null
  downloadedMB: string
  totalMB: string
  error: string | null
  onCheck: () => void
  onInstall: () => void
  onDismiss: () => void
}

function UpdateBanner({
  phase,
  version,
  notes,
  percent,
  downloadedMB,
  totalMB,
  error,
  onCheck,
  onInstall,
  onDismiss,
}: UpdateBannerProps) {
  const variants: Record<UpdatePhase, string> = {
    idle: 'border-border/40 bg-card/30',
    checking: 'border-border/40 bg-card/30',
    available: 'border-primary/40 bg-primary/[0.04]',
    downloading: 'border-primary/40 bg-primary/[0.04]',
    ready: 'border-emerald-500/40 bg-emerald-500/[0.04]',
    error: 'border-destructive/40 bg-destructive/[0.04]',
  }

  return (
    <div className={cn('rounded-lg border px-4 py-3.5 transition-colors', variants[phase])}>
      <div className="flex items-start gap-3">
        <PhaseIcon phase={phase} />
        <div className="flex-1 min-w-0">
          <p className="text-[13px] font-semibold text-foreground/90">
            {phaseTitle(phase, version)}
          </p>
          <p className="text-[11.5px] text-muted-foreground mt-0.5 leading-snug">
            {phase === 'error' && error ? error : phaseSubtitle(phase)}
          </p>

          {phase === 'available' && notes && (
            <div className="mt-3 rounded-md border border-border/40 bg-background/40 px-3 py-2.5 space-y-1.5">
              <p className="text-[11px] uppercase tracking-wider text-muted-foreground/70">
                Release notes
              </p>
              <p className="text-[11.5px] text-foreground/85 whitespace-pre-line">{notes}</p>
            </div>
          )}

          {phase === 'downloading' && (
            <div className="mt-3">
              <div className="h-1.5 rounded-full bg-border/40 overflow-hidden">
                <div
                  className="h-full bg-primary transition-all"
                  style={{ width: percent !== null ? `${percent}%` : '40%' }}
                />
              </div>
              <p className="mt-1 text-[10.5px] tabular-nums text-muted-foreground">
                {percent !== null ? `${percent}%` : 'Downloading'} · {downloadedMB} MB of {totalMB} MB
              </p>
            </div>
          )}
        </div>

        <div className="shrink-0 flex items-center gap-2">
          {phase === 'idle' && (
            <Button size="sm" variant="outline" onClick={onCheck}>
              Check now
            </Button>
          )}
          {phase === 'checking' && (
            <Button size="sm" variant="outline" disabled>
              <RefreshCw className="h-3.5 w-3.5 animate-spin" />
              Checking…
            </Button>
          )}
          {phase === 'available' && (
            <>
              <Button size="sm" variant="ghost" onClick={onDismiss}>
                Later
              </Button>
              <Button size="sm" onClick={onInstall}>
                <Download className="h-3.5 w-3.5" />
                Install
              </Button>
            </>
          )}
          {phase === 'downloading' && (
            <Button size="sm" variant="outline" disabled>
              {percent !== null ? `${percent}%` : 'Working…'}
            </Button>
          )}
          {phase === 'ready' && (
            <Button size="sm" disabled>
              Relaunching…
            </Button>
          )}
          {phase === 'error' && (
            <Button size="sm" variant="outline" onClick={onCheck}>
              Retry
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

function PhaseIcon({ phase }: { phase: UpdatePhase }) {
  if (phase === 'available' || phase === 'downloading') {
    return (
      <div className="h-9 w-9 rounded-md bg-primary/15 border border-primary/30 flex items-center justify-center">
        <Download className="h-4 w-4 text-primary" />
      </div>
    )
  }
  if (phase === 'ready') {
    return (
      <div className="h-9 w-9 rounded-md bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center">
        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
      </div>
    )
  }
  if (phase === 'error') {
    return (
      <div className="h-9 w-9 rounded-md bg-destructive/15 border border-destructive/30 flex items-center justify-center">
        <AlertCircle className="h-4 w-4 text-destructive" />
      </div>
    )
  }
  return (
    <div className="h-9 w-9 rounded-md bg-accent/40 border border-border/40 flex items-center justify-center">
      <RefreshCw
        className={cn('h-4 w-4 text-muted-foreground', phase === 'checking' && 'animate-spin')}
      />
    </div>
  )
}

function phaseTitle(phase: UpdatePhase, version: string | null) {
  switch (phase) {
    case 'idle':
      return 'You’re up to date'
    case 'checking':
      return 'Checking the release feed…'
    case 'available':
      return version ? `Spectra ${version} is available` : 'Update available'
    case 'downloading':
      return version ? `Downloading Spectra ${version}` : 'Downloading update'
    case 'ready':
      return 'Update installed · relaunching'
    case 'error':
      return 'Update check failed'
  }
}

function phaseSubtitle(phase: UpdatePhase) {
  switch (phase) {
    case 'idle':
      return 'New releases will appear here as soon as Spectra checks.'
    case 'checking':
      return 'This usually takes a couple of seconds.'
    case 'available':
      return 'Install replaces the running bundle and relaunches Spectra.'
    case 'downloading':
      return 'Spectra keeps working while the new build is fetched in the background.'
    case 'ready':
      return 'The new version is taking over. This window will reopen.'
    case 'error':
      return 'Network or signature error. Check your connection and try again.'
  }
}

function formatRelative(iso: string): string {
  const diff = Math.max(0, Date.now() - new Date(iso).getTime())
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min} min ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`
  return new Date(iso).toLocaleString()
}

function Toggle({ value, onChange }: { value: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={() => onChange(!value)}
      className={cn(
        'h-6 w-11 rounded-full transition-colors relative shrink-0',
        value ? 'bg-primary' : 'bg-muted-foreground/30',
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-5 w-5 rounded-full bg-background shadow transition-all',
          value ? 'left-[22px]' : 'left-0.5',
        )}
      />
    </button>
  )
}

interface SegmentedOption {
  value: string
  label: string
}

function SegmentedToggle({
  value,
  options,
  onChange,
}: {
  value: string
  options: SegmentedOption[]
  onChange: (v: string) => void
}) {
  return (
    <div className="inline-flex items-center gap-1 rounded-md border border-border/40 p-0.5">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={cn(
            'px-2.5 py-0.5 text-[11px] rounded',
            value === opt.value
              ? 'bg-accent text-foreground'
              : 'text-muted-foreground hover:text-foreground',
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}

function SelectButton({
  value,
  options,
  onChange,
}: {
  value: string
  options: SegmentedOption[]
  onChange: (v: string) => void
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="appearance-none h-7 pl-2.5 pr-7 text-[12px] rounded-md border border-border/60 bg-input/60 dark:bg-input/40 focus:outline-none focus:ring-1 focus:ring-ring"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
    </div>
  )
}

function Mono({ value }: { value: string }) {
  return <span className="font-mono text-[12px] text-muted-foreground">{value}</span>
}
