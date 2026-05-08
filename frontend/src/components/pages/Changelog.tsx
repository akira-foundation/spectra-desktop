import { useEffect, useState } from 'react'
import { Plus, Minus, Pencil, ChevronRight, History } from 'lucide-react'
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

  const added = diff.added ?? []
  const removed = diff.removed ?? []
  const changed = diff.changed ?? []
  const empty = added.length === 0 && removed.length === 0 && changed.length === 0

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="px-4 py-2.5 border-b border-border/40">
        <div className="flex items-center gap-3 text-[11.5px]">
          <span className="text-muted-foreground">Snapshot</span>
          <span className="font-mono tabular-nums">{formatDate(new Date(diff.scannedAt))}</span>
          <div className="flex items-center gap-2 ml-auto text-[10.5px] font-mono">
            {added.length > 0 && (
              <span className="text-emerald-500">+{added.length}</span>
            )}
            {removed.length > 0 && (
              <span className="text-rose-500">-{removed.length}</span>
            )}
            {changed.length > 0 && (
              <span className="text-amber-500">~{changed.length}</span>
            )}
          </div>
        </div>
      </div>
      <div className="flex-1 overflow-auto">
        {empty ? (
          <p className="p-6 text-[12px] text-muted-foreground italic text-center">
            No changes since the previous snapshot.
          </p>
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
  const bgClass = {
    emerald: 'hover:bg-emerald-500/5',
    rose: 'hover:bg-rose-500/5',
    amber: 'hover:bg-amber-500/5',
  }[tone]
  return (
    <li className={cn('flex items-center gap-2 px-2 py-1 rounded-md transition-colors', bgClass)}>
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
    </li>
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
