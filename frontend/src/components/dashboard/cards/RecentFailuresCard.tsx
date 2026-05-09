import { AlertTriangle } from 'lucide-react'
import type { HistoryListItem } from '@/services/historyService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { shortUrl, timeAgo } from '@/lib/format'

interface Props {
  entries: HistoryListItem[]
  onOpen: (id?: string) => void
}

export function RecentFailuresCard({ entries, onOpen }: Props) {
  const { getMethodColor } = useHttpMethod()
  return (
    <div className="rounded-lg border border-border/40 bg-card/30 p-4">
      <div className="flex items-center gap-1.5 mb-3">
        <AlertTriangle className="w-3 h-3 text-rose-500/80" />
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          Recent failures
        </h3>
        <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{entries.length}</span>
      </div>
      {entries.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground/70 text-center py-6">
          No failures recently. Nice.
        </p>
      ) : (
        <ul className="m-0 p-0 list-none space-y-1">
          {entries.map((e) => (
            <li key={e.id}>
              <button
                type="button"
                onClick={() => onOpen(e.endpointID)}
                className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 text-left"
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(e.method),
                  )}
                >
                  {e.method}
                </span>
                <span className="text-[10.5px] font-mono tabular-nums text-rose-500/90 shrink-0 w-9">
                  {e.error ? 'ERR' : e.responseStatus}
                </span>
                <code className="text-[11.5px] font-mono truncate flex-1 text-foreground/85">
                  {shortUrl(e.url)}
                </code>
                <span className="text-[10px] text-muted-foreground/70 shrink-0">
                  {timeAgo(new Date(e.createdAt))}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
