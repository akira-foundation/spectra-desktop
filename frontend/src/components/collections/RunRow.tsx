import { useState } from 'react'
import { Check, X, AlertTriangle, ChevronRight, ExternalLink } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { cn } from '@/lib/utils'

interface Props {
  item: any
  projectId: string
  getMethodColor: (m: string) => string
}

export function RunRow({ item: r, projectId, getMethodColor }: Props) {
  const [open, setOpen] = useState(false)
  const setSelectedEndpoint = useUIStore((s) => s.setSelectedEndpoint)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)
  const setInspectorPending = useUIStore((s) => s.setInspectorPending)
  const openInInspector = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!r.endpointID) return
    setSelectedEndpoint(projectId, r.endpointID)
    setInspectorPending({ endpointId: r.endpointID, openHistoryLatest: true })
    setCurrentPage('inspector')
  }
  const failedTests = (r.testResults ?? []).filter((t: any) => !t.pass)
  const passedTests = (r.testResults ?? []).filter((t: any) => t.pass)
  const reqFailed = !r.skipped && (r.error || (r.status && r.status >= 400))
  const testsFailed = !r.skipped && failedTests.length > 0
  const summary = r.skipped
    ? r.error || 'skipped'
    : reqFailed && testsFailed
      ? `request ${r.status} · ${failedTests.length} test${failedTests.length === 1 ? '' : 's'} failed`
      : reqFailed
        ? r.error
          ? `error: ${r.error}`
          : `request returned ${r.status}`
        : testsFailed
          ? `${failedTests.length} test${failedTests.length === 1 ? '' : 's'} failed`
          : null
  const hasDetail = (r.testResults && r.testResults.length > 0) || summary

  return (
    <li className="border-b border-border/20 group">
      <div className={cn('w-full flex items-center gap-2 px-3 py-1.5', hasDetail && 'hover:bg-accent/30')}>
        <button
          type="button"
          onClick={() => hasDetail && setOpen((o) => !o)}
          className={cn('flex-1 min-w-0 flex items-center gap-2 text-left', hasDetail && 'cursor-pointer')}
        >
          {hasDetail ? (
            <ChevronRight
              className={cn(
                'w-3 h-3 text-muted-foreground/60 shrink-0 transition-transform',
                open && 'rotate-90',
              )}
            />
          ) : (
            <span className="w-3 shrink-0" />
          )}
          {r.skipped ? (
            <AlertTriangle className="w-3 h-3 text-amber-500 shrink-0" />
          ) : r.pass ? (
            <Check className="w-3 h-3 text-emerald-500 shrink-0" />
          ) : (
            <X className="w-3 h-3 text-rose-500 shrink-0" />
          )}
          <span
            className={cn(
              'inline-flex w-12 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
              r.method ? getMethodColor(r.method) : 'bg-muted/40 text-muted-foreground',
            )}
          >
            {r.method}
          </span>
          <span className="text-[11px] font-mono text-foreground/85 truncate flex-1 min-w-0">{r.path}</span>
          {!r.skipped && (
            <span
              className={cn(
                'text-[10px] font-mono tabular-nums shrink-0 w-24 text-right',
                r.status >= 400 ? 'text-rose-500/90' : 'text-muted-foreground',
              )}
            >
              {r.status} · {r.durationMs}ms
            </span>
          )}
        </button>
        {r.endpointID && (
          <button
            type="button"
            onClick={openInInspector}
            title="Open in Inspector"
            className="opacity-0 group-hover:opacity-100 inline-flex h-6 w-6 items-center justify-center text-muted-foreground hover:text-foreground rounded hover:bg-accent/60 shrink-0"
          >
            <ExternalLink className="w-3 h-3" />
          </button>
        )}
      </div>
      {open && hasDetail && (
        <div className="px-3 pb-2 pl-9 space-y-1">
          {summary && <p className="text-[10.5px] text-muted-foreground italic">{summary}</p>}
          {(r.testResults ?? []).length > 0 && (
            <ul className="m-0 p-0 list-none space-y-0.5">
              {[...failedTests, ...passedTests].map((t: any, i: number) => (
                <li key={i} className="flex items-start gap-1.5 text-[10.5px] font-mono">
                  {t.pass ? (
                    <Check className="w-3 h-3 text-emerald-500/80 shrink-0 mt-0.5" />
                  ) : (
                    <X className="w-3 h-3 text-rose-500/80 shrink-0 mt-0.5" />
                  )}
                  <span className={cn('flex-1', t.pass ? 'text-muted-foreground' : 'text-foreground/85')}>
                    {t.name}
                    {t.message && <span className="text-rose-500/80"> — {t.message}</span>}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </li>
  )
}
