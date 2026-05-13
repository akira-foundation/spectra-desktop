import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Terminal, Wand2, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
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

  const useDetected = () => {
    if (!detected) return
    setValue(detected)
  }

  return (
    <div>
      <SettingsHeader
        icon={Terminal}
        title="Runtime"
        description="Where Spectra finds binaries it needs to scan projects."
      />

      <SettingsCard>
        <SettingsRow
          label="PHP binary"
          description={
            detected
              ? `Auto-detected: ${detected}. Override below if you want a different version.`
              : 'PHP not found automatically. Provide a full path so Laravel projects can be scanned.'
          }
          control={
            <div className="flex items-center gap-2 w-[360px]">
              <Input
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={detected || '/path/to/php'}
                className="h-7 text-[12px] font-mono"
              />
              <Button
                size="icon-sm"
                variant="ghost"
                onClick={useDetected}
                disabled={!detected}
                title="Use auto-detected"
              >
                <Wand2 className="h-3.5 w-3.5" />
              </Button>
              <Button size="icon-sm" onClick={save} disabled={busy} title="Save">
                <Save className="h-3.5 w-3.5" />
              </Button>
            </div>
          }
        />
      </SettingsCard>
    </div>
  )
}
