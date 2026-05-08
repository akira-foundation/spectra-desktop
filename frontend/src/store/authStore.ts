import { create } from 'zustand'
import { authService, type ProjectAuthState } from '@/services/authService'

interface AuthState {
  byProject: Record<string, ProjectAuthState | null>
  loading: Record<string, boolean>

  load: (projectId: string) => Promise<void>
  refresh: (projectId: string) => Promise<void>
  clear: (projectId: string) => Promise<void>
  get: (projectId: string | null) => ProjectAuthState | null
}

export const useAuthStore = create<AuthState>((set, get) => ({
  byProject: {},
  loading: {},

  load: async (projectId) => {
    if (!projectId) return
    if (get().byProject[projectId] !== undefined) return
    set((s) => ({ loading: { ...s.loading, [projectId]: true } }))
    try {
      const state = await authService.get(projectId)
      set((s) => ({
        byProject: { ...s.byProject, [projectId]: state },
        loading: { ...s.loading, [projectId]: false },
      }))
    } catch (err) {
      console.error('load auth failed:', err)
      set((s) => ({ loading: { ...s.loading, [projectId]: false } }))
    }
  },

  refresh: async (projectId) => {
    if (!projectId) return
    try {
      const state = await authService.get(projectId)
      set((s) => ({ byProject: { ...s.byProject, [projectId]: state } }))
    } catch (err) {
      console.error('refresh auth failed:', err)
    }
  },

  clear: async (projectId) => {
    if (!projectId) return
    await authService.clear(projectId)
    set((s) => ({ byProject: { ...s.byProject, [projectId]: null } }))
  },

  get: (projectId) => {
    if (!projectId) return null
    return get().byProject[projectId] ?? null
  },
}))
