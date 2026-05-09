import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import {
  accountsService,
  type ProjectAccount,
  type SaveAccountInput,
} from '@/services/accountsService'

interface AccountsState {
  byProject: Record<string, ProjectAccount[]>
  loading: Record<string, boolean>
  activeByProject: Record<string, string | null>
  activeByTab: Record<string, string>

  list: (projectId: string, force?: boolean) => Promise<ProjectAccount[]>
  save: (input: SaveAccountInput) => Promise<ProjectAccount>
  remove: (projectId: string, accountId: string) => Promise<void>
  setDefault: (projectId: string, accountId: string) => Promise<void>
  refresh: (projectId: string, accountId: string) => Promise<ProjectAccount>
  totp: (accountId: string) => Promise<string>

  setActive: (projectId: string, accountId: string | null) => void
  setActiveForTab: (tabId: string, accountId: string | null) => void
  resolveActive: (projectId: string, tabId?: string | null) => ProjectAccount | null
  resolveActiveId: (projectId: string, tabId?: string | null) => string | null
}

export const useAccountsStore = create<AccountsState>()(
  persist(
    (set, get) => ({
      byProject: {},
      loading: {},
      activeByProject: {},
      activeByTab: {},

      list: async (projectId, force) => {
        if (!projectId) return []
        const cached = get().byProject[projectId]
        if (cached && !force) return cached
        set((s) => ({ loading: { ...s.loading, [projectId]: true } }))
        try {
          const rows = await accountsService.list(projectId)
          set((s) => ({
            byProject: { ...s.byProject, [projectId]: rows },
            loading: { ...s.loading, [projectId]: false },
          }))
          if (!get().activeByProject[projectId]) {
            const def = rows.find((r) => r.isDefault) ?? rows[0]
            if (def) {
              set((s) => ({
                activeByProject: { ...s.activeByProject, [projectId]: def.id },
              }))
            }
          }
          return rows
        } catch (err) {
          console.error('list accounts failed:', err)
          set((s) => ({ loading: { ...s.loading, [projectId]: false } }))
          return []
        }
      },

      save: async (input) => {
        const saved = await accountsService.save(input)
        const projectId = saved.projectID
        const next = await accountsService.list(projectId)
        set((s) => ({ byProject: { ...s.byProject, [projectId]: next } }))
        return saved
      },

      remove: async (projectId, accountId) => {
        await accountsService.remove(accountId)
        const next = (get().byProject[projectId] ?? []).filter((a) => a.id !== accountId)
        set((s) => ({
          byProject: { ...s.byProject, [projectId]: next },
          activeByProject:
            s.activeByProject[projectId] === accountId
              ? { ...s.activeByProject, [projectId]: next[0]?.id ?? null }
              : s.activeByProject,
        }))
      },

      setDefault: async (projectId, accountId) => {
        await accountsService.setDefault(projectId, accountId)
        const next = await accountsService.list(projectId)
        set((s) => ({ byProject: { ...s.byProject, [projectId]: next } }))
      },

      refresh: async (projectId, accountId) => {
        const refreshed = await accountsService.refresh(accountId)
        const list = (get().byProject[projectId] ?? []).map((a) =>
          a.id === accountId ? refreshed : a,
        )
        set((s) => ({ byProject: { ...s.byProject, [projectId]: list } }))
        return refreshed
      },

      totp: async (accountId) => accountsService.totp(accountId),

      setActive: (projectId, accountId) => {
        set((s) => ({
          activeByProject: { ...s.activeByProject, [projectId]: accountId },
        }))
      },

      setActiveForTab: (tabId, accountId) => {
        set((s) => {
          const next = { ...s.activeByTab }
          if (accountId) next[tabId] = accountId
          else delete next[tabId]
          return { activeByTab: next }
        })
      },

      resolveActive: (projectId, tabId) => {
        return resolve(get(), projectId, tabId)
      },

      resolveActiveId: (projectId, tabId) => {
        return resolve(get(), projectId, tabId)?.id ?? null
      },
    }),
    {
      name: 'spectra:accounts',
      partialize: (s) => ({
        activeByProject: s.activeByProject,
        activeByTab: s.activeByTab,
      }),
    },
  ),
)

function resolve(
  state: AccountsState,
  projectId: string,
  tabId?: string | null,
): ProjectAccount | null {
  if (!projectId) return null
  const list = state.byProject[projectId] ?? []
  if (list.length === 0) return null
  if (tabId && state.activeByTab[tabId]) {
    const tabAcc = list.find((a) => a.id === state.activeByTab[tabId])
    if (tabAcc) return tabAcc
  }
  const activeId = state.activeByProject[projectId]
  if (activeId) {
    const found = list.find((a) => a.id === activeId)
    if (found) return found
  }
  return list.find((a) => a.isDefault) ?? list[0]
}
