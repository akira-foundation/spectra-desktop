import { cn } from '@/lib/utils'
import type { AccountKind } from '@/services/accountsService'

const TONES: Record<AccountKind, string> = {
  bearer: 'bg-blue-500/15 text-blue-500 border-blue-500/30',
  basic: 'bg-amber-500/15 text-amber-500 border-amber-500/30',
  apikey: 'bg-purple-500/15 text-purple-500 border-purple-500/30',
  oauth2: 'bg-emerald-500/15 text-emerald-500 border-emerald-500/30',
  login: 'bg-rose-500/15 text-rose-500 border-rose-500/30',
}

const LABELS: Record<AccountKind, string> = {
  bearer: 'Bearer',
  basic: 'Bearer',
  apikey: 'API Key',
  oauth2: 'Bearer',
  login: 'Bearer',
}

export function AccountKindBadge({ kind, className }: { kind: AccountKind; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded border px-1.5 py-px text-[10px] font-medium uppercase tracking-wider',
        TONES[kind] ?? 'bg-muted text-muted-foreground border-border/50',
        className,
      )}
    >
      {LABELS[kind] ?? kind}
    </span>
  )
}
