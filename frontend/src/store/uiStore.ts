import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

export type PageType = 'inspector' | 'dashboard' | 'settings'

export interface UIState {
  sidebarOpen: boolean
  isCommandPaletteOpen: boolean
  isAuthDrawerOpen: boolean
  isAddProjectOpen: boolean
  activeAuthMethod: string
  currentPage: PageType
  selectedEndpointByProject: Record<string, string>
  requestBodyByEndpoint: Record<string, string>
  requestHeadersByEndpoint: Record<string, Array<{ key: string; value: string; enabled: boolean }>>

  setSelectedEndpoint: (projectId: string, tag: string | null) => void
  setRequestBody: (endpointId: string, body: string) => void
  clearRequestBody: (endpointId: string) => void
  setRequestHeaders: (endpointId: string, headers: Array<{ key: string; value: string; enabled: boolean }>) => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  toggleCommandPalette: () => void
  setCommandPaletteOpen: (open: boolean) => void
  setAuthDrawerOpen: (open: boolean) => void
  setAddProjectOpen: (open: boolean) => void
  setActiveAuthMethod: (method: string) => void
  setCurrentPage: (page: PageType) => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: false,
      isCommandPaletteOpen: false,
      isAuthDrawerOpen: false,
      isAddProjectOpen: false,
      activeAuthMethod: 'current-user',
      currentPage: 'inspector',
      selectedEndpointByProject: {},
      requestBodyByEndpoint: {},
      requestHeadersByEndpoint: {},

      setSelectedEndpoint: (projectId, tag) =>
        set((state) => {
          const next = { ...state.selectedEndpointByProject }
          if (tag) next[projectId] = tag
          else delete next[projectId]
          return { selectedEndpointByProject: next }
        }),
      setRequestBody: (endpointId, body) =>
        set((state) => ({
          requestBodyByEndpoint: { ...state.requestBodyByEndpoint, [endpointId]: body },
        })),
      clearRequestBody: (endpointId) =>
        set((state) => {
          const next = { ...state.requestBodyByEndpoint }
          delete next[endpointId]
          return { requestBodyByEndpoint: next }
        }),
      setRequestHeaders: (endpointId, hs) =>
        set((state) => ({
          requestHeadersByEndpoint: { ...state.requestHeadersByEndpoint, [endpointId]: hs },
        })),

      toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),

      toggleCommandPalette: () =>
        set((state) => ({ isCommandPaletteOpen: !state.isCommandPaletteOpen })),
      setCommandPaletteOpen: (open) => set({ isCommandPaletteOpen: open }),

      setAuthDrawerOpen: (open) => set({ isAuthDrawerOpen: open }),
      setAddProjectOpen: (open) => set({ isAddProjectOpen: open }),

      setActiveAuthMethod: (method) => set({ activeAuthMethod: method }),
      setCurrentPage: (page) => set({ currentPage: page }),
    }),
    {
      name: 'spectra:ui',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        currentPage: state.currentPage,
        activeAuthMethod: state.activeAuthMethod,
        sidebarOpen: state.sidebarOpen,
        selectedEndpointByProject: state.selectedEndpointByProject,
        requestBodyByEndpoint: state.requestBodyByEndpoint,
        requestHeadersByEndpoint: state.requestHeadersByEndpoint,
      }),
      version: 1,
    },
  ),
)
