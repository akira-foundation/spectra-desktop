import { create } from 'zustand'
import {
  environmentService,
  type EnvironmentDTO,
  type SaveEnvironmentInput,
} from '@/services/environmentService'

interface EnvState {
  byProject: Record<string, EnvironmentDTO[]>

  load: (projectId: string) => Promise<void>
  refresh: (projectId: string) => Promise<void>
  save: (input: SaveEnvironmentInput) => Promise<EnvironmentDTO | null>
  remove: (projectId: string, envId: string) => Promise<void>
  setActive: (projectId: string, envId: string) => Promise<void>
  list: (projectId: string | null) => EnvironmentDTO[]
}

export const useEnvironmentStore = create<EnvState>((set, get) => ({
  byProject: {},

  load: async (projectId) => {
    if (!projectId) return
    if (get().byProject[projectId] !== undefined) return
    await get().refresh(projectId)
  },

  refresh: async (projectId) => {
    if (!projectId) return
    try {
      const rows = await environmentService.list(projectId)
      set((s) => ({ byProject: { ...s.byProject, [projectId]: rows } }))
    } catch (err) {
      console.error('environments refresh failed:', err)
    }
  },

  save: async (input) => {
    const env = await environmentService.save(input)
    if (env) {
      await get().refresh(input.projectID)
    }
    return env
  },

  remove: async (projectId, envId) => {
    await environmentService.delete(envId)
    await get().refresh(projectId)
  },

  setActive: async (projectId, envId) => {
    await environmentService.setActive(projectId, envId)
  },

  list: (projectId) => {
    if (!projectId) return []
    return get().byProject[projectId] ?? []
  },
}))
