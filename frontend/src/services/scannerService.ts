import {
  ScanWorkspace,
  ScanActiveProject,
  ListEndpoints,
  GetProjectStats,
} from '../../wailsjs/go/app/App'
import type { core, domain } from '../../wailsjs/go/models'

export type ScannedEndpoint = core.Endpoint
export type ProjectStats = domain.ProjectStats

export interface ScannerService {
  scanProject(projectId: string): Promise<ScannedEndpoint[]>
  scanActive(): Promise<ScannedEndpoint[]>
  listEndpoints(projectId: string): Promise<ScannedEndpoint[]>
  getStats(projectId: string): Promise<ProjectStats>
}

export const scannerService: ScannerService = {
  async scanProject(projectId) {
    const rows = await ScanWorkspace(projectId)
    return rows ?? []
  },
  async scanActive() {
    const rows = await ScanActiveProject()
    return rows ?? []
  },
  async listEndpoints(projectId) {
    const rows = await ListEndpoints(projectId)
    return rows ?? []
  },
  async getStats(projectId) {
    return GetProjectStats(projectId)
  },
}
