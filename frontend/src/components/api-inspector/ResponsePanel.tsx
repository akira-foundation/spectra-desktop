import { useEffect, useState } from 'react'
import { Copy, Download, Trash2, RefreshCw, Play, ChevronLeft } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { historyService } from '@/services/historyService'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { JsonEditor } from './JsonEditor'
import { useHistoryStore } from '@/store/historyStore'
import type { HistoryListItem } from '@/services/historyService'
import { useProjectStore } from '@/store/projectStore'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

const EMPTY_HISTORY: HistoryListItem[] = []

interface ResponsePanelProps {
  responseData: any
  responseHeaders?: Record<string, string[]>
  onReplay?: (entryId: string) => void
  endpointId?: string
}

export function ResponsePanel({
  responseData,
  responseHeaders,
  onReplay,
  endpointId,
}: ResponsePanelProps) {
  const formatted = formatBody(responseData)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const allHistory = useHistoryStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_HISTORY : EMPTY_HISTORY,
  )
  const history = endpointId
    ? allHistory.filter((h) => h.endpointID === endpointId)
    : allHistory
  const [expandedId, setExpandedId] = useState<string | null>(null)
  useEffect(() => {
    setExpandedId(null)
  }, [endpointId, activeProjectId])
  const refreshHistory = useHistoryStore((s) => s.refresh)
  const clearHistory = useHistoryStore((s) => s.clear)
  const loadHistory = useHistoryStore((s) => s.load)

  useEffect(() => {
    if (activeProjectId) void loadHistory(activeProjectId)
  }, [activeProjectId, loadHistory])

  return (
    <div className="flex flex-col min-w-0 min-h-0 h-full bg-transparent">
      <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
        <div className="flex items-center gap-1.5">
          <Download className="w-3.5 h-3.5 text-muted-foreground" />
          <h3 className="text-[11.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            Response
          </h3>
        </div>
        <Button variant="ghost" size="icon-sm" className="h-6 w-6">
          <Copy className="w-3 h-3 text-muted-foreground" />
        </Button>
      </div>

      <Tabs defaultValue="json" className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {[
            { v: 'json', label: 'JSON' },
            { v: 'raw', label: 'Raw' },
            { v: 'headers', label: 'Headers' },
            { v: 'history', label: `History${history.length > 0 ? ` · ${history.length}` : ''}` },
          ].map((t) => (
            <TabsTrigger
              key={t.v}
              value={t.v}
              className="text-[11.5px] px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
            >
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="json" className="flex-1 min-h-0 p-3 overflow-hidden mt-0">
          {responseData == null ? (
            <p className="h-full flex items-center justify-center text-[11.5px] text-muted-foreground italic">
              No response yet
            </p>
          ) : (
            <div className="h-full min-h-0 overflow-auto">
              <JsonEditor value={formatted} onChange={() => undefined} readOnly />
            </div>
          )}
        </TabsContent>

        <TabsContent value="raw" className="flex-1 p-3 overflow-auto mt-0">
          <pre className="text-[11.5px] text-foreground font-mono whitespace-pre-wrap">
            {formatted}
          </pre>
        </TabsContent>

        <TabsContent value="headers" className="flex-1 p-3 overflow-auto mt-0 mt-0">
          <HeadersList headers={responseHeaders} />
        </TabsContent>

        <TabsContent value="history" className="flex-1 min-h-0 mt-0 overflow-hidden flex flex-col">
          {expandedId ? (
            <HistoryDetailView
              entryId={expandedId}
              entry={history.find((h) => h.id === expandedId)}
              onBack={() => setExpandedId(null)}
              onReplay={onReplay}
            />
          ) : (
            <>
              <div className="px-3 py-1.5 border-b border-border/40 flex items-center justify-between">
                <span className="text-[10.5px] text-muted-foreground">{history.length} runs</span>
                <div className="flex items-center gap-1">
                  <Button
                    size="icon-sm"
                    variant="ghost"
                    className="h-6 w-6"
                    onClick={() => activeProjectId && refreshHistory(activeProjectId)}
                    title="Refresh"
                  >
                    <RefreshCw className="w-3 h-3" />
                  </Button>
                  <Button
                    size="icon-sm"
                    variant="ghost"
                    className="h-6 w-6 text-muted-foreground hover:text-destructive"
                    onClick={() => activeProjectId && clearHistory(activeProjectId)}
                    title="Clear history"
                    disabled={history.length === 0}
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              </div>
              <div className="flex-1 overflow-auto">
                {history.length === 0 ? (
                  <p className="p-4 text-center text-[11.5px] text-muted-foreground italic">
                    No requests yet. Execute a request to see history.
                  </p>
                ) : (
                  <ul className="space-y-px p-1">
                    {history.map((entry) => (
                      <HistoryRow
                        key={entry.id}
                        entry={entry}
                        onReplay={onReplay}
                        onOpen={() => setExpandedId(entry.id)}
                      />
                    ))}
                  </ul>
                )}
              </div>
            </>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}

interface HistoryRowProps {
  entry: ReturnType<typeof useHistoryStore.getState>['byProject'][string][number]
  onReplay?: (entryId: string) => void
  onOpen: () => void
}

function HistoryRow({ entry, onReplay, onOpen }: HistoryRowProps) {
  const { getMethodColor } = useHttpMethod()
  const ago = timeAgo(new Date(entry.createdAt))
  const statusTone =
    entry.error
      ? 'text-destructive'
      : entry.responseStatus >= 500
        ? 'text-destructive'
        : entry.responseStatus >= 400
          ? 'text-amber-500'
          : entry.responseStatus >= 200
            ? 'text-emerald-500'
            : 'text-muted-foreground'

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
          <span className={cn('text-[11px] font-mono tabular-nums w-9 text-right shrink-0', statusTone)}>
            {entry.error ? 'ERR' : entry.responseStatus}
          </span>
          <span className="text-[11.5px] font-mono truncate flex-1 text-foreground/85">
            {shortUrl(entry.url)}
          </span>
          <span className="text-[10px] text-muted-foreground tabular-nums shrink-0">
            {entry.durationMs}ms
          </span>
          <span className="text-[10px] text-muted-foreground/70 shrink-0 w-12 text-right">
            {ago}
          </span>
        </button>
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

interface HistoryDetailViewProps {
  entryId: string
  entry?: ReturnType<typeof useHistoryStore.getState>['byProject'][string][number]
  onBack: () => void
  onReplay?: (entryId: string) => void
}

function HistoryDetailView({ entryId, entry, onBack, onReplay }: HistoryDetailViewProps) {
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

  const statusTone = entry?.error
    ? 'text-destructive'
    : (entry?.responseStatus ?? 0) >= 500
      ? 'text-destructive'
      : (entry?.responseStatus ?? 0) >= 400
        ? 'text-amber-500'
        : (entry?.responseStatus ?? 0) >= 200
          ? 'text-emerald-500'
          : 'text-muted-foreground'

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
            <span className={cn('text-[11px] font-mono tabular-nums', statusTone)}>
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
            <JsonEditor
              value={prettyJSON(detail.responseBody) || ''}
              onChange={() => undefined}
              readOnly
            />
          </div>
        ) : (
          <p className="text-[11.5px] text-muted-foreground italic">Failed to load entry</p>
        )}
      </div>
    </div>
  )
}

function prettyJSON(raw: string): string {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

function HeadersList({ headers }: { headers?: Record<string, string[]> }) {
  const entries = Object.entries(headers ?? {})
  if (entries.length === 0) {
    return <p className="text-[11.5px] text-muted-foreground italic text-center">No headers</p>
  }
  return (
    <ul className="space-y-1">
      {entries.map(([k, vs]) => (
        <li key={k} className="flex gap-2 text-[11.5px] font-mono">
          <span className="text-muted-foreground/80 shrink-0 min-w-[140px]">{k}:</span>
          <span className="text-foreground/90 break-all">{vs.join(', ')}</span>
        </li>
      ))}
    </ul>
  )
}

function formatBody(data: unknown): string {
  if (data == null) return ''
  if (typeof data === 'string') return data
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

function shortUrl(url: string): string {
  try {
    const u = new URL(url)
    return u.pathname + u.search
  } catch {
    return url
  }
}

function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}
