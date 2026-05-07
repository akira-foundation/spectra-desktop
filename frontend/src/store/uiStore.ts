import { create } from 'zustand'

export type PageType = 'inspector' | 'dashboard' | 'settings'

export interface UIState {
  sidebarOpen: boolean
  isCommandPaletteOpen: boolean
  isAuthDrawerOpen: boolean
  isAddProjectOpen: boolean
  activeAuthMethod: string
  currentPage: PageType

  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  toggleCommandPalette: () => void
  setCommandPaletteOpen: (open: boolean) => void
  setAuthDrawerOpen: (open: boolean) => void
  setAddProjectOpen: (open: boolean) => void
  setActiveAuthMethod: (method: string) => void
  setCurrentPage: (page: PageType) => void
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: false,
  isCommandPaletteOpen: false,
  isAuthDrawerOpen: false,
  isAddProjectOpen: false,
  activeAuthMethod: 'current-user',
  currentPage: 'inspector',

  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),

  toggleCommandPalette: () =>
    set((state) => ({ isCommandPaletteOpen: !state.isCommandPaletteOpen })),
  setCommandPaletteOpen: (open) => set({ isCommandPaletteOpen: open }),

  setAuthDrawerOpen: (open) => set({ isAuthDrawerOpen: open }),
  setAddProjectOpen: (open) => set({ isAddProjectOpen: open }),

  setActiveAuthMethod: (method) => set({ activeAuthMethod: method }),
  setCurrentPage: (page) => set({ currentPage: page }),
}))
