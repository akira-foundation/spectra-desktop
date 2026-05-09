import { create } from 'zustand'
import { capturesService, type CapturedValue } from '@/services/capturesService'

interface CapturesState {
  byProject: Record<string, CapturedValue[]>
  refresh: (projectId: string) => Promise<void>
  set: (projectId: string, values: CapturedValue[]) => void
  clear: (projectId: string) => Promise<void>
  list: (projectId: string | null) => CapturedValue[]
}

export const useCapturesStore = create<CapturesState>((set, get) => ({
  byProject: {},

  refresh: async (projectId) => {
    if (!projectId) return
    try {
      const vals = await capturesService.listValues(projectId)
      set((s) => ({ byProject: { ...s.byProject, [projectId]: vals } }))
    } catch {}
  },

  set: (projectId, values) => {
    set((s) => ({ byProject: { ...s.byProject, [projectId]: values } }))
  },

  clear: async (projectId) => {
    if (!projectId) return
    await capturesService.clearValues(projectId)
    set((s) => ({ byProject: { ...s.byProject, [projectId]: [] } }))
  },

  list: (projectId) => {
    if (!projectId) return []
    return get().byProject[projectId] ?? []
  },
}))
