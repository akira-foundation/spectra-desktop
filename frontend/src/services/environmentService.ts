import {
  ListEnvironments,
  SaveEnvironment,
  DeleteEnvironment,
  SetActiveEnvironment,
} from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type EnvironmentDTO = app.EnvironmentDTO

export interface SaveEnvironmentInput {
  id?: string
  projectID: string
  name: string
  vars: Record<string, string>
  sortOrder?: number
}

export const environmentService = {
  async list(projectId: string): Promise<EnvironmentDTO[]> {
    const rows = await ListEnvironments(projectId)
    return rows ?? []
  },
  async save(input: SaveEnvironmentInput): Promise<EnvironmentDTO | null> {
    const result = await SaveEnvironment(
      app.SaveEnvironmentInput.createFrom({
        id: input.id ?? '',
        projectID: input.projectID,
        name: input.name,
        vars: input.vars,
        sortOrder: input.sortOrder ?? 0,
      }),
    )
    return (result as EnvironmentDTO | null) ?? null
  },
  async delete(id: string): Promise<void> {
    await DeleteEnvironment(id)
  },
  async setActive(projectId: string, envId: string): Promise<void> {
    await SetActiveEnvironment(projectId, envId)
  },
}
