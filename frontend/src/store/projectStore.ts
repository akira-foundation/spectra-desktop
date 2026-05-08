import { create } from 'zustand'
import type { Project } from '@/types/project'
import { syncService } from '@/services/syncService'
import {
  projectStorageService,
  type ProjectInput,
  type DetectionResult,
} from '@/services/projectStorageService'
import { projectFromRecord } from '@/lib/project-factory'

interface ProjectState {
  projects: Project[]
  activeProjectId: string | null
  detections: Record<string, DetectionResult>
  isLoading: boolean
  isSyncing: string | null
  syncingProjects: Set<string>
  error: string | null
  lastSyncTime: Record<string, number>

  loadFromStorage: () => Promise<void>
  setActiveProject: (id: string) => Promise<void>
  addProjectFromInput: (input: ProjectInput) => Promise<Project>
  removeProject: (id: string) => Promise<void>
  refreshDetection: (id: string) => Promise<void>
  refreshProject: (id: string) => Promise<void>
  updateBaseURL: (id: string, baseUrl: string) => Promise<void>
  updateAuthRoutes: (id: string, loginId: string, logoutId: string, tokenPath: string) => Promise<void>
  setActiveEnvironment: (id: string, envId: string) => void
  syncProject: (projectId: string) => Promise<void>
  testConnection: (projectId: string) => Promise<boolean>
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  activeProjectId: null,
  detections: {},
  isLoading: false,
  isSyncing: null,
  syncingProjects: new Set(),
  error: null,
  lastSyncTime: {},

  loadFromStorage: async () => {
    set({ isLoading: true, error: null })
    try {
      const [records, persistedActive] = await Promise.all([
        projectStorageService.list(),
        projectStorageService.getActive(),
      ])
      const projects = records.map(projectFromRecord)
      const valid = projects.find((p) => p.id === persistedActive)
      const activeId = valid?.id ?? projects[0]?.id ?? null
      set({ projects, activeProjectId: activeId, isLoading: false })
      if (activeId && activeId !== persistedActive) {
        await projectStorageService.setActive(activeId).catch(() => undefined)
      }
      projects.forEach((p) => {
        void get().refreshDetection(p.id)
      })
    } catch (err) {
      set({ error: errorMessage(err), isLoading: false })
    }
  },

  setActiveProject: async (id) => {
    const exists = get().projects.some((p) => p.id === id)
    if (!exists) return
    set({ activeProjectId: id })
    try {
      await projectStorageService.setActive(id)
    } catch (err) {
      console.error('persist active project failed:', err)
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
    try {
      await projectStorageService.setActive(project.id)
    } catch (err) {
      console.error('persist active project failed:', err)
    }
    void get().refreshDetection(project.id)
    return project
  },

  removeProject: async (id) => {
    await projectStorageService.remove(id)
    set((state) => {
      const filtered = state.projects.filter((p) => p.id !== id)
      const newActiveId =
        state.activeProjectId === id ? filtered[0]?.id ?? null : state.activeProjectId
      const detections = { ...state.detections }
      delete detections[id]
      return { projects: filtered, activeProjectId: newActiveId, detections }
    })
    try {
      const { useEndpointsStore } = await import('./endpointsStore')
      const { useUIStore } = await import('./uiStore')
      const { useAuthStore } = await import('./authStore')
      const { useStatsStore } = await import('./statsStore')
      useEndpointsStore.getState().clear(id)
      useAuthStore.setState((s) => {
        const next = { ...s.byProject }
        delete next[id]
        return { byProject: next }
      })
      useStatsStore.setState((s) => {
        const next = { ...s.byProject }
        delete next[id]
        const loading = { ...s.loading }
        delete loading[id]
        return { byProject: next, loading }
      })
      useUIStore.setState((s) => {
        const sel = { ...s.selectedEndpointByProject }
        delete sel[id]
        const bodies = { ...s.requestBodyByEndpoint }
        const headers = { ...s.requestHeadersByEndpoint }
        for (const key of Object.keys(bodies)) {
          if (key.startsWith(id + '::')) delete bodies[key]
        }
        for (const key of Object.keys(headers)) {
          if (key.startsWith(id + '::')) delete headers[key]
        }
        const pinned = { ...s.pinnedEndpointsByProject }
        delete pinned[id]
        const order = { ...s.groupOrderByProject }
        delete order[id]
        return {
          selectedEndpointByProject: sel,
          requestBodyByEndpoint: bodies,
          requestHeadersByEndpoint: headers,
          pinnedEndpointsByProject: pinned,
          groupOrderByProject: order,
        }
      })
    } catch (err) {
      console.error('cleanup project stores failed:', err)
    }
    if (get().activeProjectId) {
      await projectStorageService.setActive(get().activeProjectId!).catch(() => undefined)
    } else {
      await projectStorageService.setActive('').catch(() => undefined)
    }
  },

  updateBaseURL: async (id, baseUrl) => {
    await projectStorageService.updateBaseURL(id, baseUrl)
    set((state) => ({
      projects: state.projects.map((p) => (p.id === id ? { ...p, baseUrl } : p)),
    }))
  },

  updateAuthRoutes: async (id, loginId, logoutId, tokenPath) => {
    await projectStorageService.updateAuthRoutes(id, loginId, logoutId, tokenPath)
    set((state) => ({
      projects: state.projects.map((p) =>
        p.id === id
          ? {
              ...p,
              loginEndpointId: loginId,
              loginTokenPath: tokenPath,
              logoutEndpointId: logoutId,
            }
          : p,
      ),
    }))
  },

  setActiveEnvironment: (id, envId) => {
    set((state) => ({
      projects: state.projects.map((p) =>
        p.id === id ? { ...p, activeEnvironmentId: envId } : p,
      ),
    }))
  },

  refreshProject: async (id) => {
    try {
      const records = await projectStorageService.list()
      const fresh = records.find((r) => r.id === id)
      if (!fresh) return
      const project = projectFromRecord(fresh)
      set((state) => ({
        projects: state.projects.map((p) => (p.id === id ? { ...project, stats: p.stats, lastSyncTime: p.lastSyncTime } : p)),
      }))
    } catch (err) {
      console.error('refresh project failed:', err)
    }
  },

  refreshDetection: async (id) => {
    try {
      const result = await projectStorageService.detect(id)
      set((state) => ({ detections: { ...state.detections, [id]: result } }))
    } catch (err) {
      console.error('detect failed:', id, err)
    }
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
      void get().refreshDetection(projectId)
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
