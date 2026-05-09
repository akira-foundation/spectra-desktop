import {
  GetEndpointDataset,
  SaveEndpointDataset,
  GenerateDatasetRows,
} from '../../wailsjs/go/app/App'

export const datasetsService = {
  async get(projectId: string, endpointKey: string): Promise<unknown[]> {
    const json = await GetEndpointDataset(projectId, endpointKey)
    if (!json) return []
    try {
      const parsed = JSON.parse(json)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  },
  async save(projectId: string, endpointKey: string, rows: unknown[]): Promise<void> {
    await SaveEndpointDataset(projectId, endpointKey, JSON.stringify(rows))
  },
  async generate(endpointId: string, count: number): Promise<unknown[]> {
    const json = await GenerateDatasetRows(endpointId, count)
    try {
      const parsed = JSON.parse(json)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  },
}
