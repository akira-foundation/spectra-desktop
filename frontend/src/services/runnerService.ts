import { ExecuteRequest } from '../../wailsjs/go/app/App'
import { app, httpclient } from '../../wailsjs/go/models'

export type RunnerResponse = httpclient.Response
export type RunnerInput = app.ExecuteRequestInput

export interface RunnerService {
  execute(input: RunnerInput): Promise<RunnerResponse>
}

export const runnerService: RunnerService = {
  async execute(input) {
    return ExecuteRequest(app.ExecuteRequestInput.createFrom(input))
  },
}
