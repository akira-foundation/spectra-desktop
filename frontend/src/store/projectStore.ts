import { create } from 'zustand'
import type { Project } from '@/types/project'
import { syncService } from '@/services/syncService'
import { projectStorageService, type ProjectInput } from '@/services/projectStorageService'
import { projectFromRecord } from '@/lib/project-factory'

interface ProjectState {
  projects: Project[]
  activeProjectId: string | null
  isLoading: boolean
  isSyncing: string | null
  syncingProjects: Set<string>
  error: string | null
  lastSyncTime: Record<string, number>

  setActiveProject: (id: string) => void
  loadFromStorage: () => Promise<void>
  addProjectFromInput: (input: ProjectInput) => Promise<Project>
  removeProject: (id: string) => Promise<void>
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

  setActiveProject: (id) => set({ activeProjectId: id }),

  loadFromStorage: async () => {
    set({ isLoading: true, error: null })
    try {
      const records = await projectStorageService.list()
      const projects = records.map(projectFromRecord)
      set((state) => ({
        projects,
        activeProjectId: state.activeProjectId ?? projects[0]?.id ?? null,
        isLoading: false,
      }))
    } catch (err) {
      set({ error: errorMessage(err), isLoading: false })
    }
  },

  addProjectFromInput: async (input) => {
    const record = await projectStorageService.save(input)
    const project = projectFromRecord(record)
    set((state) => {
      const others = state.projects.filter((p) => p.id !== project.id)
      return {
        projects: [...others, project],
        activeProjectId: project.id,
        error: null,
      }
    })
    return project
  },

  removeProject: async (id) => {
    await projectStorageService.remove(id)
    set((state) => {
      const filtered = state.projects.filter((p) => p.id !== id)
      const newActiveId = state.activeProjectId === id ? filtered[0]?.id ?? null : state.activeProjectId
      return { projects: filtered, activeProjectId: newActiveId }
    })
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
      if (!response.success) {
        set({ error: response.message })
        return
      }
      try {
        await projectStorageService.markSynced(projectId)
      } catch (err) {
        console.error('markSynced failed:', err)
      }
      set((s) => ({
        projects: s.projects.map((p) =>
          p.id === projectId
            ? { ...p, stats: response.stats, lastSyncTime: new Date(), status: 'connected' }
            : p,
        ),
        lastSyncTime: { ...s.lastSyncTime, [projectId]: Date.now() },
        error: null,
      }))
    } catch (err) {
      set({ error: errorMessage(err) })
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
      set({ error: errorMessage(err) })
      return false
    }
  },
}))

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
