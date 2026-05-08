import type { ScannedEndpoint } from '@/services/scannerService'

export interface EndpointGroup {
  category: string
  count: number
  items: GroupedEndpoint[]
}

export interface GroupedEndpoint {
  id: string
  method: string
  name: string
  path: string
  tag: string
  controller?: string
  middleware?: string[]
  authRequired?: boolean
  framework?: string
  confidence?: number
}

const PREFIX_SKIP = new Set(['api', 'v1', 'v2', 'v3', 'admin'])

export function groupEndpoints(endpoints: ScannedEndpoint[]): EndpointGroup[] {
  const buckets = new Map<string, GroupedEndpoint[]>()

  for (const endpoint of endpoints) {
    const category = pickCategory(endpoint.path)
    const grouped = toGrouped(endpoint)
    const list = buckets.get(category) ?? []
    list.push(grouped)
    buckets.set(category, list)
  }

  return [...buckets.entries()]
    .map(([category, items]) => ({ category, count: items.length, items: sortItems(items) }))
    .sort((a, b) => a.category.localeCompare(b.category))
}

function pickCategory(path: string): string {
  const segments = path.split('/').filter(Boolean)
  for (const segment of segments) {
    const clean = segment.toLowerCase()
    if (clean.startsWith('{')) continue
    if (PREFIX_SKIP.has(clean)) continue
    return clean.toUpperCase()
  }
  return 'GENERAL'
}

function toGrouped(endpoint: ScannedEndpoint): GroupedEndpoint {
  const auth = inferAuth(endpoint.middleware)
  return {
    id: endpoint.id || `${endpoint.method}:${endpoint.path}`,
    method: endpoint.method as string,
    name: endpoint.name || deriveName(endpoint.path),
    path: endpoint.path,
    tag: endpoint.id || `${endpoint.method}:${endpoint.path}`,
    controller: endpoint.handler,
    middleware: endpoint.middleware ?? undefined,
    authRequired: auth,
    framework: endpoint.framework,
    confidence: endpoint.confidence,
  }
}

function deriveName(path: string): string {
  const segments = path.split('/').filter(Boolean)
  return segments[segments.length - 1] ?? path
}

function sortItems(items: GroupedEndpoint[]): GroupedEndpoint[] {
  const order: Record<string, number> = {
    GET: 0,
    POST: 1,
    PUT: 2,
    PATCH: 3,
    DELETE: 4,
    OPTIONS: 5,
    HEAD: 6,
  }
  return [...items].sort((a, b) => {
    const byPath = a.path.localeCompare(b.path)
    if (byPath !== 0) return byPath
    return (order[a.method] ?? 99) - (order[b.method] ?? 99)
  })
}

function inferAuth(middleware?: string[]): boolean | undefined {
  if (!middleware || middleware.length === 0) return undefined
  return middleware.some((m) => /auth|sanctum|passport|jwt|guard/i.test(m))
}
