import { create } from 'zustand'
import type { Project } from '@/types/project'
import { syncService } from '@/services/syncService'
import { storageService } from '@/services/storageService'

interface ProjectState {
  projects: Project[]
  activeProjectId: string | null
  isLoading: boolean
  isSyncing: string | null
  syncingProjects: Set<string>
  error: string | null
  lastSyncTime: Record<string, number>

  setProjects: (projects: Project[]) => void
  addProject: (project: Project) => void
  removeProject: (id: string) => void
  setActiveProject: (id: string) => void
  updateProject: (id: string, updates: Partial<Project>) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  loadFromStorage: () => void
  saveToStorage: () => void
  syncProject: (projectId: string) => Promise<void>
  testConnection: (projectId: string) => Promise<boolean>
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  activeProjectId: null,
  isLoading: false,
  isSyncing: null,
  syncingProjects: new Set(),
  error: null,
  lastSyncTime: {},

  setProjects: (projects) => set({ projects }),

  addProject: (project) => set((state) => {
    const newProjects = [...state.projects, project]
    storageService.saveProjects(newProjects)
    return {
      projects: newProjects,
      activeProjectId: project.id,
    }
  }),

  removeProject: (id) => set((state) => {
    const filtered = state.projects.filter((p) => p.id !== id)
    const newActiveId = state.activeProjectId === id
      ? filtered[0]?.id || null
      : state.activeProjectId

    storageService.saveProjects(filtered)

    return {
      projects: filtered,
      activeProjectId: newActiveId,
    }
  }),

  setActiveProject: (id) => set({ activeProjectId: id }),

  updateProject: (id, updates) => set((state) => {
    const newProjects = state.projects.map((p) =>
      p.id === id ? { ...p, ...updates } : p
    )
    storageService.saveProjects(newProjects)
    return { projects: newProjects }
  }),

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error }),

  loadFromStorage: () => {
    const projects = storageService.getProjects()
    set({ projects })
  },

  saveToStorage: () => {
    const state = get()
    storageService.saveProjects(state.projects)
  },

  syncProject: async (projectId) => {
    const state = get()
    const project = state.projects.find((p) => p.id === projectId)

    if (!project) {
      set({ error: 'Project not found' })
      return
    }

    set((s) => ({
      isSyncing: projectId,
      syncingProjects: new Set([...s.syncingProjects, projectId]),
    }))

    try {
      const response = await syncService.syncProject(project.path, project.framework)

      if (response.success) {
        const now = Date.now()
        set((s) => ({
          projects: s.projects.map((p) =>
            p.id === projectId
              ? {
                  ...p,
                  stats: response.stats,
                  lastSyncTime: new Date(),
                  status: 'connected' as const,
                }
              : p
          ),
          lastSyncTime: { ...s.lastSyncTime, [projectId]: now },
          error: null,
        }))
      } else {
        set({ error: response.message })
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      set({ error: errorMessage })
    } finally {
      set((s) => {
        const newSyncing = new Set(s.syncingProjects)
        newSyncing.delete(projectId)
        return { isSyncing: null, syncingProjects: newSyncing }
      })
    }
  },

  testConnection: async (projectId) => {
    const state = get()
    const project = state.projects.find((p) => p.id === projectId)

    if (!project) {
      set({ error: 'Project not found' })
      return false
    }

    try {
      const response = await syncService.testConnection(project.path)
      set({ error: null })
      return response.success
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      set({ error: errorMessage })
      return false
    }
  },
}))
