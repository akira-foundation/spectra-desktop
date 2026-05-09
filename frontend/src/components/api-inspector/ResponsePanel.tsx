import { useEffect, useMemo, useState } from 'react'
import { Copy, Download, Trash2, Play, ChevronLeft, Check } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { historyService } from '@/services/historyService'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { JsonEditor } from './JsonEditor'
import { useHistoryStore } from '@/store/historyStore'
import type { HistoryListItem } from '@/services/historyService'
import { useProjectStore } from '@/store/projectStore'
import { useUIStore } from '@/store/uiStore'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

const EMPTY_HISTORY: HistoryListItem[] = []

interface ResponsePanelProps {
  responseData: any
  responseHeaders?: Record<string, string[]>
  onReplay?: (entryId: string) => void
  endpointId?: string
  endpointMethod?: string
  endpointPath?: string
}

export function ResponsePanel({
  responseData,
  responseHeaders,
  onReplay,
  endpointId,
  endpointMethod,
  endpointPath,
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
  const [tab, setTab] = useState<string>('json')
  const inspectorPending = useUIStore((s) => s.inspectorPending)
  const setInspectorPending = useUIStore((s) => s.setInspectorPending)
  useEffect(() => {
    setExpandedId(null)
  }, [endpointId, activeProjectId])

  useEffect(() => {
    if (!inspectorPending || !endpointId) return
    if (inspectorPending.endpointId !== endpointId) return
    if (inspectorPending.openHistoryLatest) {
      setTab('history')
      const latest = history[0]
      if (latest) setExpandedId(latest.id)
    }
    setInspectorPending(null)
  }, [inspectorPending, endpointId, history])
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
        <div className="flex items-center gap-1">
          <SaveButton text={formatted} method={endpointMethod} path={endpointPath} />
          <CopyButton text={formatted} />
        </div>
      </div>

      <Tabs value={tab} onValueChange={setTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {[
            { v: 'json', label: 'JSON' },
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
            <ResponseBodyView raw={formatted} />
          )}
        </TabsContent>

        <TabsContent value="headers" className="flex-1 p-3 overflow-auto mt-0">
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

function CopyHistoryButton({ entryId }: { entryId: string }) {
  const [copied, setCopied] = useState(false)
  const handle = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      const detail = await historyService.get(entryId)
      const body = detail?.responseBody ?? ''
      if (!body) return
      await navigator.clipboard.writeText(body)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  return (
    <button
      type="button"
      onClick={handle}
      title={copied ? 'Copied!' : 'Copy response'}
      className={cn(
        'inline-flex h-5 w-5 items-center justify-center rounded transition-all',
        copied
          ? 'text-emerald-500 bg-emerald-500/10 opacity-100'
          : 'text-muted-foreground/40 opacity-0 group-hover:opacity-100 hover:text-foreground hover:bg-accent/60',
      )}
    >
      {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
    </button>
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

function ResponseBodyView({ raw }: { raw: string }) {
  type Mode = 'json' | 'tree' | 'table' | 'raw'
  const [mode, setMode] = useState<Mode>('json')
  const parsed = useMemo(() => {
    try {
      return JSON.parse(raw)
    } catch {
      return null
    }
  }, [raw])

  const tableRows = useMemo(() => extractTableRows(parsed), [parsed])
  const isJson = parsed !== null

  return (
    <div className="h-full min-h-0 flex flex-col gap-2">
      <div className="flex items-center gap-1 shrink-0">
        {(['json', 'tree', 'table', 'raw'] as Mode[])
          .filter((m) => {
            if (m === 'json' || m === 'tree') return isJson
            if (m === 'table') return tableRows != null
            return true
          })
          .map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setMode(m)}
              className={cn(
                'h-6 px-2 text-[10.5px] rounded transition-colors',
                mode === m ? 'bg-primary/15 text-primary' : 'text-muted-foreground hover:bg-accent/40',
              )}
            >
              {m.toUpperCase()}
            </button>
          ))}
      </div>
      <div className="flex-1 min-h-0 overflow-auto">
        {mode === 'json' && (isJson ? <JsonEditor value={raw} onChange={() => undefined} readOnly /> : <RawView raw={raw} />)}
        {mode === 'tree' && parsed !== null && <TreeView value={parsed} />}
        {mode === 'table' && tableRows && <TableView rows={tableRows} />}
        {mode === 'raw' && <RawView raw={raw} />}
      </div>
    </div>
  )
}

function RawView({ raw }: { raw: string }) {
  return (
    <pre className="h-full w-full m-0 p-3 text-[11px] font-mono whitespace-pre-wrap break-all text-foreground/85 bg-muted/20 rounded-md border border-border/40 overflow-auto">
      {raw}
    </pre>
  )
}

function TreeView({ value, depth = 0, label }: { value: any; depth?: number; label?: string }) {
  const [open, setOpen] = useState(depth < 2)
  const isObj = value && typeof value === 'object'
  const isArr = Array.isArray(value)
  if (!isObj) {
    return (
      <div className="flex items-baseline gap-2 py-0.5" style={{ paddingLeft: depth * 12 }}>
        {label && <code className="text-[11px] font-mono text-foreground/70">{label}:</code>}
        <code className={cn('text-[11px] font-mono', valueTone(value))}>{formatLeaf(value)}</code>
      </div>
    )
  }
  const entries = isArr
    ? (value as any[]).map((v, i) => [`[${i}]`, v] as [string, any])
    : Object.entries(value as Record<string, any>)
  const summary = isArr ? `Array(${entries.length})` : `{${entries.length}}`
  return (
    <div style={{ paddingLeft: depth * 12 }}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex items-baseline gap-2 py-0.5 hover:bg-accent/20 px-1 rounded"
      >
        <span className="text-[10px] text-muted-foreground/60 w-3">{open ? '▼' : '▶'}</span>
        {label && <code className="text-[11px] font-mono text-foreground/70">{label}:</code>}
        <code className="text-[10.5px] font-mono text-muted-foreground">{summary}</code>
      </button>
      {open && (
        <div>
          {entries.map(([k, v]) => (
            <TreeView key={k} value={v} depth={depth + 1} label={k} />
          ))}
        </div>
      )}
    </div>
  )
}

function TableView({ rows }: { rows: { columns: string[]; data: any[][] } }) {
  return (
    <div className="overflow-auto rounded-md border border-border/40">
      <table className="w-full text-[11px] font-mono">
        <thead className="sticky top-0 bg-muted/50">
          <tr>
            {rows.columns.map((c) => (
              <th key={c} className="px-2 py-1.5 text-left text-[10px] uppercase tracking-wider font-semibold text-muted-foreground/80 border-b border-border/40">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.data.map((r, i) => (
            <tr key={i} className="border-b border-border/20 hover:bg-accent/20">
              {r.map((v, j) => (
                <td key={j} className={cn('px-2 py-1 align-top truncate max-w-[300px]', valueTone(v))}>
                  {formatLeaf(v)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function extractTableRows(value: any): { columns: string[]; data: any[][] } | null {
  const arr = findArrayOfObjects(value)
  if (!arr || arr.length === 0) return null
  const columns = Array.from(
    new Set(arr.flatMap((row) => (row && typeof row === 'object' ? Object.keys(row) : []))),
  )
  if (columns.length === 0) return null
  const data = arr.map((row) =>
    columns.map((c) => (row && typeof row === 'object' ? (row as any)[c] : undefined)),
  )
  return { columns, data }
}

function findArrayOfObjects(value: any, depth = 0): any[] | null {
  if (depth > 6) return null
  if (Array.isArray(value)) {
    if (value.length > 0 && typeof value[0] === 'object' && value[0] !== null && !Array.isArray(value[0])) {
      return value
    }
    return null
  }
  if (value && typeof value === 'object') {
    for (const v of Object.values(value)) {
      const found = findArrayOfObjects(v, depth + 1)
      if (found) return found
    }
  }
  return null
}

function formatLeaf(v: any): string {
  if (v === null || v === undefined) return v === null ? 'null' : ''
  if (typeof v === 'string') return `"${v}"`
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function valueTone(v: any): string {
  if (v === null || v === undefined) return 'text-muted-foreground/60 italic'
  if (typeof v === 'string') return 'text-emerald-500/90'
  if (typeof v === 'number') return 'text-purple-400'
  if (typeof v === 'boolean') return 'text-amber-500'
  return 'text-foreground/85'
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const handle = async () => {
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  return (
    <button
      type="button"
      onClick={handle}
      disabled={!text}
      className={cn(
        'inline-flex h-6 w-6 items-center justify-center rounded transition-colors',
        copied ? 'text-emerald-500 bg-emerald-500/10' : 'text-muted-foreground hover:text-foreground hover:bg-accent/60',
        'disabled:opacity-40 disabled:hover:bg-transparent',
      )}
      title={copied ? 'Copied!' : 'Copy response'}
    >
      {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
    </button>
  )
}

function SaveButton({ text, method, path }: { text: string; method?: string; path?: string }) {
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const handle = async () => {
    if (!text || saving) return
    setSaving(true)
    try {
      const { SaveResponseToFile } = await import('../../../wailsjs/go/app/App')
      const result = await SaveResponseToFile(method ?? '', path ?? '', text)
      if (result) {
        setSaved(true)
        setTimeout(() => setSaved(false), 1500)
      }
    } catch {} finally {
      setSaving(false)
    }
  }
  return (
    <button
      type="button"
      onClick={handle}
      disabled={!text || saving}
      className={cn(
        'inline-flex h-6 w-6 items-center justify-center rounded transition-colors',
        saved ? 'text-emerald-500 bg-emerald-500/10' : 'text-muted-foreground hover:text-foreground hover:bg-accent/60',
        'disabled:opacity-40 disabled:hover:bg-transparent',
      )}
      title={saved ? 'Saved!' : 'Save response to file'}
    >
      {saved ? <Check className="w-3 h-3" /> : <Download className="w-3 h-3" />}
    </button>
  )
}
