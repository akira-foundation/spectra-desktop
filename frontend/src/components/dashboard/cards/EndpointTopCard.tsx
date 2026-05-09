import type { LucideIcon } from 'lucide-react'
import type { EndpointMetricDTO } from '@/services/metricsService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { Card } from './Card'

interface Props {
  title: string
  icon: LucideIcon
  entries: EndpointMetricDTO[]
  metric: 'ms' | 'count' | 'errorRate'
  onOpen: (id: string) => void
}

export function EndpointTopCard({ title, icon, entries, metric, onOpen }: Props) {
  const { getMethodColor } = useHttpMethod()
  const formatMetric = (e: EndpointMetricDTO) => {
    switch (metric) {
      case 'ms':
        return `${e.avgMs}ms`
      case 'count':
        return `${e.count}`
      case 'errorRate':
        return `${(e.errorRate * 100).toFixed(0)}%`
    }
  }
  return (
    <Card title={title} icon={icon}>
      {entries.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">No data yet.</p>
      ) : (
        <ul className="space-y-px">
          {entries.map((e) => (
            <li key={e.endpointID}>
              <button
                type="button"
                onClick={() => onOpen(e.endpointID)}
                className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 transition-colors"
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(e.method),
                  )}
                >
                  {e.method}
                </span>
                <span className="text-[11.5px] font-mono truncate flex-1 text-left text-foreground/85">
                  {e.path}
                </span>
                <span className="text-[10.5px] font-mono tabular-nums text-foreground/85 shrink-0">
                  {formatMetric(e)}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
