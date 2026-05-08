import {
  SelectProjectFolder,
  InspectProject,
  PreviewAPIRoutes,
} from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type ProjectInfo = app.ProjectInfo
export type APIDetection = app.APIDetection

export type APIFilterMode = 'auto' | 'middleware' | 'prefix' | 'all'

export interface ProjectService {
  pickFolder(): Promise<string | null>
  inspect(path: string): Promise<ProjectInfo>
  previewRoutes(path: string, mode: APIFilterMode, value: string): Promise<APIDetection>
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

  async previewRoutes(path, mode, value) {
    return PreviewAPIRoutes(path, mode, value)
  },
}
