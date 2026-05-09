import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import {
  mockService,
  type MockOverride,
  type MockStatus,
  type MockLogEvent,
  type SaveMockOverrideInput,
} from '@/services/mockService'

const LOG_LIMIT = 200

interface MockState {
  status: MockStatus
  overridesByProject: Record<string, MockOverride[]>
  useMockByProject: Record<string, boolean>
  logs: MockLogEvent[]
  unsubscribe: (() => void) | null

  init: () => Promise<void>
  start: (projectId: string, port: number) => Promise<void>
  stop: () => Promise<void>
  refreshStatus: () => Promise<void>
  listOverrides: (projectId: string) => Promise<MockOverride[]>
  saveOverride: (input: SaveMockOverrideInput) => Promise<MockOverride>
  removeOverride: (projectId: string, id: string) => Promise<void>
  clearLogs: () => void
  setUseMockForProject: (projectId: string, value: boolean) => void
  resolveExecutionBaseURL: (projectId: string, configured: string) => string
}

export const useMockStore = create<MockState>()(
  persist(
    (set, get) => ({
      status: { running: false, requestCount: 0 },
      overridesByProject: {},
      useMockByProject: {},
      logs: [],
      unsubscribe: null,

      init: async () => {
        if (get().unsubscribe) return
        const unsubscribe = mockService.onRequest((ev) => {
          set((s) => {
            const next = [ev, ...s.logs]
            if (next.length > LOG_LIMIT) next.length = LOG_LIMIT
            return {
              logs: next,
              status: { ...s.status, requestCount: s.status.requestCount + 1 },
            }
          })
        })
        set({ unsubscribe })
        await get().refreshStatus()
      },

      start: async (projectId, port) => {
        const status = await mockService.start(projectId, port)
        set({ status })
      },

      stop: async () => {
        await mockService.stop()
        set({ status: { running: false, requestCount: 0 } })
      },

      refreshStatus: async () => {
        const status = await mockService.status()
        set({ status })
      },

      listOverrides: async (projectId) => {
        const rows = await mockService.list(projectId)
        set((s) => ({ overridesByProject: { ...s.overridesByProject, [projectId]: rows } }))
        return rows
      },

      saveOverride: async (input) => {
        const saved = await mockService.save(input)
        const next = await mockService.list(input.projectID)
        set((s) => ({ overridesByProject: { ...s.overridesByProject, [input.projectID]: next } }))
        return saved
      },

      removeOverride: async (projectId, id) => {
        await mockService.remove(id)
        const next = (get().overridesByProject[projectId] ?? []).filter((o) => o.id !== id)
        set((s) => ({ overridesByProject: { ...s.overridesByProject, [projectId]: next } }))
      },

      clearLogs: () => set({ logs: [] }),

      setUseMockForProject: (projectId, value) => {
        set((s) => ({ useMockByProject: { ...s.useMockByProject, [projectId]: value } }))
      },

      resolveExecutionBaseURL: (projectId, configured) => {
        const state = get()
        const enabled = state.useMockByProject[projectId]
        if (!enabled) return configured
        if (!state.status.running) return configured
        if (!state.status.url) return configured
        return state.status.url
      },
    }),
    {
      name: 'spectra:mock',
      partialize: (s) => ({ useMockByProject: s.useMockByProject }),
    },
  ),
)
