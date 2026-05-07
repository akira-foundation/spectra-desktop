import {
  ListProjects,
  SaveProject,
  DeleteProject,
  MarkProjectSynced,
} from '../../wailsjs/go/app/App'
import { domain } from '../../wailsjs/go/models'

export type ProjectRecord = domain.Project
export type ProjectInput = domain.ProjectInput

export interface ProjectStorageService {
  list(): Promise<ProjectRecord[]>
  save(input: ProjectInput): Promise<ProjectRecord>
  remove(id: string): Promise<void>
  markSynced(id: string): Promise<void>
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
}
