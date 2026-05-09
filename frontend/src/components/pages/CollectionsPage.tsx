import { useEffect, useMemo, useState } from 'react'
import { Plus, Trash2, Play, X, Loader2, Check, AlertTriangle, GripVertical, FolderKanban, ChevronRight, ExternalLink, Repeat, Database, Sparkles } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogTitle, DialogHeader, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { useCollectionsStore } from '@/store/collectionsStore'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { useUIStore } from '@/store/uiStore'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { collectionsService, type Collection, type CollectionRun } from '@/services/collectionsService'
import { datasetsService } from '@/services/datasetsService'
import { cn } from '@/lib/utils'

const EMPTY_COLLECTIONS: Collection[] = []
const EMPTY_ENDPOINTS: any[] = []
const BODY_METHODS = new Set(['POST', 'PUT', 'PATCH'])

export function CollectionsPage() {
  const projectId = useProjectStore((s) => s.activeProjectId)
  const collections = useCollectionsStore((s) =>
    projectId ? s.byProject[projectId] ?? EMPTY_COLLECTIONS : EMPTY_COLLECTIONS,
  )
  const lastRun = useCollectionsStore((s) => s.lastRun)
  const refresh = useCollectionsStore((s) => s.refresh)
  const runCol = useCollectionsStore((s) => s.run)
  const endpoints = useEndpointsStore((s) =>
    projectId ? (s.byProject[projectId] ?? EMPTY_ENDPOINTS) : EMPTY_ENDPOINTS,
  ) as { id: string; method: string; path: string }[]

  const [activeId, setActiveId] = useState<string | null>(null)
  const [running, setRunning] = useState<string | null>(null)
  const [progress, setProgress] = useState<{ items: any[]; total: number } | null>(null)

  useEffect(() => {
    if (projectId) void refresh(projectId)
  }, [projectId, refresh])

  useEffect(() => {
    if (!activeId && collections.length > 0) {
      setActiveId(collections[0].id)
    }
  }, [collections, activeId])

  const active = useMemo(() => collections.find((c) => c.id === activeId) ?? null, [collections, activeId])

  const createNew = async () => {
    if (!projectId) return
    const result = await collectionsService.save({
      projectID: projectId,
      name: `Collection ${collections.length + 1}`,
      items: [],
    })
    await refresh(projectId)
    if (result?.id) setActiveId(result.id)
  }

  const handleRun = async (id: string) => {
    if (!projectId) return
    setRunning(id)
    setProgress({ items: [], total: 0 })
    try {
      await runCol(projectId, id)
    } finally {
      setRunning(null)
      setProgress(null)
    }
  }

  useEffect(() => {
    EventsOn('collection:run:start', (data: any) => {
      setProgress({ items: [], total: data?.total ?? 0 })
    })
    EventsOn('collection:run:progress', (data: any) => {
      setProgress((prev) => ({
        items: [...(prev?.items ?? []), data?.item],
        total: data?.total ?? prev?.total ?? 0,
      }))
    })
    return () => {
      EventsOff('collection:run:start')
      EventsOff('collection:run:progress')
    }
  }, [])

  if (!projectId) {
    return (
      <div className="h-full flex items-center justify-center text-[12px] text-muted-foreground">
        Select a project to view collections.
      </div>
    )
  }

  return (
    <div className="h-full flex gap-2 p-2 overflow-hidden">
      <aside className="w-64 shrink-0 rounded-md border border-border/40 bg-card/30 flex flex-col">
        <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
          <div className="flex items-center gap-1.5">
            <FolderKanban className="w-3 h-3 text-muted-foreground" />
            <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
              Collections
            </span>
            <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">
              {collections.length}
            </span>
          </div>
          <button
            type="button"
            onClick={() => void createNew()}
            className="inline-flex items-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground"
          >
            <Plus className="w-3 h-3" />
            New
          </button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {collections.length === 0 ? (
            <p className="px-3 py-6 text-[11px] italic text-muted-foreground/70 text-center">
              No collections yet.
            </p>
          ) : (
            <ul className="m-0 p-0 list-none">
              {collections.map((c) => (
                <SidebarItem
                  key={c.id}
                  collection={c}
                  active={active?.id === c.id}
                  run={lastRun[c.id]}
                  onSelect={() => setActiveId(c.id)}
                  onChanged={() => projectId && refresh(projectId)}
                />
              ))}
            </ul>
          )}
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 rounded-md border border-border/40 bg-card/30 overflow-hidden">
        {!active ? (
          <div className="flex-1 flex items-center justify-center text-[12px] text-muted-foreground">
            Select or create a collection
          </div>
        ) : (
          <CollectionDetail
            key={active.id}
            projectId={projectId}
            collection={active}
            endpoints={endpoints}
            run={lastRun[active.id]}
            isRunning={running === active.id}
            progress={running === active.id ? progress : null}
            onRun={() => handleRun(active.id)}
            onChanged={() => projectId && refresh(projectId)}
          />
        )}
      </main>
    </div>
  )
}

