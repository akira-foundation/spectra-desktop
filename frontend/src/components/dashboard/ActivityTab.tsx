import { useMemo } from 'react'
import type { HistoryListItem } from '@/services/historyService'
import { useProjectStore } from '@/store/projectStore'
import { useCollectionsStore } from '@/store/collectionsStore'
import {
  RecentActivityCard,
  RecentFailuresCard,
  CollectionRunsCard,
  SnapshotCard,
} from './cards'

interface Props {
  history: HistoryListItem[]
  recent: HistoryListItem[]
  latestSnapshot: any
  onOpen: (id?: string) => void
  onOpenSnapshots: () => void
  onOpenCollections: () => void
}

export function ActivityTab({
  history,
  recent,
  latestSnapshot,
  onOpen,
  onOpenSnapshots,
  onOpenCollections,
}: Props) {
  const failures = useMemo(
    () => history.filter((h) => h.error || h.responseStatus >= 400).slice(0, 8),
    [history],
  )
  const projectId = useProjectStore((s) => s.activeProjectId)
  const collections = useCollectionsStore((s) => (projectId ? s.byProject[projectId] : undefined))
  const lastRun = useCollectionsStore((s) => s.lastRun)
  const recentRuns = useMemo(() => {
    return (collections ?? [])
      .map((c) => ({ collection: c, run: lastRun[c.id] }))
      .filter((x) => x.run)
      .sort((a, b) => (b.run?.startedAt ?? 0) - (a.run?.startedAt ?? 0))
      .slice(0, 5)
  }, [collections, lastRun])

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <RecentActivityCard entries={recent} onOpen={onOpen} />
        <RecentFailuresCard entries={failures} onOpen={onOpen} />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <CollectionRunsCard runs={recentRuns} onOpen={onOpenCollections} />
        <SnapshotCard snapshot={latestSnapshot} onOpen={onOpenSnapshots} />
      </div>
    </div>
  )
}
