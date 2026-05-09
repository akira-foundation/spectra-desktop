import { create } from 'zustand'
import { collectionsService, type Collection, type CollectionRun } from '@/services/collectionsService'

interface CollectionsState {
  byProject: Record<string, Collection[]>
  lastRun: Record<string, CollectionRun | null>
  load: (projectId: string) => Promise<void>
  refresh: (projectId: string) => Promise<void>
  run: (projectId: string, collectionId: string) => Promise<CollectionRun | null>
  list: (projectId: string | null) => Collection[]
}

export const useCollectionsStore = create<CollectionsState>((set, get) => ({
  byProject: {},
  lastRun: {},

  load: async (projectId) => {
    if (!projectId) return
    if (get().byProject[projectId]) return
    await get().refresh(projectId)
  },

  refresh: async (projectId) => {
    if (!projectId) return
    try {
      const rows = await collectionsService.list(projectId)
      set((s) => ({ byProject: { ...s.byProject, [projectId]: rows } }))
    } catch (err) {
      console.error('collections refresh failed:', err)
    }
  },

  run: async (projectId, collectionId) => {
    const result = await collectionsService.run(collectionId)
    set((s) => ({ lastRun: { ...s.lastRun, [collectionId]: result } }))
    if (projectId) await get().refresh(projectId)
    return result
  },

  list: (projectId) => {
    if (!projectId) return []
    return get().byProject[projectId] ?? []
  },
}))
