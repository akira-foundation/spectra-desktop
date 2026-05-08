const ROUTE_PARAM_REGEX = /\{([^}]+)\}/g

export function extractRouteParams(path: string): string[] {
  if (!path) return []
  const matches = path.matchAll(ROUTE_PARAM_REGEX)
  const result: string[] = []
  for (const match of matches) {
    result.push(match[1])
  }
  return result
}

export function resolveRoutePath(path: string, values: string[]): string {
  if (!path) return ''
  let i = 0
  return path.replace(ROUTE_PARAM_REGEX, (match) => {
    const value = values[i++]
    return value && value.length > 0 ? value : match
  })
}

export interface QueryParam {
  key: string
  value: string
}

export function buildQueryString(params: QueryParam[]): string {
  const filled = params.filter((p) => p.key.trim().length > 0)
  if (filled.length === 0) return ''
  const search = filled
    .map((p) => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`)
    .join('&')
  return `?${search}`
}
