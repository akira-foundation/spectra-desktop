import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

export type PageType = 'inspector' | 'dashboard' | 'collections' | 'scratch' | 'settings' | 'changelog'

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
  pinnedEndpointsByProject: Record<string, string[]>
  groupOrderByProject: Record<string, string[]>
  editingEnvId: string | null
  setEditingEnvId: (id: string | null) => void
  inspectorPending: { endpointId: string; openHistoryLatest: boolean } | null
  setInspectorPending: (p: UIState['inspectorPending']) => void
  endpointListCollapsed: boolean
  setEndpointListCollapsed: (v: boolean) => void
  pendingCurl: {
    method: string
    url: string
    baseURL?: string
    path?: string
    headers: Record<string, string>
    body?: string
    query?: Record<string, string>
  } | null
  setPendingCurl: (c: UIState['pendingCurl']) => void
  navBack: PageType[]
  navForward: PageType[]
  goBack: () => void
  goForward: () => void

  setSelectedEndpoint: (projectId: string, tag: string | null) => void
  togglePinnedEndpoint: (projectId: string, endpointKey: string) => void
  setGroupOrder: (projectId: string, order: string[]) => void
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
      pinnedEndpointsByProject: {},
      groupOrderByProject: {},
      editingEnvId: null,
      setEditingEnvId: (id) => set({ editingEnvId: id }),
      inspectorPending: null,
      setInspectorPending: (p) => set({ inspectorPending: p }),
      endpointListCollapsed: false,
      setEndpointListCollapsed: (v) => set({ endpointListCollapsed: v }),
      pendingCurl: null,
      setPendingCurl: (c) => set({ pendingCurl: c }),
      navBack: [],
      navForward: [],
      goBack: () =>
        set((s) => {
          if (s.navBack.length === 0) return s
          const prev = s.navBack[s.navBack.length - 1]
          return {
            currentPage: prev,
            navBack: s.navBack.slice(0, -1),
            navForward: [...s.navForward, s.currentPage],
          }
        }),
      goForward: () =>
        set((s) => {
          if (s.navForward.length === 0) return s
          const next = s.navForward[s.navForward.length - 1]
          return {
            currentPage: next,
            navForward: s.navForward.slice(0, -1),
            navBack: [...s.navBack, s.currentPage],
          }
        }),

      togglePinnedEndpoint: (projectId, endpointKey) =>
        set((state) => {
          const current = state.pinnedEndpointsByProject[projectId] ?? []
          const next = current.includes(endpointKey)
            ? current.filter((k) => k !== endpointKey)
            : [...current, endpointKey]
          return {
            pinnedEndpointsByProject: { ...state.pinnedEndpointsByProject, [projectId]: next },
          }
        }),
      setGroupOrder: (projectId, order) =>
        set((state) => ({
          groupOrderByProject: { ...state.groupOrderByProject, [projectId]: order },
        })),

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
      setCurrentPage: (page) =>
        set((s) => {
          if (page === s.currentPage) return s
          return {
            currentPage: page,
            navBack: [...s.navBack, s.currentPage].slice(-50),
            navForward: [],
          }
        }),
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
        pinnedEndpointsByProject: state.pinnedEndpointsByProject,
        groupOrderByProject: state.groupOrderByProject,
      }),
      version: 1,
    },
  ),
)
