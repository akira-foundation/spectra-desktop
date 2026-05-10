import { useEffect, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

interface Props {
  open: boolean
  title: string
  description?: string
  confirmLabel: string
  requireConfirmation?: boolean
  onClose: () => void
  onSubmit: (passphrase: string) => Promise<void> | void
}

export function PassphraseDialog({
  open,
  title,
  description,
  confirmLabel,
  requireConfirmation,
  onClose,
  onSubmit,
}: Props) {
  const [value, setValue] = useState('')
  const [confirm, setConfirm] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) {
      setValue('')
      setConfirm('')
      setError(null)
    }
  }, [open])

  const submit = async () => {
    if (value.length < 6) {
      setError('Use at least 6 characters')
      return
    }
    if (requireConfirmation && value !== confirm) {
      setError('Passphrases do not match')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      await onSubmit(value)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>

        <div className="grid gap-3 py-2">
          <div className="grid gap-1">
            <Label className="text-[11px] text-muted-foreground">Passphrase</Label>
            <Input
              type="password"
              autoFocus
              value={value}
              onChange={(e) => setValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !requireConfirmation) {
                  e.preventDefault()
                  void submit()
                }
              }}
            />
          </div>
          {requireConfirmation && (
            <div className="grid gap-1">
              <Label className="text-[11px] text-muted-foreground">Confirm passphrase</Label>
              <Input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    void submit()
                  }
                }}
              />
            </div>
          )}
        </div>

        {error && <p className="text-[12px] text-destructive">{error}</p>}

        <DialogFooter>
          <Button variant="ghost" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={submitting}>
            {submitting ? 'Working…' : confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
