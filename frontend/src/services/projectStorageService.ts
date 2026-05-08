import {
  ListProjects,
  SaveProject,
  DeleteProject,
  MarkProjectSynced,
  GetActiveProjectID,
  SetActiveProjectID,
  DetectProject,
} from '../../wailsjs/go/app/App'
import { domain, core } from '../../wailsjs/go/models'

export type ProjectRecord = domain.Project
export type ProjectInput = domain.ProjectInput
export type DetectionResult = core.DetectionResult

export interface ProjectStorageService {
  list(): Promise<ProjectRecord[]>
  save(input: ProjectInput): Promise<ProjectRecord>
  remove(id: string): Promise<void>
  markSynced(id: string): Promise<void>
  getActive(): Promise<string>
  setActive(id: string): Promise<void>
  detect(id: string): Promise<DetectionResult>
}

export const projectStorageService: ProjectStorageService = {
  async list() {
    const rows = await ListProjects()
    return rows ?? []
  },
  async save(input) {
    return SaveProject(domain.ProjectInput.createFrom(input))
  },
  async remove(id) {
    await DeleteProject(id)
  },
  async markSynced(id) {
    await MarkProjectSynced(id)
  },
  async getActive() {
    return GetActiveProjectID()
  },
  async setActive(id) {
    await SetActiveProjectID(id)
  },
  async detect(id) {
    return DetectProject(id)
  },
}
