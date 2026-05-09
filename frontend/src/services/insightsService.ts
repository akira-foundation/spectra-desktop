import { GetInsights } from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type Insights = app.InsightsDTO
export type EndpointLatencySeries = app.EndpointLatencySeriesDTO
export type HourlyCell = app.HourlyCellDTO
export type FlakyEndpoint = app.FlakyEndpointDTO

export const insightsService = {
  async get(projectId: string, days: number): Promise<Insights | null> {
    const result = await GetInsights(projectId, days)
    return (result as Insights | null) ?? null
  },
}
