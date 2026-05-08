import {
  ListHistory,
  GetHistoryEntry,
  ClearHistory,
} from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type HistoryListItem = app.HistoryListItem
export type HistoryEntryDetail = app.HistoryEntryDetail

export const historyService = {
  async list(projectId: string, limit = 100): Promise<HistoryListItem[]> {
    const rows = await ListHistory(projectId, limit)
    return rows ?? []
  },
  async get(id: string): Promise<HistoryEntryDetail | null> {
    const entry = await GetHistoryEntry(id)
    return (entry as HistoryEntryDetail | null) ?? null
  },
  async clear(projectId: string): Promise<void> {
    await ClearHistory(projectId)
  },
}
