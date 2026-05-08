import { ListEndpointTests, SaveEndpointTests } from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type EndpointTest = app.EndpointTestDTO
export type TestResult = app.TestResultDTO

export interface SaveTestsInput {
  projectID: string
  endpointKey: string
  tests: EndpointTest[]
}

export const testsService = {
  async list(projectId: string, endpointKey: string): Promise<EndpointTest[]> {
    const rows = await ListEndpointTests(projectId, endpointKey)
    return rows ?? []
  },
  async save(input: SaveTestsInput): Promise<void> {
    await SaveEndpointTests(
      app.SaveTestsInput.createFrom({
        projectID: input.projectID,
        endpointKey: input.endpointKey,
        tests: input.tests,
      }),
    )
  },
}
