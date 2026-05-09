import { useEffect, useMemo, useState } from 'react'
import { Plus, Star, Trash2, Pencil, KeyRound, ShieldCheck } from 'lucide-react'
import { useProjectStore } from '@/store/projectStore'
import { useAccountsStore } from '@/store/accountsStore'
import { Button } from '@/components/ui/button'
import { AccountKindBadge } from '@/components/accounts/AccountKindBadge'
import { AccountFormDialog } from '@/components/accounts/AccountFormDialog'
import type { ProjectAccount } from '@/services/accountsService'
import { cn } from '@/lib/utils'

const EMPTY_ACCOUNTS: ProjectAccount[] = []

export function AccountsPage() {
  const projectId = useProjectStore((s) => s.activeProjectId ?? '')
  const project = useProjectStore((s) =>
    projectId ? s.projects.find((p) => p.id === projectId) ?? null : null,
  )
  const accounts = useAccountsStore((s) => s.byProject[projectId] ?? EMPTY_ACCOUNTS)
  const list = useAccountsStore((s) => s.list)
  const remove = useAccountsStore((s) => s.remove)
  const setDefault = useAccountsStore((s) => s.setDefault)
  const totp = useAccountsStore((s) => s.totp)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<ProjectAccount | null>(null)
  const [totpCodes, setTotpCodes] = useState<Record<string, string>>({})

  useEffect(() => {
    if (projectId) void list(projectId)
  }, [projectId, list])

  const totpAccounts = useMemo(() => accounts.filter((a) => a.hasTotp), [accounts])
  useEffect(() => {
    if (totpAccounts.length === 0) return
    let cancelled = false
    const tick = async () => {
      const next: Record<string, string> = {}
      for (const acc of totpAccounts) {
        try {
          next[acc.id] = await totp(acc.id)
        } catch {}
      }
      if (!cancelled) setTotpCodes(next)
    }
    void tick()
    const interval = setInterval(() => void tick(), 5_000)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [totpAccounts, totp])

  if (!project) {
    return (
      <div className="flex h-full items-center justify-center text-muted-foreground text-[12px]">
        Select a project first
      </div>
    )
  }

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-2xl mx-auto p-6 space-y-4">
        <div className="flex items-end justify-between gap-4">
          <div className="min-w-0">
            <h1 className="text-xl font-semibold tracking-tight">Accounts</h1>
            <p className="text-muted-foreground text-[12.5px] mt-1">
              Authenticated identities for{' '}
              <span className="font-medium text-foreground/80">{project.name}</span>. Switch
              in the Inspector to test as different users.
            </p>
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              setEditing(null)
              setDialogOpen(true)
            }}
          >
            <Plus className="h-3.5 w-3.5" />
            New account
          </Button>
        </div>

        {accounts.length === 0 ? (
          <EmptyState
            onAdd={() => {
              setEditing(null)
              setDialogOpen(true)
            }}
          />
        ) : (
          <div className="rounded-md border border-border/40 overflow-hidden">
            <ul className="divide-y divide-border/40">
              {accounts.map((acc) => (
                <AccountRow
                  key={acc.id}
                  account={acc}
                  totpCode={totpCodes[acc.id]}
                  onSetDefault={() => void setDefault(projectId, acc.id)}
                  onEdit={() => {
                    setEditing(acc)
                    setDialogOpen(true)
                  }}
                  onDelete={() => {
                    if (confirm(`Delete account "${acc.label}"?`)) {
                      void remove(projectId, acc.id)
                    }
                  }}
                />
              ))}
            </ul>
          </div>
        )}
      </div>

      <AccountFormDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        projectId={projectId}
        account={editing}
      />
    </div>
  )
}

interface AccountRowProps {
  account: ProjectAccount
  totpCode?: string
  onSetDefault: () => void
  onEdit: () => void
  onDelete: () => void
}

function AccountRow({ account, totpCode, onSetDefault, onEdit, onDelete }: AccountRowProps) {
  return (
    <li className="px-3.5 py-2.5 flex items-center gap-3 hover:bg-accent/20 transition-colors">
      <button
        type="button"
        onClick={onSetDefault}
        title={account.isDefault ? 'Default account' : 'Set as default'}
        className={cn(
          'h-6 w-6 rounded-full flex items-center justify-center shrink-0 transition-colors',
          account.isDefault
            ? 'text-primary'
            : 'text-muted-foreground/60 hover:text-foreground',
        )}
      >
        <Star
          className="h-3.5 w-3.5"
          fill={account.isDefault ? 'currentColor' : 'none'}
          strokeWidth={1.75}
        />
      </button>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-medium text-[12.5px] truncate">{account.label}</span>
          <AccountKindBadge kind={account.kind} />
          {account.hasTotp && (
            <span
              className="inline-flex items-center gap-1 text-[10.5px] text-emerald-500"
              title="Two-factor authentication enabled"
            >
              <ShieldCheck className="h-3 w-3" />
              {totpCode ?? '2FA'}
            </span>
          )}
        </div>
        <p className="text-[11px] text-muted-foreground/80 mt-0.5 truncate font-mono">
          {summarize(account)}
        </p>
      </div>

      <div className="flex items-center gap-0.5 shrink-0">
        <Button size="icon-sm" variant="ghost" onClick={onEdit} title="Edit">
          <Pencil className="h-3.5 w-3.5" />
        </Button>
        <Button
          size="icon-sm"
          variant="ghost"
          className="text-muted-foreground hover:text-destructive"
          onClick={onDelete}
          title="Delete"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>
    </li>
  )
}

function EmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <div className="rounded-md border border-dashed border-border/60 px-6 py-10 text-center">
      <KeyRound className="h-7 w-7 mx-auto text-muted-foreground/50" strokeWidth={1.5} />
      <p className="text-[12.5px] mt-2 font-medium">No accounts yet</p>
      <p className="text-[11.5px] text-muted-foreground mt-1">
        Add credentials for any identity you need to test against.
      </p>
      <Button size="sm" variant="outline" className="mt-3" onClick={onAdd}>
        <Plus className="h-3.5 w-3.5" />
        Add the first account
      </Button>
    </div>
  )
}

function summarize(acc: ProjectAccount): string {
  if (acc.kind === 'apikey') {
    return `${acc.apiKeyHeader || 'X-API-Key'} via ${acc.apiKeyIn || 'header'} · ${
      acc.hasApiKey ? '••••' : 'no key'
    }`
  }
  // Bearer / legacy kinds collapse to token + optional credentials.
  const parts: string[] = []
  if (acc.hasToken) parts.push(`token ${acc.tokenPreview ?? '••••'}`)
  if (acc.username) parts.push(acc.username)
  else if (acc.hasPassword) parts.push('credentials saved')
  if (parts.length === 0) return 'No token configured'
  return parts.join(' · ')
}
