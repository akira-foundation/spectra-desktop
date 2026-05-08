import { useEffect, useState } from 'react'
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { Plus, Minus, Pencil, ChevronRight, History, Search, X } from 'lucide-react'
import { useProjectStore } from '@/store/projectStore'
import { useChangelogStore } from '@/store/changelogStore'
import { Welcome } from '@/components/pages/Welcome'
import { Skeleton } from '@/components/ui/skeleton'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import type { SnapshotSummary, SnapshotDiff, SnapshotDiffEntry } from '@/services/changelogService'
import { cn } from '@/lib/utils'

const EMPTY_SNAPSHOTS: SnapshotSummary[] = []

export function Changelog() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const project = useProjectStore((s) => s.projects.find((p) => p.id === activeProjectId))
  const snapshots = useChangelogStore((s) =>
    activeProjectId ? s.snapshotsByProject[activeProjectId] ?? EMPTY_SNAPSHOTS : EMPTY_SNAPSHOTS,
  )
  const loading = useChangelogStore((s) => (activeProjectId ? s.loading[activeProjectId] : false))
  const load = useChangelogStore((s) => s.load)
  const [selected, setSelected] = useState<string | null>(null)

  useEffect(() => {
    if (activeProjectId) void load(activeProjectId)
  }, [activeProjectId, load])

  useEffect(() => {
    setSelected(snapshots[0]?.id ?? null)
  }, [snapshots])

  if (!project) return <Welcome />

  return (
    <div className="h-full flex gap-2 p-2 overflow-hidden">
      <div className="w-72 shrink-0 flex flex-col rounded-md border border-border/40 bg-foreground/[0.025] dark:bg-white/[0.02] overflow-hidden">
        <div className="px-3 py-2 border-b border-border/40 flex items-center gap-2">
          <History className="w-3.5 h-3.5 text-muted-foreground" />
          <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Snapshots
          </span>
          <span className="ml-auto text-[10px] font-mono text-muted-foreground/70">
            {snapshots.length}
          </span>
        </div>
        <div className="flex-1 overflow-auto">
          {loading && snapshots.length === 0 ? (
            <div className="p-2 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : snapshots.length === 0 ? (
            <p className="p-4 text-[11.5px] text-muted-foreground italic text-center">
              No scans yet. Run Sync to capture the first snapshot.
            </p>
          ) : (
            <ul className="p-1 space-y-px">
              {snapshots.map((snap) => (
                <li key={snap.id}>
                  <button
                    type="button"
                    onClick={() => setSelected(snap.id)}
                    className={cn(
                      'w-full text-left px-2.5 py-2 rounded-md hover:bg-accent/40 transition-colors',
                      selected === snap.id && 'bg-accent text-foreground',
                    )}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-[11.5px] font-medium">
                        {formatDate(new Date(snap.scannedAt))}
                      </span>
                      <ChevronRight className="w-3 h-3 text-muted-foreground" />
                    </div>
                    <div className="flex items-center gap-2 mt-1 text-[10.5px] font-mono">
                      <span className="text-muted-foreground/80">{snap.endpointCount}</span>
                      {snap.added > 0 && (
                        <span className="text-emerald-500">+{snap.added}</span>
                      )}
                      {snap.removed > 0 && (
                        <span className="text-rose-500">-{snap.removed}</span>
                      )}
                      {snap.changed > 0 && (
                        <span className="text-amber-500">~{snap.changed}</span>
                      )}
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <div className="flex-1 rounded-md border border-border/40 bg-card/30 overflow-hidden flex flex-col">
        {selected ? <DiffView snapshotId={selected} /> : <EmptyDiff />}
      </div>
    </div>
  )
}

interface DiffViewProps {
  snapshotId: string
}

function DiffView({ snapshotId }: DiffViewProps) {
  const loadDiff = useChangelogStore((s) => s.loadDiff)
  const diff = useChangelogStore((s) => s.diffsByID[snapshotId])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')

  useEffect(() => {
    setQuery('')
  }, [snapshotId])

  useEffect(() => {
    if (diff) return
    setLoading(true)
    loadDiff(snapshotId).finally(() => setLoading(false))
  }, [snapshotId, diff, loadDiff])

  if (loading || !diff) {
    return (
      <div className="p-4 space-y-2">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-4 w-2/3" />
      </div>
    )
  }

  const filterEntries = (entries: SnapshotDiffEntry[]) => {
    const q = query.trim().toLowerCase()
    if (!q) return entries
    return entries.filter(
      (e) =>
        e.method.toLowerCase().includes(q) ||
        e.path.toLowerCase().includes(q) ||
        (e.changes ?? []).some((c) => c.toLowerCase().includes(q)),
    )
  }
  const added = filterEntries(diff.added ?? [])
  const removed = filterEntries(diff.removed ?? [])
  const changed = filterEntries(diff.changed ?? [])
  const totalRaw =
    (diff.added?.length ?? 0) + (diff.removed?.length ?? 0) + (diff.changed?.length ?? 0)
  const empty = added.length === 0 && removed.length === 0 && changed.length === 0
  const isFirst = !diff.previousID

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="px-4 py-2.5 border-b border-border/40 space-y-2">
        <div className="flex items-center gap-3 text-[11.5px]">
          <span className="text-muted-foreground">Snapshot</span>
          <span className="font-mono tabular-nums">{formatDate(new Date(diff.scannedAt))}</span>
          <div className="flex items-center gap-2 ml-auto text-[10.5px] font-mono">
            {(diff.added?.length ?? 0) > 0 && (
              <span className="text-emerald-500">+{diff.added?.length}</span>
            )}
            {(diff.removed?.length ?? 0) > 0 && (
              <span className="text-rose-500">-{diff.removed?.length}</span>
            )}
            {(diff.changed?.length ?? 0) > 0 && (
              <span className="text-amber-500">~{diff.changed?.length}</span>
            )}
          </div>
        </div>
        {totalRaw > 0 && (
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Filter by method, path, or change kind"
              className="w-full h-7 pl-7 pr-7 text-[12px] bg-input/60 border border-border/50 rounded-md focus:outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/70"
            />
            {query && (
              <button
                type="button"
                onClick={() => setQuery('')}
                aria-label="Clear filter"
                className="absolute right-1.5 top-1/2 -translate-y-1/2 inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/50"
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </div>
        )}
      </div>
      <div className="flex-1 overflow-auto">
        {empty ? (
          <div className="p-8 text-center space-y-1">
            <p className="text-[12.5px] font-medium text-foreground/85">
              {query
                ? 'No matches'
                : isFirst
                  ? 'First snapshot'
                  : 'No changes'}
            </p>
            <p className="text-[11.5px] text-muted-foreground">
              {query
                ? 'Try a different filter.'
                : isFirst
                  ? 'No prior snapshot to compare against.'
                  : 'API definition matches the previous scan.'}
            </p>
          </div>
        ) : (
          <div className="p-3 space-y-4">
            {added.length > 0 && (
              <Section title="Added" tone="emerald" icon={Plus} entries={added} />
            )}
            {removed.length > 0 && (
              <Section title="Removed" tone="rose" icon={Minus} entries={removed} />
            )}
            {changed.length > 0 && (
              <Section title="Changed" tone="amber" icon={Pencil} entries={changed} />
            )}
          </div>
        )}
      </div>
    </div>
  )
}

interface SectionProps {
  title: string
  tone: 'emerald' | 'rose' | 'amber'
  icon: React.ComponentType<{ className?: string }>
  entries: SnapshotDiffEntry[]
}

function Section({ title, tone, icon: Icon, entries }: SectionProps) {
  const toneClass = {
    emerald: 'text-emerald-500',
    rose: 'text-rose-500',
    amber: 'text-amber-500',
  }[tone]
  return (
    <section className="space-y-1.5">
      <h3 className={cn('flex items-center gap-1.5 text-[10.5px] font-semibold uppercase tracking-wider', toneClass)}>
        <Icon className="w-3 h-3" />
        {title}
        <span className="text-muted-foreground/70 font-mono ml-1">{entries.length}</span>
      </h3>
      <ul className="space-y-px">
        {entries.map((entry, i) => (
          <DiffRow key={`${entry.method}-${entry.path}-${i}`} entry={entry} tone={tone} />
        ))}
      </ul>
    </section>
  )
}

function DiffRow({ entry, tone }: { entry: SnapshotDiffEntry; tone: 'emerald' | 'rose' | 'amber' }) {
  const { getMethodColor } = useHttpMethod()
  const [open, setOpen] = useState(false)
  const bgClass = {
    emerald: 'hover:bg-emerald-500/5',
    rose: 'hover:bg-rose-500/5',
    amber: 'hover:bg-amber-500/5',
  }[tone]
  const expandable = entry.kind === 'changed' || entry.kind === 'added' || entry.kind === 'removed'
  return (
    <li className={cn('rounded-md transition-colors', bgClass)}>
      <button
        type="button"
        onClick={() => expandable && setOpen((v) => !v)}
        className="w-full flex items-center gap-2 px-2 py-1 text-left"
      >
        <span
          className={cn(
            'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
            getMethodColor(entry.method),
          )}
        >
          {entry.method}
        </span>
        <span className="text-[11.5px] font-mono truncate flex-1">{entry.path}</span>
        {entry.changes && entry.changes.length > 0 && (
          <div className="flex items-center gap-1 shrink-0">
            {entry.changes.map((c) => (
              <span
                key={c}
                className="text-[9.5px] uppercase tracking-wider px-1.5 py-0.5 rounded bg-muted/40 text-muted-foreground"
              >
                {c}
              </span>
            ))}
          </div>
        )}
      </button>
      {open && expandable && <DiffDetail entry={entry} />}
    </li>
  )
}

function DiffDetail({ entry }: { entry: SnapshotDiffEntry }) {
  const prev = entry.previous
  const cur = entry.current
  const changes = entry.changes ?? []

  return (
    <div className="px-3 pb-2 space-y-2">
      {entry.kind === 'added' && cur && (
        <div className="rounded-md border border-emerald-500/30 bg-emerald-500/5 p-2 space-y-1">
          <p className="text-[10px] uppercase tracking-wider text-emerald-500">Added</p>
          <EndpointInfo data={cur} />
        </div>
      )}
      {entry.kind === 'removed' && prev && (
        <div className="rounded-md border border-rose-500/30 bg-rose-500/5 p-2 space-y-1">
          <p className="text-[10px] uppercase tracking-wider text-rose-500">Removed</p>
          <EndpointInfo data={prev} />
        </div>
      )}
      {entry.kind === 'changed' && prev && cur && (
        <>
          {changes.includes('handler') && (
            <FieldDiff
              label="Handler"
              before={prev.handler ?? ''}
              after={cur.handler ?? ''}
            />
          )}
          {changes.includes('authRole') && (
            <FieldDiff
              label="Auth role"
              before={prev.authRole || '—'}
              after={cur.authRole || '—'}
            />
          )}
          {changes.includes('middleware') && (
            <ListDiff
              label="Middleware"
              before={prev.middleware ?? []}
              after={cur.middleware ?? []}
            />
          )}
          {changes.includes('schema') && (
            <SchemaFieldsDiff
              before={prev.schemaFields ?? []}
              after={cur.schemaFields ?? []}
            />
          )}
        </>
      )}
    </div>
  )
}

function EndpointInfo({ data }: { data: NonNullable<SnapshotDiffEntry['current']> }) {
  return (
    <dl className="grid grid-cols-[80px_1fr] gap-x-2 gap-y-0.5 text-[11px]">
      {data.handler && (
        <>
          <dt className="text-muted-foreground">Handler</dt>
          <dd className="font-mono break-all">{data.handler}</dd>
        </>
      )}
      {data.authRole && (
        <>
          <dt className="text-muted-foreground">Auth role</dt>
          <dd className="font-mono">{data.authRole}</dd>
        </>
      )}
      {data.middleware && data.middleware.length > 0 && (
        <>
          <dt className="text-muted-foreground">Middleware</dt>
          <dd className="font-mono break-all">{data.middleware.join(', ')}</dd>
        </>
      )}
      {data.schemaFields && data.schemaFields.length > 0 && (
        <>
          <dt className="text-muted-foreground">Fields</dt>
          <dd className="font-mono">
            {data.schemaFields.map((f) => `${f.name}:${f.type}${f.required ? '*' : ''}`).join(', ')}
          </dd>
        </>
      )}
    </dl>
  )
}

function FieldDiff({ label, before, after }: { label: string; before: string; after: string }) {
  return (
    <div className="rounded-md border border-border/40 bg-muted/20 p-2 text-[11px]">
      <p className="text-[9.5px] uppercase tracking-wider text-muted-foreground mb-1">{label}</p>
      <div className="grid grid-cols-2 gap-2">
        <div className="text-rose-400 font-mono break-all">- {before || '—'}</div>
        <div className="text-emerald-400 font-mono break-all">+ {after || '—'}</div>
      </div>
    </div>
  )
}

function ListDiff({ label, before, after }: { label: string; before: string[]; after: string[] }) {
  const beforeSet = new Set(before)
  const afterSet = new Set(after)
  const added = after.filter((v) => !beforeSet.has(v))
  const removed = before.filter((v) => !afterSet.has(v))
  return (
    <div className="rounded-md border border-border/40 bg-muted/20 p-2 text-[11px]">
      <p className="text-[9.5px] uppercase tracking-wider text-muted-foreground mb-1">{label}</p>
      <div className="space-y-0.5">
        {added.map((v) => (
          <div key={`a-${v}`} className="text-emerald-400 font-mono break-all">
            + {v}
          </div>
        ))}
        {removed.map((v) => (
          <div key={`r-${v}`} className="text-rose-400 font-mono break-all">
            − {v}
          </div>
        ))}
      </div>
    </div>
  )
}

function SchemaFieldsDiff({
  before,
  after,
}: {
  before: NonNullable<SnapshotDiffEntry['previous']>['schemaFields']
  after: NonNullable<SnapshotDiffEntry['current']>['schemaFields']
}) {
  const prev = before ?? []
  const next = after ?? []
  const prevByName = new Map(prev.map((f) => [f.name, f]))
  const nextByName = new Map(next.map((f) => [f.name, f]))
  const added = next.filter((f) => !prevByName.has(f.name))
  const removed = prev.filter((f) => !nextByName.has(f.name))
  const changed = next.filter((f) => {
    const old = prevByName.get(f.name)
    if (!old) return false
    return old.type !== f.type || !!old.required !== !!f.required
  })
  return (
    <div className="rounded-md border border-border/40 bg-muted/20 p-2 text-[11px]">
      <p className="text-[9.5px] uppercase tracking-wider text-muted-foreground mb-1">
        Schema fields
      </p>
      <div className="space-y-0.5">
        {added.map((f) => (
          <div key={`a-${f.name}`} className="text-emerald-400 font-mono">
            + {f.name}: {f.type}
            {f.required ? '*' : ''}
          </div>
        ))}
        {removed.map((f) => (
          <div key={`r-${f.name}`} className="text-rose-400 font-mono">
            − {f.name}: {f.type}
            {f.required ? '*' : ''}
          </div>
        ))}
        {changed.map((f) => {
          const old = prevByName.get(f.name)!
          return (
            <div key={`c-${f.name}`} className="text-amber-400 font-mono">
              ~ {f.name}: {old.type}
              {old.required ? '*' : ''} → {f.type}
              {f.required ? '*' : ''}
            </div>
          )
        })}
        {added.length === 0 && removed.length === 0 && changed.length === 0 && (
          <p className="italic text-muted-foreground">Examples differ but field shape unchanged.</p>
        )}
      </div>
    </div>
  )
}

function EmptyDiff() {
  return (
    <div className="h-full flex items-center justify-center text-[12px] text-muted-foreground italic">
      Select a snapshot to view its diff.
    </div>
  )
}

function formatDate(date: Date): string {
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return `Today · ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  }
  return date.toLocaleString([], {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
