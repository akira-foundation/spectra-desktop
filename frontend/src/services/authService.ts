import { GetProjectAuth, ClearProjectAuth } from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type ProjectAuthState = app.ProjectAuthState

export const authService = {
  async get(projectId: string): Promise<ProjectAuthState | null> {
    const result = (await GetProjectAuth(projectId)) as ProjectAuthState | null | undefined
    return result ?? null
  },
  async clear(projectId: string): Promise<void> {
    await ClearProjectAuth(projectId)
  },
}
