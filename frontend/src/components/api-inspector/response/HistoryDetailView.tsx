import { useEffect, useState } from 'react'
import { ChevronLeft, Play } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { historyService } from '@/services/historyService'
import { JsonEditor } from '../JsonEditor'
import type { HistoryListItem } from '@/services/historyService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { shortUrl, prettyJSON, statusTone } from '@/lib/format'

interface Props {
  entryId: string
  entry?: HistoryListItem
  onBack: () => void
  onReplay?: (entryId: string) => void
}

export function HistoryDetailView({ entryId, entry, onBack, onReplay }: Props) {
  const { getMethodColor } = useHttpMethod()
  const [detail, setDetail] = useState<Awaited<ReturnType<typeof historyService.get>> | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    historyService
      .get(entryId)
      .then((d) => setDetail(d))
      .finally(() => setLoading(false))
  }, [entryId])

  const tone = statusTone(entry?.responseStatus ?? 0, entry?.error)

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="px-3 py-1.5 border-b border-border/40 flex items-center gap-2">
        <Button size="icon-sm" variant="ghost" className="h-6 w-6" onClick={onBack} title="Back">
          <ChevronLeft className="w-3.5 h-3.5" />
        </Button>
        {entry && (
          <>
            <span
              className={cn(
                'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                getMethodColor(entry.method),
              )}
            >
              {entry.method}
            </span>
            <span className={cn('text-[11px] font-mono tabular-nums', tone)}>
              {entry.error ? 'ERR' : entry.responseStatus}
            </span>
            <span className="text-[11.5px] font-mono truncate flex-1 text-foreground/85">
              {shortUrl(entry.url)}
            </span>
            <span className="text-[10px] text-muted-foreground tabular-nums shrink-0">
              {entry.durationMs}ms
            </span>
            <Button
              size="icon-sm"
              variant="ghost"
              className="h-6 w-6 hover:text-emerald-500 hover:bg-emerald-500/10"
              onClick={() => onReplay?.(entryId)}
              title="Replay"
            >
              <Play className="w-3 h-3" />
            </Button>
          </>
        )}
      </div>
      <div className="flex-1 min-h-0 p-3 overflow-hidden">
        {loading ? (
          <div className="space-y-2">
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-3/4" />
            <Skeleton className="h-3 w-2/3" />
            <Skeleton className="h-3 w-1/2" />
          </div>
        ) : detail ? (
          <div className="h-full">
            <JsonEditor value={prettyJSON(detail.responseBody) || ''} onChange={() => undefined} readOnly />
          </div>
        ) : (
          <p className="text-[11.5px] text-muted-foreground italic">Failed to load entry</p>
        )}
      </div>
    </div>
  )
}
