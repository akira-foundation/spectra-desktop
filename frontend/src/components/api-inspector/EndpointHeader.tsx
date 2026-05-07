import { Badge } from '@/components/ui/badge'
import { useHttpMethod } from '@/hooks/useHttpMethod'

interface EndpointHeaderProps {
  method: string
  path: string
  statusCode: number
  responseTime: string
  responseSize: string
}

export function EndpointHeader({ method, path, statusCode, responseTime, responseSize }: EndpointHeaderProps) {
  const { getMethodColor } = useHttpMethod()

  return (
    <div className="p-4 border-b border-border/50 bg-card/30">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge className={`font-bold ${getMethodColor(method)}`}>{method}</Badge>
          <span className="text-sm font-mono">{path}</span>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/20">
              {statusCode}
            </Badge>
            <span className="text-xs text-muted-foreground">{responseTime}</span>
            <span className="text-xs text-muted-foreground">{responseSize}</span>
          </div>
          <button className="px-3 py-1 text-xs bg-primary/10 hover:bg-primary/20 text-primary rounded-md">
            Copy
          </button>
        </div>
      </div>
    </div>
  )
}
