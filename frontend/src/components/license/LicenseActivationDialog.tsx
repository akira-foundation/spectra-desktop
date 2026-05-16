import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { KeyRound, CheckCircle2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useLicenseStore } from '@/store/licenseStore'

interface Props {
  open: boolean
  onClose: () => void
}

export function LicenseActivationDialog({ open, onClose }: Props) {
  const activate = useLicenseStore((s) => s.activate)
  const license = useLicenseStore((s) => s.license)
  const activating = useLicenseStore((s) => s.activating)
  const [deviceName, setDeviceName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (!open) {
      setError(null)
      setSuccess(false)
      setDeviceName('')
    }
  }, [open])

  const run = async () => {
    setError(null)
    try {
      await activate(deviceName.trim())
      setSuccess(true)
      toast.success('Device activated')
      setTimeout(onClose, 900)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-primary" />
            Activate this device
          </DialogTitle>
          <DialogDescription>
            Spectra binds your licence to this machine. The signed snapshot is verified locally
            on every launch.
          </DialogDescription>
        </DialogHeader>

        {success ? (
          <div className="py-6 text-center space-y-2">
            <CheckCircle2 className="h-10 w-10 text-emerald-500 mx-auto" />
            <p className="text-[13px] font-medium">Device activated</p>
            <p className="text-[11.5px] text-muted-foreground">
              {license?.plan ? `Plan: ${license.plan}` : 'License snapshot stored locally.'}
            </p>
          </div>
        ) : (
          <div className="space-y-3 py-2">
            <label className="block">
              <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
                Device name (optional)
              </span>
              <input
                value={deviceName}
                onChange={(e) => setDeviceName(e.target.value)}
                placeholder="MacBook Pro · Studio"
                className="mt-1 h-9 w-full rounded-md border border-border/60 bg-input/60 dark:bg-input/40 px-2.5 text-[13px] focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </label>
            {error && <p className="text-[12px] text-destructive">{error}</p>}
          </div>
        )}

        {!success && (
          <DialogFooter>
            <Button variant="ghost" onClick={onClose} disabled={activating}>
              Cancel
            </Button>
            <Button onClick={() => void run()} disabled={activating}>
              {activating ? 'Activating…' : 'Activate device'}
            </Button>
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  )
}
