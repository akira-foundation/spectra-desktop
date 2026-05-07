import { Copy } from 'lucide-react'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

interface EndpointHeaderProps {
  method: string
  path: string
  statusCode: number
  responseTime: string
  responseSize: string
}

export function EndpointHeader({
  method,
  path,
  statusCode,
  responseTime,
  responseSize,
}: EndpointHeaderProps) {
  const { getMethodColor } = useHttpMethod()
  const ok = statusCode >= 200 && statusCode < 300

  return (
    <div className="h-10 px-3 border-b border-border/60 bg-card/30 flex items-center justify-between">
      <div className="flex items-center gap-2 min-w-0">
        <span
          className={cn(
            'inline-flex w-12 shrink-0 justify-center text-[10px] font-bold tracking-wider rounded px-1 py-0.5',
            getMethodColor(method),
          )}
        >
          {method}
        </span>
        <span className="font-mono text-[12px] text-foreground truncate">{path}</span>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <span
          className={cn(
            'inline-flex items-center gap-1 text-[10.5px] font-mono px-1.5 py-0.5 rounded border',
            ok
              ? 'text-emerald-500 border-emerald-500/30 bg-emerald-500/10'
              : 'text-destructive border-destructive/30 bg-destructive/10',
          )}
        >
          <span className="size-1 rounded-full bg-current" />
          {statusCode}
        </span>
        <span className="text-[11px] text-muted-foreground tabular-nums">{responseTime}</span>
        <span className="text-[11px] text-muted-foreground tabular-nums">{responseSize}</span>
        <button className="h-6 px-2 text-[10.5px] inline-flex items-center gap-1 rounded text-muted-foreground hover:text-foreground hover:bg-accent/60 transition-colors">
          <Copy className="w-3 h-3" />
          Copy
        </button>
      </div>
    </div>
  )
}
