export function shortUrl(url: string): string {
  try {
    const u = new URL(url)
    return u.pathname + u.search
  } catch {
    return url
  }
}

export function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}

export function formatDate(date: Date): string {
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return `Today · ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  }
  return date.toLocaleString([], {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function prettyJSON(raw: string): string {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

export function formatBody(data: unknown): string {
  if (data == null) return ''
  if (typeof data === 'string') return data
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

export function formatLeaf(v: unknown): string {
  if (v === null || v === undefined) return v === null ? 'null' : ''
  if (typeof v === 'string') return `"${v}"`
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

export function valueTone(v: unknown): string {
  if (v === null || v === undefined) return 'text-muted-foreground/60 italic'
  if (typeof v === 'string') return 'text-emerald-500/90'
  if (typeof v === 'number') return 'text-purple-400'
  if (typeof v === 'boolean') return 'text-amber-500'
  return 'text-foreground/85'
}

export function statusTone(status: number, error?: string | null): string {
  if (error) return 'text-destructive'
  if (status >= 500) return 'text-destructive'
  if (status >= 400) return 'text-amber-500'
  if (status >= 200) return 'text-emerald-500'
  return 'text-muted-foreground'
}

export interface TableRows {
  columns: string[]
  data: unknown[][]
}

export function extractTableRows(value: unknown): TableRows | null {
  const arr = findArrayOfObjects(value)
  if (!arr || arr.length === 0) return null
  const columns = Array.from(
    new Set(arr.flatMap((row) => (row && typeof row === 'object' ? Object.keys(row as object) : []))),
  )
  if (columns.length === 0) return null
  const data = arr.map((row) =>
    columns.map((c) => (row && typeof row === 'object' ? (row as Record<string, unknown>)[c] : undefined)),
  )
  return { columns, data }
}

function findArrayOfObjects(value: unknown, depth = 0): unknown[] | null {
  if (depth > 6) return null
  if (Array.isArray(value)) {
    if (value.length > 0 && typeof value[0] === 'object' && value[0] !== null && !Array.isArray(value[0])) {
      return value
    }
    return null
  }
  if (value && typeof value === 'object') {
    for (const v of Object.values(value as object)) {
      const found = findArrayOfObjects(v, depth + 1)
      if (found) return found
    }
  }
  return null
}