interface SidebarItemProps {
  collection: Collection
  active: boolean
  run?: CollectionRun | null
  onSelect: () => void
  onChanged: () => void
}

function SidebarItem({ collection: c, active, run, onSelect, onChanged }: SidebarItemProps) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(c.name)

  useEffect(() => setName(c.name), [c.id, c.name])

  useEffect(() => {
    if (!editing) return
    const trimmed = name.trim()
    if (!trimmed || trimmed === c.name) return
    const t = setTimeout(async () => {
      await collectionsService.save({
        id: c.id,
        projectID: c.projectID,
        name: trimmed,
        description: c.description ?? '',
        items: c.items ?? [],
      })
      onChanged()
    }, 500)
    return () => clearTimeout(t)
  }, [name, editing])

  const commit = () => {
    const trimmed = name.trim()
    if (!trimmed) setName(c.name)
    setEditing(false)
  }

  const remove = async () => {
    await collectionsService.remove(c.id)
    onChanged()
  }

  return (
    <li>
      <div
        className={cn(
          'group flex items-center gap-1.5 px-2 py-1.5 text-[11.5px] hover:bg-accent/40',
          active && 'bg-accent/60',
        )}
      >
        {editing ? (
          <input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            onBlur={() => void commit()}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void commit()
              if (e.key === 'Escape') {
                setName(c.name)
                setEditing(false)
              }
            }}
            className="flex-1 min-w-0 h-5 px-1 rounded bg-input/30 border border-border/50 outline-none focus:border-border text-[11.5px]"
          />
        ) : (
          <button
            type="button"
            onClick={onSelect}
            onDoubleClick={() => setEditing(true)}
            className="flex-1 min-w-0 truncate text-left"
            title="Double-click to rename"
          >
            {c.name}
          </button>
        )}
        {run && !editing && (
          <span className="text-[9.5px] font-mono text-muted-foreground tabular-nums shrink-0">
            {run.passCount}/{(run.passCount ?? 0) + (run.failCount ?? 0) + (run.skipCount ?? 0)}
          </span>
        )}
        {!editing && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              void remove()
            }}
            title="Delete collection"
            className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive shrink-0"
          >
            <Trash2 className="w-3 h-3" />
          </button>
        )}
      </div>
    </li>
  )
}

interface CollectionDetailProps {
  projectId: string
  collection: Collection
  endpoints: { id: string; method: string; path: string }[]
  run?: CollectionRun | null
  isRunning: boolean
  progress?: { items: any[]; total: number } | null
  onRun: () => void
  onChanged: () => void
}

