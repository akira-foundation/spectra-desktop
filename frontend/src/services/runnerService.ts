import { ExecuteRequest } from '../../wailsjs/go/app/App'
import { app, httpclient } from '../../wailsjs/go/models'

export type RunnerResponse = httpclient.Response

export interface RunnerInput {
  projectID: string
  endpointID?: string
  accountID?: string
  method: string
  path: string
  headers?: Record<string, string>
  body?: string
  multipart?: { name: string; value?: string; filePath?: string }[]
  baseUrl?: string
  timeoutMs?: number
  skipAuth?: boolean
}

export interface RunnerService {
  execute(input: RunnerInput): Promise<RunnerResponse>
}

export const runnerService: RunnerService = {
  async execute(input) {
    return ExecuteRequest(app.ExecuteRequestInput.createFrom(input))
  },
}
