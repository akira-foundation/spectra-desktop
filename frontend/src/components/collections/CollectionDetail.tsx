import { useEffect, useMemo, useState } from 'react'
import {
  Plus,
  Play,
  X,
  Loader2,
  GripVertical,
  Database,
  Download,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { collectionsService, type Collection, type CollectionRun } from '@/services/collectionsService'
import { cn } from '@/lib/utils'
import { DatasetDialog } from './DatasetDialog'
import { EndpointPicker } from './EndpointPicker'
import { RunRow } from './RunRow'

const BODY_METHODS = new Set(['POST', 'PUT', 'PATCH'])

interface Props {
  projectId: string
  collection: Collection
  endpoints: { id: string; method: string; path: string }[]
  run?: CollectionRun | null
  isRunning: boolean
  progress?: { items: any[]; total: number } | null
  onRun: () => void
  onChanged: () => void
}

export function CollectionDetail({
  projectId,
  collection,
  endpoints,
  run,
  isRunning,
  progress,
  onRun,
  onChanged,
}: Props) {
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

  const addItems = (endpointIds: string[]) => {
    if (endpointIds.length === 0) return
    setItems((prev) => [...prev, ...endpointIds.map((id) => ({ endpointID: id }) as any)])
    setPicker(false)
  }

  const removeItem = (idx: number) => setItems((prev) => prev.filter((_, i) => i !== idx))

  const toggleIterate = (idx: number) => {
    setItems((prev) =>
      prev.map((it, i) => (i === idx ? { ...it, iterateDataset: !it.iterateDataset } : it)),
    )
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
        <Button
          size="sm"
          variant="outline"
          className="h-7 px-3 gap-1.5 text-[11.5px]"
          onClick={async () => {
            try {
              await collectionsService.exportToFile(collection.id)
            } catch (err) {
              alert(`Export failed: ${(err as Error).message ?? err}`)
            }
          }}
          title="Export as JSON"
        >
          <Download className="w-3 h-3" />
          Export
        </Button>
        <Button
          size="sm"
          className="h-7 px-3 gap-1.5 text-[11.5px]"
          onClick={onRun}
          disabled={isRunning || items.length === 0}
        >
          {isRunning ? (
            <Loader2 className="w-3 h-3 animate-spin" />
          ) : (
            <Play className="w-3 h-3 fill-current" />
          )}
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
                        onClick={() =>
                          setDatasetFor({
                            index: idx,
                            endpointId: it.endpointID,
                            method: ep.method,
                            path: ep.path,
                          })
                        }
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
            <EndpointPicker endpoints={endpoints} onPick={addItems} onClose={() => setPicker(false)} />
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
