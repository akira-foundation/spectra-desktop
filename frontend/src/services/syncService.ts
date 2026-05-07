export interface ProjectStats {
  routes: number
  models: number
  middleware: number
  controllers: number
  errors: number
}

export interface SyncResponse {
  success: boolean
  stats: ProjectStats
  message: string
}

export interface ConnectionCheck {
  name: string
  passed: boolean
}

export interface ConnectionResult {
  success: boolean
  checks: ConnectionCheck[]
}

const mockStats: ProjectStats = {
  routes: 24,
  models: 8,
  middleware: 6,
  controllers: 12,
  errors: 0,
}

export const syncService = {
  async syncProject(projectPath: string, framework: string): Promise<SyncResponse> {
    await new Promise((r) => setTimeout(r, 600))
    return {
      success: true,
      stats: mockStats,
      message: `Mock sync for ${framework} at ${projectPath}`,
    }
  },

  async testConnection(_projectPath: string): Promise<ConnectionResult> {
    await new Promise((r) => setTimeout(r, 300))
    return {
      success: true,
      checks: [
        { name: 'Path exists', passed: true },
        { name: 'Framework detected', passed: true },
        { name: 'SDK reachable', passed: true },
      ],
    }
  },

  async installSDK(projectPath: string, framework: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        console.log(`SDK installed for ${framework} at ${projectPath}`)
        resolve()
      }, 800)
    })
  },
}
