import { useHttpMethod } from '@/hooks/useHttpMethod'
import { MockSourceBadge } from './MockSourceBadge'
import type { MockLogEvent } from '@/services/mockService'
import { cn } from '@/lib/utils'

interface Props {
  logs: MockLogEvent[]
}

export function MockLogList({ logs }: Props) {
  const { getMethodColor } = useHttpMethod()

  if (logs.length === 0) {
    return (
      <p className="px-3.5 py-6 text-center text-[11.5px] italic text-muted-foreground/70">
        No requests yet. Hit the URL above to see live activity.
      </p>
    )
  }

  return (
    <ul className="divide-y divide-border/40 max-h-[40vh] overflow-auto">
      {logs.map((log, i) => (
        <li
          key={`${log.timestamp}-${i}`}
          className="px-3.5 py-2 flex items-center gap-2.5 text-[11.5px] hover:bg-accent/20"
        >
          <span className="text-[10.5px] tabular-nums text-muted-foreground/80 shrink-0 w-16">
            {formatTime(log.timestamp)}
          </span>
          <span
            className={cn(
              'inline-flex shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1.5 py-px w-12',
              getMethodColor(log.method),
            )}
          >
            {log.method}
          </span>
          <span className="font-mono truncate flex-1 min-w-0">{log.path}</span>
          <span
            className={cn(
              'tabular-nums text-[10.5px] shrink-0',
              log.status >= 400
                ? 'text-rose-500'
                : log.status >= 300
                  ? 'text-amber-500'
                  : 'text-emerald-500',
            )}
          >
            {log.status}
          </span>
          <span className="text-[10.5px] text-muted-foreground tabular-nums shrink-0 w-12 text-right">
            {log.durationMs}ms
          </span>
          <MockSourceBadge source={log.source} />
        </li>
      ))}
    </ul>
  )
}

function formatTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString(undefined, {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return ''
  }
}
