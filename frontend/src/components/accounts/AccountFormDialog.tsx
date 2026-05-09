import { useEffect, useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAccountsStore } from '@/store/accountsStore'
import type {
  AccountKind,
  APIKeyLocation,
  ProjectAccount,
  SaveAccountInput,
} from '@/services/accountsService'

interface Props {
  open: boolean
  onClose: () => void
  projectId: string
  account?: ProjectAccount | null
}

interface FormState {
  label: string
  kind: AccountKind
  scheme: string
  username: string
  password: string
  apiKey: string
  apiKeyHeader: string
  apiKeyIn: APIKeyLocation
  token: string
  refreshToken: string
  totpSecret: string
  totpParam: string
}

const empty = (): FormState => ({
  label: '',
  kind: 'bearer',
  scheme: 'Bearer',
  username: '',
  password: '',
  apiKey: '',
  apiKeyHeader: 'X-API-Key',
  apiKeyIn: 'header',
  token: '',
  refreshToken: '',
  totpSecret: '',
  totpParam: 'X-OTP',
})

function fromAccount(acc: ProjectAccount): FormState {
  // Legacy accounts (basic/oauth2/login) collapse into bearer.
  const kind: AccountKind = acc.kind === 'apikey' ? 'apikey' : 'bearer'
  return {
    ...empty(),
    label: acc.label,
    kind,
    scheme: acc.scheme || 'Bearer',
    username: acc.username,
    apiKeyHeader: acc.apiKeyHeader || 'X-API-Key',
    apiKeyIn: acc.apiKeyIn || 'header',
    totpParam: acc.totpParam || 'X-OTP',
  }
}

export function AccountFormDialog({ open, onClose, projectId, account }: Props) {
  const save = useAccountsStore((s) => s.save)
  const [form, setForm] = useState<FormState>(empty())
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setForm(account ? fromAccount(account) : empty())
    setError(null)
  }, [open, account])

  function update<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSubmit() {
    setSubmitting(true)
    setError(null)
    try {
      const input: SaveAccountInput = {
        id: account?.id,
        projectID: projectId,
        label: form.label.trim(),
        kind: form.kind,
        scheme: form.scheme.trim() || undefined,
        username: form.username,
        apiKeyHeader: form.apiKeyHeader,
        apiKeyIn: form.apiKeyIn,
        totpParam: form.totpParam,
      }
      if (form.password) input.password = form.password
      if (form.apiKey) input.apiKey = form.apiKey
      if (form.token) input.token = form.token
      if (form.totpSecret) input.totpSecret = form.totpSecret
      await save(input)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{account ? 'Edit account' : 'New account'}</DialogTitle>
        </DialogHeader>

        <div className="grid gap-3 py-2">
          <div className="grid grid-cols-2 gap-2">
            <Field label="Label">
              <Input
                value={form.label}
                onChange={(e) => update('label', e.target.value)}
                placeholder="Admin · staging"
              />
            </Field>
            <Field label="Type">
              <Select value={form.kind} onValueChange={(v) => update('kind', v as AccountKind)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="bearer">Bearer / Token</SelectItem>
                  <SelectItem value="apikey">API Key</SelectItem>
                </SelectContent>
              </Select>
            </Field>
          </div>

          {form.kind === 'bearer' && (
            <>
              <Field label="Scheme">
                <Input
                  value={form.scheme}
                  onChange={(e) => update('scheme', e.target.value)}
                  placeholder="Bearer"
                />
              </Field>

              <div className="grid grid-cols-2 gap-2">
                <Field label="Username / Email (optional)">
                  <Input
                    value={form.username}
                    onChange={(e) => update('username', e.target.value)}
                    placeholder="admin@example.com"
                  />
                </Field>
                <Field
                  label={account?.hasPassword ? 'Password (leave empty to keep)' : 'Password (optional)'}
                >
                  <Input
                    type="password"
                    value={form.password}
                    onChange={(e) => update('password', e.target.value)}
                  />
                </Field>
              </div>

              <Field label={account?.hasToken ? 'Token (leave empty to keep)' : 'Token (optional)'}>
                <Input
                  type="password"
                  value={form.token}
                  onChange={(e) => update('token', e.target.value)}
                  placeholder={account?.hasToken ? '••••••••' : 'eyJhbGciOi...'}
                />
              </Field>

              <div className="rounded border border-border/40 bg-muted/30 px-3 py-2 text-[11px] text-muted-foreground space-y-1">
                <p className="font-medium text-foreground/80">How it works</p>
                <p>
                  Set this account active in the Inspector. If you fill credentials, run
                  the project's login endpoint and the token is captured automatically.
                  Otherwise paste a token directly above.
                </p>
              </div>
            </>
          )}

          {form.kind === 'apikey' && (
            <>
              <Field label={account?.hasApiKey ? 'Key (leave empty to keep)' : 'Key'}>
                <Input
                  type="password"
                  value={form.apiKey}
                  onChange={(e) => update('apiKey', e.target.value)}
                />
              </Field>
              <div className="grid grid-cols-2 gap-2">
                <Field label="Name">
                  <Input
                    value={form.apiKeyHeader}
                    onChange={(e) => update('apiKeyHeader', e.target.value)}
                  />
                </Field>
                <Field label="Location">
                  <Select
                    value={form.apiKeyIn}
                    onValueChange={(v) => update('apiKeyIn', v as APIKeyLocation)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="header">Header</SelectItem>
                      <SelectItem value="query">Query</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </div>
            </>
          )}

          <div className="grid grid-cols-2 gap-2 border-t border-border/40 pt-3">
            <Field
              label={account?.hasTotp ? '2FA secret (leave empty to keep)' : '2FA secret (optional)'}
            >
              <Input
                type="password"
                value={form.totpSecret}
                onChange={(e) => update('totpSecret', e.target.value)}
                placeholder="JBSWY3DPEHPK3PXP"
              />
            </Field>
            <Field label="2FA injection (header or ?query)">
              <Input
                value={form.totpParam}
                onChange={(e) => update('totpParam', e.target.value)}
                placeholder="X-OTP"
              />
            </Field>
          </div>
        </div>

        {error && <p className="text-[12px] text-destructive">{error}</p>}

        <DialogFooter>
          <Button variant="ghost" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || !form.label.trim()}>
            {submitting ? 'Saving…' : 'Save'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1">
      <Label className="text-[11px] text-muted-foreground">{label}</Label>
      {children}
    </div>
  )
}
