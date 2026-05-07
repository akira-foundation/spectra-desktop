import type { Project } from '@/types/project'
import type { ProjectInfo } from '@/services/projectService'

export function projectFromInfo(info: ProjectInfo): Project {
  const framework = (info.framework || 'other') as Project['framework']
  return {
    id: createId(info.path),
    name: info.name || fallbackName(info.path),
    path: info.path,
    framework: framework === 'laravel' || framework === 'symfony' ? framework : 'other',
    frameworkVersion: info.frameworkVersion || '',
    sdkVersion: '',
    lastSyncTime: null,
    status: 'disconnected',
    stats: {
      routes: 0,
      models: 0,
      middleware: 0,
      controllers: 0,
      errors: 0,
    },
  }
}

function createId(path: string): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${path}-${Date.now()}`
}

function fallbackName(path: string): string {
  const parts = path.split(/[/\\]/).filter(Boolean)
  return parts[parts.length - 1] ?? 'Untitled Project'
}
