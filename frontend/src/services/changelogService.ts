import { ListSnapshots, GetSnapshotDiff } from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type SnapshotSummary = app.SnapshotSummary
export type SnapshotDiff = app.SnapshotDiff
export type SnapshotDiffEntry = app.SnapshotDiffEntry

export const changelogService = {
  async list(projectId: string, limit = 50): Promise<SnapshotSummary[]> {
    const rows = await ListSnapshots(projectId, limit)
    return rows ?? []
  },
  async getDiff(snapshotId: string): Promise<SnapshotDiff | null> {
    const result = await GetSnapshotDiff(snapshotId)
    return (result as SnapshotDiff | null) ?? null
  },
}
