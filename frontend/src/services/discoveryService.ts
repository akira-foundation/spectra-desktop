import { GetDiscovery } from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'

export type Discovery = app.DiscoveryDTO
export type EndpointDiscovery = app.EndpointDiscoveryDTO

export const discoveryService = {
  async get(projectId: string, staleAfterDays = 30): Promise<Discovery | null> {
    const result = await GetDiscovery(projectId, staleAfterDays)
    return (result as Discovery | null) ?? null
  },
}
