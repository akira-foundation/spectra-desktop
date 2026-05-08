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

const PREFIX_SKIP = new Set([
  'api',
  'v1',
  'v2',
  'v3',
  'v4',
  'web',
  'admin',
  'http',
  'controllers',
  'app',
])

export const PINNED_CATEGORY = 'Pinned'

export interface GroupOptions {
  pinnedKeys?: string[]
  groupOrder?: string[]
}

export function endpointKey(method: string, path: string): string {
  return `${method} ${path}`
}

export function groupEndpoints(
  endpoints: ScannedEndpoint[],
  options: GroupOptions = {},
): EndpointGroup[] {
  const buckets = new Map<string, GroupedEndpoint[]>()
  const pinnedSet = new Set(options.pinnedKeys ?? [])
  const pinnedItems: GroupedEndpoint[] = []

  for (const endpoint of endpoints) {
    const grouped = toGrouped(endpoint)
    if (pinnedSet.has(endpointKey(endpoint.method, endpoint.path))) {
      pinnedItems.push(grouped)
    }
    const category = pickCategory(endpoint)
    const list = buckets.get(category) ?? []
    list.push(grouped)
    buckets.set(category, list)
  }

  const groups: EndpointGroup[] = []
  if (pinnedItems.length > 0) {
    groups.push({
      category: PINNED_CATEGORY,
      count: pinnedItems.length,
      items: sortItems(pinnedItems),
    })
  }

  const userOrder = options.groupOrder ?? []
  const seen = new Set<string>()
  for (const cat of userOrder) {
    const items = buckets.get(cat)
    if (!items) continue
    seen.add(cat)
    groups.push({ category: cat, count: items.length, items: sortItems(items) })
  }
  const remaining = [...buckets.entries()]
    .filter(([cat]) => !seen.has(cat))
    .map(([category, items]) => ({ category, count: items.length, items: sortItems(items) }))
    .sort((a, b) => a.category.localeCompare(b.category))
  groups.push(...remaining)

  return groups
}

function pickCategory(endpoint: ScannedEndpoint): string {
  return (
    fromName(endpoint.name) ??
    fromController(endpoint.handler) ??
    fromPath(endpoint.path) ??
    'GENERAL'
  )
}

function fromName(name?: string): string | null {
  if (!name) return null
  const raw = name
    .split('.')
    .map((s) => s.trim())
    .filter(Boolean)
  if (raw.length === 0) return null

  // strip common module prefixes (api, v1, web, admin)
  let segments = raw.filter((s) => !PREFIX_SKIP.has(s.toLowerCase()))
  if (segments.length === 0) segments = raw

  // drop trailing action (index, show, store, update, destroy, etc)
  if (segments.length > 1 && isCommonAction(segments[segments.length - 1].toLowerCase())) {
    segments = segments.slice(0, -1)
  }

  if (segments.length === 0) return null

  // pick the deepest "module" — second-to-last when nested, else the only one
  const pick = segments.length >= 2 ? segments[segments.length - 2] : segments[0]
  return prettifySegment(pick)
}

const VERB_PREFIXES = new Set([
  'get',
  'list',
  'find',
  'fetch',
  'show',
  'index',
  'store',
  'create',
  'update',
  'destroy',
  'delete',
  'remove',
  'put',
  'patch',
  'post',
])

function prettifySegment(value: string): string {
  if (!value) return value
  let parts = splitWords(value)
  if (parts.length > 1 && VERB_PREFIXES.has(parts[0].toLowerCase())) {
    parts = parts.slice(1)
  }
  if (parts.length === 0) return value.toUpperCase()
  return parts
    .map((p) => p.charAt(0).toUpperCase() + p.slice(1).toLowerCase())
    .join(' ')
}

function splitWords(value: string): string[] {
  return value
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
    .split(/[\s\-_]+/)
    .filter(Boolean)
}

function fromController(handler?: string): string | null {
  if (!handler) return null
  const cleaned = handler.split('@')[0]
  if (!cleaned || cleaned.toLowerCase() === 'closure') return null
  const segments = cleaned.split('\\').filter(Boolean)
  if (segments.length === 0) return null
  const last = segments[segments.length - 1]
  const stripped = last.replace(/Controller$/i, '').trim()
  if (stripped) {
    return prettifySegment(stripped)
  }
  if (segments.length >= 2) {
    return prettifySegment(segments[segments.length - 2])
  }
  return null
}

function fromPath(path?: string): string | null {
  if (!path) return null
  const segments = path.split('/').filter(Boolean)
  for (const segment of segments) {
    const clean = segment.toLowerCase()
    if (clean.startsWith('{')) continue
    if (PREFIX_SKIP.has(clean)) continue
    return prettifySegment(segment)
  }
  return null
}

function isCommonAction(value: string): boolean {
  return [
    'index',
    'show',
    'store',
    'update',
    'destroy',
    'create',
    'edit',
    'list',
    'get',
    'save',
    'delete',
  ].includes(value)
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
