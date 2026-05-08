import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { scannerService, type ScannedEndpoint } from '@/services/scannerService'

export type ScanStatus = 'idle' | 'loading' | 'scanning' | 'ready' | 'empty' | 'error'

export interface ScanError {
  message: string
  code?: string
}

interface EndpointsState {
  byProject: Record<string, ScannedEndpoint[]>
  status: Record<string, ScanStatus>
  errors: Record<string, ScanError | null>

  load: (projectId: string) => Promise<void>
  scan: (projectId: string) => Promise<void>
  clear: (projectId: string) => void
  getEndpoints: (projectId: string | null) => ScannedEndpoint[]
  getStatus: (projectId: string | null) => ScanStatus
  getError: (projectId: string | null) => ScanError | null
}

export const useEndpointsStore = create<EndpointsState>()(
  persist(
    (set, get) => ({
  byProject: {},
  status: {},
  errors: {},

  load: async (projectId) => {
    if (!projectId) return
    set((s) => ({
      status: { ...s.status, [projectId]: 'loading' },
      errors: { ...s.errors, [projectId]: null },
    }))
    try {
      const cached = await scannerService.listEndpoints(projectId)
      if (cached.length > 0) {
        const hasAnyAuthRole = cached.some((e) => e.authRole)
        if (!hasAnyAuthRole) {
          await get().scan(projectId)
          return
        }
        set((s) => ({
          byProject: { ...s.byProject, [projectId]: cached },
          status: { ...s.status, [projectId]: 'ready' },
        }))
        return
      }
      await get().scan(projectId)
    } catch (err) {
      set((s) => ({
        status: { ...s.status, [projectId]: 'error' },
        errors: { ...s.errors, [projectId]: toScanError(err) },
      }))
    }
  },

  scan: async (projectId) => {
    if (!projectId) return
    set((s) => ({
      status: { ...s.status, [projectId]: 'scanning' },
      errors: { ...s.errors, [projectId]: null },
    }))
    try {
      const endpoints = await scannerService.scanProject(projectId)
      set((s) => ({
        byProject: { ...s.byProject, [projectId]: endpoints },
        status: { ...s.status, [projectId]: endpoints.length === 0 ? 'empty' : 'ready' },
      }))
      try {
        const { useProjectStore } = await import('./projectStore')
        await useProjectStore.getState().refreshProject(projectId)
      } catch {}
    } catch (err) {
      set((s) => ({
        status: { ...s.status, [projectId]: 'error' },
        errors: { ...s.errors, [projectId]: toScanError(err) },
      }))
    }
  },

  clear: (projectId) => {
    set((s) => {
      const byProject = { ...s.byProject }
      const status = { ...s.status }
      const errors = { ...s.errors }
      delete byProject[projectId]
      delete status[projectId]
      delete errors[projectId]
      return { byProject, status, errors }
    })
  },

  getEndpoints: (projectId) => {
    if (!projectId) return []
    return get().byProject[projectId] ?? []
  },
  getStatus: (projectId) => {
    if (!projectId) return 'idle'
    return get().status[projectId] ?? 'idle'
  },
  getError: (projectId) => {
    if (!projectId) return null
    return get().errors[projectId] ?? null
  },
    }),
    {
      name: 'spectra:endpoints',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ byProject: state.byProject }),
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (!state) return
        const status: Record<string, ScanStatus> = {}
        for (const id of Object.keys(state.byProject)) {
          if ((state.byProject[id]?.length ?? 0) > 0) status[id] = 'ready'
        }
        state.status = status
        state.errors = {}
      },
    },
  ),
)

function toScanError(err: unknown): ScanError {
  const message = err instanceof Error ? err.message : String(err)
  return { message, code: classify(message) }
}

function classify(message: string): string | undefined {
  const lower = message.toLowerCase()
  if (lower.includes('php not found')) return 'php_not_found'
  if (lower.includes('artisan not found')) return 'artisan_missing'
  if (lower.includes('invalid json')) return 'invalid_json'
  if (lower.includes('no routes')) return 'no_routes'
  if (lower.includes('artisan exited')) return 'artisan_failed'
  if (lower.includes('not a laravel')) return 'not_laravel'
  if (lower.includes('no driver')) return 'no_driver'
  return undefined
}
