import { useEffect, useMemo, useState } from 'react'
import { Plus, FolderKanban, Upload } from 'lucide-react'
import { useCollectionsStore } from '@/store/collectionsStore'
import { useHistoryStore } from '@/store/historyStore'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { collectionsService, type Collection } from '@/services/collectionsService'
import { SidebarItem, CollectionDetail } from '@/components/collections'

const EMPTY_COLLECTIONS: Collection[] = []
const EMPTY_ENDPOINTS: any[] = []

export function CollectionsPage() {
  const projectId = useProjectStore((s) => s.activeProjectId)
  const collections = useCollectionsStore((s) =>
    projectId ? s.byProject[projectId] ?? EMPTY_COLLECTIONS : EMPTY_COLLECTIONS,
  )
  const lastRun = useCollectionsStore((s) => s.lastRun)
  const refresh = useCollectionsStore((s) => s.refresh)
  const runCol = useCollectionsStore((s) => s.run)
  const endpoints = useEndpointsStore((s) =>
    projectId ? s.byProject[projectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
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

  const importFile = () => {
    if (!projectId) return
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json,application/json'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) return
      const text = await file.text()
      try {
        const result = await collectionsService.import(projectId, text)
        await refresh(projectId)
        if (result.collection?.id) setActiveId(result.collection.id)
        if (result.missingEndpoints.length > 0) {
          alert(
            `Imported with ${result.missingEndpoints.length} missing endpoint(s):\n${result.missingEndpoints.join('\n')}`,
          )
        }
      } catch (err) {
        alert(`Import failed: ${(err as Error).message}`)
      }
    }
    input.click()
  }

  const handleRun = async (id: string) => {
    if (!projectId) return
    setRunning(id)
    setProgress({ items: [], total: 0 })
    try {
      await runCol(projectId, id)
      await useHistoryStore.getState().refresh(projectId)
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
        <div className="px-3 py-2 border-t border-border/40 shrink-0">
          <button
            type="button"
            onClick={importFile}
            className="group w-full h-8 inline-flex items-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[12px] font-medium transition-colors px-2.5"
            title="Import collection JSON"
          >
            <Upload className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
            <span className="ml-2">Import collection</span>
          </button>
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
