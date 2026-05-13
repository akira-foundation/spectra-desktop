import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Terminal, CheckCircle2, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard } from './SettingsRow'
import {
  GetPHPBinaryPath,
  SetPHPBinaryPath,
  DetectPHPBinary,
} from '../../../wailsjs/go/app/App'

export function RuntimePanel() {
  const [value, setValue] = useState('')
  const [detected, setDetected] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    void GetPHPBinaryPath().then(setValue)
    void DetectPHPBinary().then(setDetected)
  }, [])

  const save = async () => {
    setBusy(true)
    try {
      await SetPHPBinaryPath(value.trim())
      toast.success(value.trim() ? 'PHP path saved' : 'PHP path cleared')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const effectivePath = value.trim() || detected
  const status: 'ok' | 'missing' = effectivePath ? 'ok' : 'missing'

  return (
    <div>
      <SettingsHeader
        icon={Terminal}
        title="Runtime"
        description="Binaries Spectra calls when scanning projects."
      />

      <SettingsCard>
        <div className="px-4 py-4 space-y-3">
          <div className="flex items-center justify-between">
            <p className="text-[13px] font-medium text-foreground/85">PHP binary</p>
            <StatusPill status={status} />
          </div>

          <div className="flex items-center gap-2">
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={detected || '/path/to/php'}
              className="h-7 text-[12px] font-mono flex-1"
            />
            <Button
              size="sm"
              onClick={save}
              disabled={busy}
              className="h-7 px-3 text-[11px] shrink-0"
            >
              {busy ? 'Saving…' : 'Save'}
            </Button>
          </div>

          {detected ? (
            <p className="text-[11.5px] text-muted-foreground">
              Detected{' '}
              <code className="font-mono text-foreground/80">{detected}</code>
              {value.trim() && value.trim() !== detected && (
                <>
                  {' · '}
                  <button
                    type="button"
                    onClick={() => setValue(detected)}
                    className="text-primary hover:underline"
                  >
                    revert to detected
                  </button>
                </>
              )}
              {!value.trim() && ' · used by default'}
            </p>
          ) : (
            <p className="text-[11.5px] text-amber-600 dark:text-amber-400">
              PHP not found on PATH or in your login shell. Laravel projects need a path here.
            </p>
          )}
        </div>
      </SettingsCard>
    </div>
  )
}

function StatusPill({ status }: { status: 'ok' | 'missing' }) {
  if (status === 'ok') {
    return (
      <span className="inline-flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400">
        <CheckCircle2 className="h-3 w-3" />
        Configured
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 text-[11px] text-amber-600 dark:text-amber-400">
      <AlertCircle className="h-3 w-3" />
      Not configured
    </span>
  )
}
