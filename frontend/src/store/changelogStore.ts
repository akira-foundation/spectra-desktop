import { create } from 'zustand'
import {
  changelogService,
  type SnapshotSummary,
  type SnapshotDiff,
} from '@/services/changelogService'

interface ChangelogState {
  snapshotsByProject: Record<string, SnapshotSummary[]>
  diffsByID: Record<string, SnapshotDiff>
  loading: Record<string, boolean>

  load: (projectId: string) => Promise<void>
  refresh: (projectId: string) => Promise<void>
  loadDiff: (snapshotId: string) => Promise<SnapshotDiff | null>
  list: (projectId: string | null) => SnapshotSummary[]
}

export const useChangelogStore = create<ChangelogState>((set, get) => ({
  snapshotsByProject: {},
  diffsByID: {},
  loading: {},

  load: async (projectId) => {
    if (!projectId) return
    if (get().snapshotsByProject[projectId] !== undefined) return
    await get().refresh(projectId)
  },

  refresh: async (projectId) => {
    if (!projectId) return
    set((s) => ({ loading: { ...s.loading, [projectId]: true } }))
    try {
      const rows = await changelogService.list(projectId)
      set((s) => ({
        snapshotsByProject: { ...s.snapshotsByProject, [projectId]: rows },
        loading: { ...s.loading, [projectId]: false },
      }))
    } catch (err) {
      console.error('changelog load failed:', err)
      set((s) => ({ loading: { ...s.loading, [projectId]: false } }))
    }
  },

  loadDiff: async (snapshotId) => {
    const cached = get().diffsByID[snapshotId]
    if (cached) return cached
    const diff = await changelogService.getDiff(snapshotId)
    if (diff) {
      set((s) => ({ diffsByID: { ...s.diffsByID, [snapshotId]: diff } }))
    }
    return diff
  },

  list: (projectId) => {
    if (!projectId) return []
    return get().snapshotsByProject[projectId] ?? []
  },
}))