function CollectionDetail({
  projectId,
  collection,
  endpoints,
  run,
  isRunning,
  progress,
  onRun,
  onChanged,
}: CollectionDetailProps) {
  const [name, setName] = useState(collection.name)
  const [description, setDescription] = useState(collection.description ?? '')
  const [items, setItems] = useState(collection.items ?? [])
  const [picker, setPicker] = useState(false)
  const [datasetFor, setDatasetFor] = useState<{ index: number; endpointId: string; method: string; path: string } | null>(null)
  const { getMethodColor } = useHttpMethod()

  useEffect(() => {
    setName(collection.name)
    setDescription(collection.description ?? '')
    setItems(collection.items ?? [])
  }, [collection.id])

  const dirty =
    name !== collection.name ||
    description !== (collection.description ?? '') ||
    JSON.stringify(items) !== JSON.stringify(collection.items ?? [])

  const save = async () => {
    await collectionsService.save({
      id: collection.id,
      projectID: projectId,
      name,
      description,
      items,
    })
    onChanged()
  }

  useEffect(() => {
    if (!dirty) return
    const t = setTimeout(() => void save(), 600)
    return () => clearTimeout(t)
  }, [dirty, name, description, items])

  const remove = async () => {
    await collectionsService.remove(collection.id)
    onChanged()
  }

  const addItems = (endpointIds: string[]) => {
    if (endpointIds.length === 0) return
    setItems((prev) => [
      ...prev,
      ...endpointIds.map((id) => ({ endpointID: id }) as any),
    ])
    setPicker(false)
  }

  const removeItem = (idx: number) => setItems((prev) => prev.filter((_, i) => i !== idx))

  const toggleIterate = (idx: number) => {
    setItems((prev) => prev.map((it, i) => (i === idx ? { ...it, iterateDataset: !it.iterateDataset } : it)))
  }

  const move = (from: number, to: number) => {
    setItems((prev) => {
      const next = [...prev]
      const [m] = next.splice(from, 1)
      next.splice(to, 0, m)
      return next
    })
  }

  const endpointMap = useMemo(() => {
    const m = new Map<string, { method: string; path: string }>()
    for (const e of endpoints) m.set(e.id, { method: e.method, path: e.path })
    return m
  }, [endpoints])

  return (
    <>
      <div className="h-9 px-3 flex items-center justify-end gap-2 border-b border-border/40 shrink-0">
        {dirty && <span className="text-[10px] text-muted-foreground/70 italic">Saving…</span>}
        <Button size="sm" className="h-7 gap-1.5" onClick={onRun} disabled={isRunning || items.length === 0}>
          {isRunning ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3 fill-current" />}
          {isRunning ? 'Running…' : 'Run collection'}
        </Button>
      </div>

      <div className="flex-1 grid grid-cols-2 min-h-0">
        <section className="border-r border-border/40 flex flex-col min-h-0">
          <div className="h-8 px-3 flex items-center shrink-0 bg-muted/30 border-b border-border/30 gap-1.5">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              Items
            </span>
            <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{items.length}</span>
          </div>
          {items.length === 0 ? (
            <div className="flex-1 flex flex-col items-center justify-center gap-2 px-6">
              <p className="text-[11.5px] italic text-muted-foreground/70">No items yet.</p>
              <button
                type="button"
                onClick={() => setPicker(true)}
                className="group h-8 inline-flex items-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[12px] font-medium transition-colors px-3"
              >
                <Plus className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                <span className="ml-2">Add request</span>
              </button>
            </div>
          ) : (
            <ul className="m-0 p-0 list-none flex-1 overflow-y-auto">
              {items.map((it, idx) => {
                const ep = endpointMap.get(it.endpointID)
                return (
                  <li
                    key={`${it.endpointID}-${idx}`}
                    className="group flex items-center gap-2 px-3 py-1.5 hover:bg-accent/30 border-b border-border/20"
                  >
                    <button
                      type="button"
                      title="Move up"
                      onClick={() => idx > 0 && move(idx, idx - 1)}
                      className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-foreground"
                    >
                      <GripVertical className="w-3 h-3" />
                    </button>
                    <span
                      className={cn(
                        'inline-flex w-12 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                        ep ? getMethodColor(ep.method) : 'bg-muted/40 text-muted-foreground',
                      )}
                    >
                      {ep?.method ?? '—'}
                    </span>
                    <span className="text-[11px] font-mono text-foreground/85 truncate flex-1">
                      {ep?.path ?? <span className="italic text-muted-foreground">missing endpoint</span>}
                    </span>
                    {ep && BODY_METHODS.has(ep.method.toUpperCase()) && (
                      <button
                        type="button"
                        onClick={() => setDatasetFor({ index: idx, endpointId: it.endpointID, method: ep.method, path: ep.path })}
                        title={it.iterateDataset ? 'Dataset active — configure' : 'Configure dataset'}
                        className={cn(
                          'inline-flex items-center gap-1 h-5 px-1.5 rounded text-[9.5px] font-mono transition-colors shrink-0',
                          it.iterateDataset
                            ? 'bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/20'
                            : 'opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-foreground hover:bg-accent/60 border border-border/40',
                        )}
                      >
                        <Database className="w-3 h-3" />
                        dataset
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => removeItem(idx)}
                      className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
          {items.length > 0 && (
            <div className="px-3 py-2 border-t border-border/40 shrink-0">
              <button
                type="button"
                onClick={() => setPicker(true)}
                className="group w-full h-8 inline-flex items-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[12px] font-medium transition-colors px-2.5"
              >
                <Plus className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                <span className="ml-2">Add request</span>
              </button>
            </div>
          )}
          {datasetFor && (
            <DatasetDialog
              projectId={projectId}
              endpointId={datasetFor.endpointId}
              method={datasetFor.method}
              path={datasetFor.path}
              iterating={!!items[datasetFor.index]?.iterateDataset}
              onToggleIterate={() => toggleIterate(datasetFor.index)}
              onClose={() => setDatasetFor(null)}
            />
          )}
          {picker && (
            <EndpointPicker
              endpoints={endpoints}
              onPick={addItems}
              onClose={() => setPicker(false)}
            />
          )}
        </section>

        <section className="flex flex-col min-h-0">
          <div className="h-8 px-3 flex items-center justify-between shrink-0 bg-muted/30 border-b border-border/30">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              {isRunning ? 'Running' : 'Last run'}
            </span>
            {isRunning && progress ? (
              <span className="text-[10px] font-mono text-muted-foreground tabular-nums">
                {progress.items.length}/{progress.total || items.length}
              </span>
            ) : run ? (
              <span className="text-[10px] font-mono text-muted-foreground">
                {run.passCount} passed · {run.failCount} failed · {run.skipCount} skipped · {run.durationMs}ms
              </span>
            ) : null}
          </div>
          {(() => {
            const live = isRunning && progress ? progress.items : null
            const display = live ?? run?.items ?? null
            if (!display) {
              return (
                <p className="px-3 py-6 text-[11px] italic text-muted-foreground/70 text-center">
                  Click Run collection to execute its requests in order.
                </p>
              )
            }
            return (
              <ul className="m-0 p-0 list-none flex-1 overflow-y-auto">
                {display.map((r: any, idx: number) => (
                  <RunRow key={idx} item={r} projectId={projectId} getMethodColor={getMethodColor} />
                ))}
                {isRunning && display.length < (progress?.total ?? items.length) && (
                  <li className="flex items-center gap-2 px-3 py-1.5 text-[10.5px] text-muted-foreground italic">
                    <Loader2 className="w-3 h-3 animate-spin" />
                    Running…
                  </li>
                )}
              </ul>
            )
          })()}
        </section>
      </div>
    </>
  )
}

function RunRow({
  item: r,
  projectId,
  getMethodColor,
}: {
  item: any
  projectId: string
  getMethodColor: (m: string) => string
}) {
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
      <div
        className={cn(
          'w-full flex items-center gap-2 px-3 py-1.5',
          hasDetail && 'hover:bg-accent/30',
        )}
      >
      <button
        type="button"
        onClick={() => hasDetail && setOpen((o) => !o)}
        className={cn(
          'flex-1 min-w-0 flex items-center gap-2 text-left',
          hasDetail && 'cursor-pointer',
        )}
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
          {summary && (
            <p className="text-[10.5px] text-muted-foreground italic">{summary}</p>
          )}
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

function Switch({
  checked,
  onCheckedChange,
  label,
}: {
  checked: boolean
  onCheckedChange: (v: boolean) => void
  label?: string
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onCheckedChange(!checked)}
      className="inline-flex items-center gap-2 group"
    >
      <span
        className={cn(
          'relative inline-flex h-4 w-7 shrink-0 items-center rounded-full transition-colors',
          checked ? 'bg-emerald-500' : 'bg-muted',
        )}
      >
        <span
          className={cn(
            'inline-block h-3 w-3 transform rounded-full bg-background shadow transition-transform',
            checked ? 'translate-x-3.5' : 'translate-x-0.5',
          )}
        />
      </span>
      {label && (
        <span className={cn('text-[10.5px]', checked ? 'text-emerald-500' : 'text-muted-foreground')}>
          {label}
        </span>
      )}
    </button>
  )
}

function DatasetDialog({
  projectId,
  endpointId,
  method,
  path,
  iterating,
  onToggleIterate,
  onClose,
}: {
  projectId: string
  endpointId: string
  method: string
  path: string
  iterating: boolean
  onToggleIterate: () => void
  onClose: () => void
}) {
  const endpointKey = `${method.toUpperCase()} ${path}`
  const [rows, setRows] = useState<unknown[]>([])
  const [count, setCount] = useState(10)
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)

  useEffect(() => {
    setLoading(true)
    void datasetsService.get(projectId, endpointKey).then((r) => {
      setRows(r)
      if (r.length > 0) setCount(r.length)
      setLoading(false)
    })
  }, [projectId, endpointKey])

  const persist = async (next: unknown[]) => {
    setRows(next)
    await datasetsService.save(projectId, endpointKey, next)
  }

  const generate = async () => {
    setGenerating(true)
    try {
      const next = await datasetsService.generate(endpointId, count)
      await persist(next)
    } finally {
      setGenerating(false)
    }
  }

  const removeRow = async (i: number) => {
    await persist(rows.filter((_, idx) => idx !== i))
  }

  const updateRow = async (i: number, value: string) => {
    try {
      const parsed = JSON.parse(value)
      await persist(rows.map((r, idx) => (idx === i ? parsed : r)))
    } catch {
      // ignore parse errors
    }
  }

  const [expanded, setExpanded] = useState<number | null>(null)
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
          <DialogTitle className="text-base">Dataset</DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Run this request multiple times with different payloads generated from the schema.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4 space-y-3">
          <div className="rounded-md border border-border/60 bg-card/40 p-3 flex items-center gap-3">
            <div className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-emerald-500/10 text-emerald-500 shrink-0">
              <Database className="w-4 h-4" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">Endpoint</p>
              <code className="text-[11.5px] font-mono text-foreground/85 truncate block">{endpointKey}</code>
            </div>
            <div className="flex flex-col items-end gap-1 shrink-0">
              <Switch checked={iterating} onCheckedChange={onToggleIterate} />
              <span className={cn('text-[10px] font-medium', iterating ? 'text-emerald-500' : 'text-muted-foreground')}>
                {iterating ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>

          <div className="rounded-md border border-border/60 bg-card/40">
            <div className="px-3 py-2 border-b border-border/40 flex items-center gap-2">
              <Sparkles className="w-3 h-3 text-muted-foreground" />
              <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">Payloads</span>
              <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{rows.length}</span>
              <div className="ml-auto flex items-center gap-1.5">
                <Input
                  type="number"
                  value={count}
                  min={1}
                  max={500}
                  onChange={(e) => setCount(Math.max(1, Math.min(500, Number(e.target.value) || 1)))}
                  className="h-7 w-14 text-[11px] font-mono text-center"
                />
                <Button
                  size="sm"
                  variant="outline"
                  className="h-7 px-2.5 text-[10.5px] gap-1.5"
                  onClick={generate}
                  disabled={generating}
                >
                  {generating ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3 text-emerald-500" />}
                  Generate
                </Button>
                {rows.length > 0 && (
                  <button
                    type="button"
                    onClick={() => void persist([])}
                    className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-destructive ml-1"
                  >
                    <Trash2 className="w-3 h-3" />
                    Clear
                  </button>
                )}
              </div>
            </div>
            {loading ? (
              <p className="px-4 py-8 text-[11px] italic text-muted-foreground/70 text-center">Loading…</p>
            ) : rows.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 px-6 py-8 text-center">
                <Sparkles className="w-5 h-5 text-muted-foreground/40" />
                <p className="text-[11.5px] text-muted-foreground/70">
                  No payloads yet. Set a count and click Generate.
                </p>
              </div>
            ) : (
              <ul className="m-0 p-0 list-none divide-y divide-border/20 max-h-96 overflow-y-auto">
                {rows.map((row, i) => (
                  <DatasetRow
                    key={i}
                    index={i}
                    row={row}
                    expanded={expanded === i}
                    onToggle={() => setExpanded((e) => (e === i ? null : i))}
                    onUpdate={(v) => void updateRow(i, v)}
                    onRemove={() => void removeRow(i)}
                  />
                ))}
              </ul>
            )}
          </div>
        </div>

        <DialogFooter className="px-6 py-3 shrink-0 border-t border-border/40">
          <Button variant="outline" size="sm" onClick={onClose}>
            Done
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function DatasetRow({
  index,
  row,
  expanded,
  onToggle,
  onUpdate,
  onRemove,
}: {
  index: number
  row: any
  expanded: boolean
  onToggle: () => void
  onUpdate: (v: string) => void
  onRemove: () => void
}) {
  const summary = useMemo(() => {
    if (!row || typeof row !== 'object') return JSON.stringify(row)
    const entries = Object.entries(row).slice(0, 3)
    return entries
      .map(([k, v]) => `${k}: ${typeof v === 'string' ? v : JSON.stringify(v)}`)
      .join(' · ')
  }, [row])

  return (
    <li className="group">
      <div
        className="flex items-center gap-2 px-3 py-1.5 hover:bg-accent/20 cursor-pointer"
        onClick={onToggle}
      >
        <ChevronRight
          className={cn(
            'w-3 h-3 text-muted-foreground/60 shrink-0 transition-transform',
            expanded && 'rotate-90',
          )}
        />
        <span className="text-[10px] font-mono text-muted-foreground tabular-nums w-6 shrink-0">
          #{index + 1}
        </span>
        <span className="text-[11px] font-mono text-foreground/75 truncate flex-1">{summary}</span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive shrink-0"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>
      {expanded && (
        <textarea
          defaultValue={JSON.stringify(row, null, 2)}
          onBlur={(e) => onUpdate(e.target.value)}
          spellCheck={false}
          className="w-full bg-input/20 text-[11px] font-mono px-3 py-2 outline-none resize-y min-h-[120px] max-h-72 border-t border-border/20"
        />
      )}
    </li>
  )
}

function EndpointPicker({
  endpoints,
  onPick,
  onClose,
}: {
  endpoints: { id: string; method: string; path: string }[]
  onPick: (ids: string[]) => void
  onClose: () => void
}) {
  const { getMethodColor } = useHttpMethod()
  const [query, setQuery] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const filtered = endpoints
    .filter((e) => `${e.method} ${e.path}`.toLowerCase().includes(query.toLowerCase()))
    .slice(0, 500)

  const toggle = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const allFilteredSelected =
    filtered.length > 0 && filtered.every((e) => selected.has(e.id))

  const toggleAll = () => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (allFilteredSelected) {
        for (const e of filtered) next.delete(e.id)
      } else {
        for (const e of filtered) next.add(e.id)
      }
      return next
    })
  }

  const confirm = () => {
    onPick(Array.from(selected))
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent
        className="p-0 gap-0 sm:max-w-xl w-[640px] flex flex-col overflow-hidden"
        style={{ height: '70vh' }}
      >
        <DialogTitle className="sr-only">Add requests</DialogTitle>
        <div className="px-3 py-2 border-b border-border/40 flex items-center gap-2 shrink-0">
          <Input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search endpoint…"
            className="h-8 text-[12px] border-0 bg-transparent focus-visible:ring-0 shadow-none"
          />
        </div>
        <div className="px-4 py-1.5 border-b border-border/30 flex items-center gap-2 shrink-0 bg-muted/20">
          <input
            type="checkbox"
            checked={allFilteredSelected}
            onChange={toggleAll}
            className="h-3.5 w-3.5 accent-primary cursor-pointer"
            aria-label="Select all"
          />
          <span className="text-[10px] text-muted-foreground">
            {allFilteredSelected ? 'Deselect all' : 'Select all'}
          </span>
          <span className="ml-auto text-[10px] font-mono text-muted-foreground/60 tabular-nums">
            {selected.size} selected · {filtered.length} of {endpoints.length}
          </span>
        </div>
        <ul className="m-0 p-0 list-none flex-1 overflow-y-auto">
          {filtered.length === 0 ? (
            <li className="px-3 py-8 text-[11px] italic text-muted-foreground/70 text-center">
              No matches.
            </li>
          ) : (
            filtered.map((e) => {
              const isSelected = selected.has(e.id)
              return (
                <li key={e.id}>
                  <label className="w-full flex items-center gap-2.5 px-4 py-1.5 text-[11.5px] hover:bg-accent/40 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggle(e.id)}
                      className="h-3.5 w-3.5 accent-primary cursor-pointer"
                    />
                    <span
                      className={cn(
                        'inline-flex w-12 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                        getMethodColor(e.method),
                      )}
                    >
                      {e.method}
                    </span>
                    <span className="font-mono text-foreground/85 truncate">{e.path}</span>
                  </label>
                </li>
              )
            })
          )}
        </ul>
        <div className="px-3 py-2 border-t border-border/40 flex items-center justify-end gap-2 shrink-0">
          <Button size="sm" variant="ghost" className="h-7 text-[11px]" onClick={onClose}>
            Cancel
          </Button>
          <Button
            size="sm"
            className="h-7 text-[11px]"
            onClick={confirm}
            disabled={selected.size === 0}
          >
            Add {selected.size > 0 ? `${selected.size} ` : ''}request{selected.size === 1 ? '' : 's'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
