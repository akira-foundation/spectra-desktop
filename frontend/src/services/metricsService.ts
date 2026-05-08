import { GetDashboardMetrics } from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type DashboardMetrics = app.DashboardMetrics
export type StatusBucketDTO = app.StatusBucketDTO
export type LatencyDTO = app.LatencyDTO
export type VolumePoint = app.VolumePoint
export type EndpointMetricDTO = app.EndpointMetricDTO

export const metricsService = {
  async get(projectId: string): Promise<DashboardMetrics | null> {
    const result = await GetDashboardMetrics(projectId)
    return (result as DashboardMetrics | null) ?? null
  },
}
