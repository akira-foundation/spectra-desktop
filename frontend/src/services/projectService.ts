import {
  SelectProjectFolder,
  InspectProject,
} from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type ProjectInfo = app.ProjectInfo

export interface ProjectService {
  pickFolder(): Promise<string | null>
  inspect(path: string): Promise<ProjectInfo>
}

export const projectService: ProjectService = {
  async pickFolder() {
    try {
      const path = await SelectProjectFolder()
      return path && path.trim() ? path : null
    } catch (err) {
      console.error('pickFolder failed:', err)
      return null
    }
  },

  async inspect(path: string) {
    return InspectProject(path)
  },
}
