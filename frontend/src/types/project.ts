export interface Project {
  id: string
  name: string
  path: string
  framework: 'laravel' | 'symfony' | 'other'
  frameworkVersion: string
  sdkVersion: string
  baseUrl: string
  loginEndpointId?: string
  loginTokenPath?: string
  logoutEndpointId?: string
  activeEnvironmentId?: string
  lastSyncTime: Date | null
  status: 'connected' | 'disconnected' | 'syncing' | 'error'
  stats: {
    routes: number
    models: number
    middleware: number
    controllers: number
    errors: number
  }
}

export interface ProjectMetadata {
  id: string
  name: string
  path: string
  framework: string
  frameworkVersion: string
}
