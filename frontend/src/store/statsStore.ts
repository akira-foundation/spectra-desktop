import { create } from 'zustand'
import { scannerService, type ProjectStats } from '@/services/scannerService'

const EMPTY_STATS: ProjectStats = {
  routes: 0,
  models: 0,
  middleware: 0,
  controllers: 0,
  errors: 0,
} as ProjectStats

interface StatsState {
  byProject: Record<string, ProjectStats>
  loading: Record<string, boolean>

  load: (projectId: string) => Promise<void>
  get: (projectId: string | null) => ProjectStats
  isLoading: (projectId: string | null) => boolean
}

export const useStatsStore = create<StatsState>((set, get) => ({
  byProject: {},
  loading: {},

  load: async (projectId) => {
    if (!projectId) return
    set((s) => ({ loading: { ...s.loading, [projectId]: true } }))
    try {
      const stats = await scannerService.getStats(projectId)
      set((s) => ({
        byProject: { ...s.byProject, [projectId]: stats },
        loading: { ...s.loading, [projectId]: false },
      }))
    } catch (err) {
      console.error('stats load failed:', err)
      set((s) => ({ loading: { ...s.loading, [projectId]: false } }))
    }
  },

  get: (projectId) => {
    if (!projectId) return EMPTY_STATS
    return get().byProject[projectId] ?? EMPTY_STATS
  },

  isLoading: (projectId) => {
    if (!projectId) return false
    return get().loading[projectId] ?? false
  },
}))
