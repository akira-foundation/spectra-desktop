import { Play } from 'lucide-react'
import type { HistoryListItem } from '@/services/historyService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { shortUrl, timeAgo, statusTone } from '@/lib/format'
import { CopyHistoryButton } from './CopyHistoryButton'

interface Props {
  entry: HistoryListItem
  onReplay?: (entryId: string) => void
  onOpen: () => void
}

export function HistoryRow({ entry, onReplay, onOpen }: Props) {
  const { getMethodColor } = useHttpMethod()
  const ago = timeAgo(new Date(entry.createdAt))
  const tone = statusTone(entry.responseStatus, entry.error)

  return (
    <li className="rounded-md hover:bg-accent/30 transition-colors group">
      <div className="flex items-center gap-2 px-2 py-1.5">
        <button
          type="button"
          onClick={onOpen}
          className="flex items-center gap-2 flex-1 min-w-0 text-left"
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
          <span className="text-[11.5px] font-mono truncate flex-1 text-foreground/85">
            {shortUrl(entry.url)}
          </span>
          <span className="text-[10px] text-muted-foreground tabular-nums shrink-0">
            {entry.durationMs}ms
          </span>
          <span className="text-[10px] text-muted-foreground/70 shrink-0 w-12 text-right">{ago}</span>
        </button>
        <CopyHistoryButton entryId={entry.id} />
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onReplay?.(entry.id)
          }}
          aria-label="Replay this request"
          title="Replay"
          className="inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground/40 opacity-0 group-hover:opacity-100 hover:text-emerald-500 hover:bg-emerald-500/10 transition-all"
        >
          <Play className="w-3 h-3" />
        </button>
      </div>
    </li>
  )
}
