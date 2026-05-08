import type { Project } from '@/types/project'
import type { ProjectInfo } from '@/services/projectService'
import type { ProjectRecord } from '@/services/projectStorageService'

const emptyStats = {
  routes: 0,
  models: 0,
  middleware: 0,
  controllers: 0,
  errors: 0,
}

export function projectInputFromInfo(info: ProjectInfo) {
  return {
    id: '',
    name: info.name,
    path: info.path,
    framework: info.framework || 'other',
    frameworkVersion: info.frameworkVersion || '',
    apiFilterMode: info.apiDetection?.mode ?? 'auto',
    apiFilterValue: info.apiDetection?.value ?? '',
    baseUrl: info.defaultBaseUrl ?? '',
  }
}

export function projectFromRecord(record: ProjectRecord): Project {
  const framework = normalizeFramework(record.framework)
  return {
    id: record.id,
    name: record.name,
    path: record.path,
    framework,
    frameworkVersion: record.frameworkVersion ?? '',
    sdkVersion: '',
    baseUrl: record.baseUrl ?? '',
    loginEndpointId: record.loginEndpointId ?? '',
    loginTokenPath: record.loginTokenPath ?? '',
    lastSyncTime: record.lastSyncedAt ? toDate(record.lastSyncedAt) : null,
    status: normalizeStatus(record.status),
    stats: { ...emptyStats },
  }
}

function toDate(value: string | Date): Date {
  return value instanceof Date ? value : new Date(value)
}

function normalizeFramework(value: string): Project['framework'] {
  if (value === 'laravel' || value === 'symfony') return value
  return 'other'
}

function normalizeStatus(value: string): Project['status'] {
  switch (value) {
    case 'connected':
    case 'syncing':
    case 'error':
      return value
    default:
      return 'disconnected'
  }
}
