import type { DashboardMetrics } from '@/services/metricsService'

interface Props {
  metrics: DashboardMetrics | null
  discovery: { totalEndpoints: number; usedEndpoints: number; coverage: number } | null
}

export function KPIStrip({ metrics, discovery }: Props) {
  const errorRate = metrics?.errorRate ?? 0
  const stats = [
    { label: 'runs', value: (metrics?.totalRuns ?? 0).toLocaleString() },
    {
      label: 'errors',
      value: `${Math.round(errorRate * 100)}%`,
      accent: errorRate > 0.1 ? 'rose' : undefined,
    },
    { label: 'avg', value: `${metrics?.latency.avg ?? 0}ms` },
    { label: 'p95', value: `${metrics?.latency.p95 ?? 0}ms` },
    {
      label: 'coverage',
      value: `${Math.round((discovery?.coverage ?? 0) * 100)}%`,
      sub: discovery ? `${discovery.usedEndpoints}/${discovery.totalEndpoints}` : '',
    },
  ]
  return (
    <div className="flex items-center gap-6 flex-wrap">
      {stats.map((s, i) => (
        <div key={s.label} className="flex items-baseline gap-1.5">
          <span
            className={`text-[18px] font-semibold tabular-nums ${
              s.accent === 'rose' ? 'text-rose-500' : 'text-foreground'
            }`}
          >
            {s.value}
          </span>
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground/60">
            {s.label}
          </span>
          {s.sub && (
            <span className="text-[10px] font-mono text-muted-foreground/50 tabular-nums">{s.sub}</span>
          )}
          {i < stats.length - 1 && <span className="text-muted-foreground/20 ml-2">·</span>}
        </div>
      ))}
    </div>
  )
}
