import { create } from 'zustand'
import { historyService, type HistoryListItem } from '@/services/historyService'

interface HistoryState {
  byProject: Record<string, HistoryListItem[]>
  loading: Record<string, boolean>

  load: (projectId: string) => Promise<void>
  refresh: (projectId: string) => Promise<void>
  clear: (projectId: string) => Promise<void>
  list: (projectId: string | null) => HistoryListItem[]
}

export const useHistoryStore = create<HistoryState>((set, get) => ({
  byProject: {},
  loading: {},

  load: async (projectId) => {
    if (!projectId) return
    if (get().byProject[projectId] !== undefined) return
    await get().refresh(projectId)
  },

  refresh: async (projectId) => {
    if (!projectId) return
    set((s) => ({ loading: { ...s.loading, [projectId]: true } }))
    try {
      const entries = await historyService.list(projectId, 100)
      set((s) => ({
        byProject: { ...s.byProject, [projectId]: entries },
        loading: { ...s.loading, [projectId]: false },
      }))
    } catch (err) {
      console.error('history refresh failed:', err)
      set((s) => ({ loading: { ...s.loading, [projectId]: false } }))
    }
  },

  clear: async (projectId) => {
    if (!projectId) return
    await historyService.clear(projectId)
    set((s) => ({ byProject: { ...s.byProject, [projectId]: [] } }))
  },

  list: (projectId) => {
    if (!projectId) return []
    return get().byProject[projectId] ?? []
  },
}))
