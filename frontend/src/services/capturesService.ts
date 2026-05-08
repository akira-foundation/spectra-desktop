import {
  ListEndpointCaptures,
  SaveEndpointCaptures,
  ListCapturedValues,
  ClearCapturedValues,
} from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type EndpointCapture = app.EndpointCaptureDTO
export type CapturedValue = app.CapturedValueDTO

export interface SaveCapturesInput {
  projectID: string
  endpointKey: string
  captures: EndpointCapture[]
}

export const capturesService = {
  async list(projectId: string, endpointKey: string): Promise<EndpointCapture[]> {
    const rows = await ListEndpointCaptures(projectId, endpointKey)
    return rows ?? []
  },
  async save(input: SaveCapturesInput): Promise<void> {
    await SaveEndpointCaptures(
      app.SaveCapturesInput.createFrom({
        projectID: input.projectID,
        endpointKey: input.endpointKey,
        captures: input.captures,
      }),
    )
  },
  async listValues(projectId: string): Promise<CapturedValue[]> {
    const rows = await ListCapturedValues(projectId)
    return rows ?? []
  },
  async clearValues(projectId: string): Promise<void> {
    await ClearCapturedValues(projectId)
  },
}
