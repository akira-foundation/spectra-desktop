import { cn } from '@/lib/utils'

const IMPORTANT_HEADERS = new Set([
  'content-type',
  'authorization',
  'set-cookie',
  'cookie',
  'cache-control',
  'location',
  'www-authenticate',
  'x-request-id',
  'x-csrf-token',
  'etag',
])

function isImportantHeader(name: string): boolean {
  return IMPORTANT_HEADERS.has(name.toLowerCase())
}

interface Props {
  headers?: Record<string, string[]>
  highlightImportant?: boolean
}

export function HeadersList({ headers, highlightImportant = false }: Props) {
  const entries = Object.entries(headers ?? {})
  if (entries.length === 0) {
    return <p className="text-[11.5px] text-muted-foreground italic text-center">No headers</p>
  }
  if (highlightImportant) {
    return (
      <ul className="m-0 p-0 list-none divide-y divide-border/20">
        {entries.map(([k, vs]) => {
          const important = isImportantHeader(k)
          return (
            <li key={k} className="grid grid-cols-[140px_1fr] gap-3 px-3 py-1.5 hover:bg-accent/20">
              <code
                className={cn(
                  'text-[10.5px] font-mono truncate',
                  important ? 'text-foreground font-semibold' : 'text-foreground/70',
                )}
                title={k}
              >
                {k}
              </code>
              <code
                className={cn(
                  'text-[10.5px] font-mono break-all',
                  important ? 'text-foreground/90' : 'text-muted-foreground',
                )}
              >
                {vs.join(', ')}
              </code>
            </li>
          )
        })}
      </ul>
    )
  }
  return (
    <ul className="space-y-1">
      {entries.map(([k, vs]) => (
        <li key={k} className="flex gap-2 text-[11.5px] font-mono">
          <span className="text-muted-foreground/80 shrink-0 min-w-[140px]">{k}:</span>
          <span className="text-foreground/90 break-all">{vs.join(', ')}</span>
        </li>
      ))}
    </ul>
  )
}
