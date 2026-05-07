import type { Project } from '@/types/project'

const PROJECTS_KEY = 'spectra:projects'
const LAST_SYNC_KEY = 'spectra:last_sync'
const CACHE_DURATION = 5 * 60 * 1000 // 5 minutes

export interface StorageData {
  projects: Project[]
  timestamp: number
}

export const storageService = {
  saveProjects(projects: Project[]): void {
    try {
      const data: StorageData = {
        projects,
        timestamp: Date.now(),
      }
      localStorage.setItem(PROJECTS_KEY, JSON.stringify(data))
    } catch (error) {
      console.error('Failed to save projects to storage:', error)
    }
  },

  getProjects(): Project[] {
    try {
      const data = localStorage.getItem(PROJECTS_KEY)
      if (!data) return []

      const parsed: StorageData = JSON.parse(data)
      // Check if cache is still valid
      const now = Date.now()
      const age = now - parsed.timestamp

      if (age > CACHE_DURATION) {
        // Cache expired, return empty array to trigger refresh
        return []
      }

      // Convert string dates back to Date objects
      return parsed.projects.map((p) => ({
        ...p,
        lastSyncTime: p.lastSyncTime ? new Date(p.lastSyncTime) : null,
      }))
    } catch (error) {
      console.error('Failed to retrieve projects from storage:', error)
      return []
    }
  },

  clearProjects(): void {
    try {
      localStorage.removeItem(PROJECTS_KEY)
    } catch (error) {
      console.error('Failed to clear projects from storage:', error)
    }
  },

  getLastSyncTime(projectId: string): number | null {
    try {
      const data = localStorage.getItem(`${LAST_SYNC_KEY}:${projectId}`)
      return data ? parseInt(data, 10) : null
    } catch (error) {
      console.error('Failed to get last sync time:', error)
      return null
    }
  },

  setLastSyncTime(projectId: string, timestamp: number): void {
    try {
      localStorage.setItem(`${LAST_SYNC_KEY}:${projectId}`, timestamp.toString())
    } catch (error) {
      console.error('Failed to set last sync time:', error)
    }
  },

  shouldRefresh(projectId: string): boolean {
    const lastSync = this.getLastSyncTime(projectId)
    if (!lastSync) return true

    const now = Date.now()
    const age = now - lastSync

    return age > CACHE_DURATION
  },
}
