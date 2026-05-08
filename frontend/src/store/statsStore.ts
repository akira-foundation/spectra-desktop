import { create } from 'zustand'
import { scannerService, type ProjectStats, type StatsReport } from '@/services/scannerService'

const EMPTY_STATS: ProjectStats = {
  routes: 0,
  models: 0,
  middleware: 0,
  controllers: 0,
  errors: 0,
} as ProjectStats

const EMPTY_REPORT = { cards: [] } as unknown as StatsReport

interface StatsState {
  byProject: Record<string, ProjectStats>
  reportByProject: Record<string, StatsReport>
  loading: Record<string, boolean>

  load: (projectId: string) => Promise<void>
  loadReport: (projectId: string) => Promise<void>
  get: (projectId: string | null) => ProjectStats
  getReport: (projectId: string | null) => StatsReport
  isLoading: (projectId: string | null) => boolean
}

export const useStatsStore = create<StatsState>((set, get) => ({
  byProject: {},
  reportByProject: {},
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

  loadReport: async (projectId) => {
    if (!projectId) return
    try {
      const report = await scannerService.getStatsReport(projectId)
      set((s) => ({ reportByProject: { ...s.reportByProject, [projectId]: report } }))
    } catch (err) {
      console.error('stats report load failed:', err)
    }
  },

  get: (projectId) => {
    if (!projectId) return EMPTY_STATS
    return get().byProject[projectId] ?? EMPTY_STATS
  },

  getReport: (projectId) => {
    if (!projectId) return EMPTY_REPORT
    return get().reportByProject[projectId] ?? EMPTY_REPORT
  },

  isLoading: (projectId) => {
    if (!projectId) return false
    return get().loading[projectId] ?? false
  },
}))
