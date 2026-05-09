import {
  StartMockServer,
  StopMockServer,
  MockServerStatus,
  ListMockOverrides,
  SaveMockOverride,
  DeleteMockOverride,
} from '../../wailsjs/go/app/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'
import type { app } from '../../wailsjs/go/models'

export interface MockStatus {
  running: boolean
  projectId?: string
  port?: number
  url?: string
  startedAt?: string
  requestCount: number
}

export type MockSource = 'auto' | 'history' | 'custom' | 'generated' | 'no-match'

export interface MockOverride {
  id: string
  projectID: string
  endpointId: string
  enabled: boolean
  status: number
  latencyMs: number
  body: string
  headers?: Record<string, string>
  source: MockSource
  updatedAt: string
}

export interface SaveMockOverrideInput {
  id?: string
  projectID: string
  endpointId: string
  enabled: boolean
  status?: number
  latencyMs?: number
  body?: string
  headers?: Record<string, string>
  source?: MockSource
}

export interface MockLogEvent {
  timestamp: string
  method: string
  path: string
  status: number
  durationMs: number
  source: MockSource
  endpointId?: string
  bodySize: number
}

export const mockService = {
  async start(projectId: string, port: number): Promise<MockStatus> {
    const result = await StartMockServer(projectId, port)
    return (result ?? { running: false, requestCount: 0 }) as unknown as MockStatus
  },
  async stop(): Promise<void> {
    await StopMockServer()
  },
  async status(): Promise<MockStatus> {
    const result = await MockServerStatus()
    return (result ?? { running: false, requestCount: 0 }) as unknown as MockStatus
  },
  async list(projectId: string): Promise<MockOverride[]> {
    const rows = await ListMockOverrides(projectId)
    return (rows ?? []) as unknown as MockOverride[]
  },
  async save(input: SaveMockOverrideInput): Promise<MockOverride> {
    const saved = await SaveMockOverride(input as unknown as app.SaveMockOverrideInput)
    return saved as unknown as MockOverride
  },
  async remove(id: string): Promise<void> {
    await DeleteMockOverride(id)
  },
  onRequest(handler: (ev: MockLogEvent) => void): () => void {
    const wrapped = (data: MockLogEvent) => handler(data)
    EventsOn('mock:request', wrapped)
    return () => EventsOff('mock:request')
  },
}
