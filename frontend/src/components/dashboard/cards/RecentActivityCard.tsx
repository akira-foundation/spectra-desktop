import { History } from 'lucide-react'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { shortUrl, timeAgo, statusTone } from '@/lib/format'
import { Card } from './Card'

interface Entry {
  id: string
  method: string
  url: string
  responseStatus: number
  durationMs: number
  error?: string
  createdAt: string | Date
  endpointID?: string
}

interface Props {
  entries: Entry[]
  onOpen: (endpointID?: string) => void
}

export function RecentActivityCard({ entries, onOpen }: Props) {
  const { getMethodColor } = useHttpMethod()
  return (
    <Card title="Recent activity" icon={History}>
      {entries.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">No requests run yet.</p>
      ) : (
        <ul className="space-y-px">
          {entries.map((entry) => {
            const tone = statusTone(entry.responseStatus, entry.error)
            return (
              <li key={entry.id}>
                <button
                  type="button"
                  onClick={() => onOpen(entry.endpointID)}
                  className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 transition-colors"
                >
                  <span
                    className={cn(
                      'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                      getMethodColor(entry.method),
                    )}
                  >
                    {entry.method}
                  </span>
                  <span className={cn('text-[11px] font-mono tabular-nums w-9 text-right shrink-0', tone)}>
                    {entry.error ? 'ERR' : entry.responseStatus}
                  </span>
                  <span className="text-[11.5px] font-mono truncate flex-1 text-left text-foreground/85">
                    {shortUrl(entry.url)}
                  </span>
                  <span className="text-[10px] text-muted-foreground tabular-nums shrink-0">
                    {entry.durationMs}ms
                  </span>
                  <span className="text-[10px] text-muted-foreground/70 shrink-0 w-12 text-right">
                    {timeAgo(new Date(entry.createdAt))}
                  </span>
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </Card>
  )
}
