import { Timer } from 'lucide-react'
import type { DashboardMetrics } from '@/services/metricsService'
import { cn } from '@/lib/utils'
import { Card } from './Card'

const LATENCY_LEGEND: Record<string, string> = {
  p50: 'Median — half of requests faster',
  p95: '95% of requests faster',
  p99: '99% of requests faster (worst-case tail)',
}

export function LatencyCard({ metrics }: { metrics: DashboardMetrics | null }) {
  const lat = metrics?.latency
  const has = lat && lat.count > 0
  return (
    <Card title="Latency" icon={Timer}>
      {has ? (
        <div className="space-y-2">
          <div className="grid grid-cols-3 gap-2">
            <LatencyStat label="p50" value={lat.p50} tone="emerald" />
            <LatencyStat label="p95" value={lat.p95} tone="amber" />
            <LatencyStat label="p99" value={lat.p99} tone="rose" />
          </div>
          <div className="flex items-center justify-between text-[10.5px] text-muted-foreground pt-1 border-t border-border/40">
            <span>Avg {lat.avg}ms</span>
            <span>Min {lat.min}ms</span>
            <span>Max {lat.max}ms</span>
          </div>
          <p className="text-[10px] text-muted-foreground/70 leading-relaxed">
            <span className="text-emerald-500/80">p50</span> typical ·{' '}
            <span className="text-amber-500/80">p95</span> slow tail ·{' '}
            <span className="text-rose-500/80">p99</span> worst case
          </p>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">No latency data yet.</p>
      )}
    </Card>
  )
}

function LatencyStat({ label, value, tone }: { label: string; value: number; tone: 'emerald' | 'amber' | 'rose' }) {
  const toneClass = {
    emerald: 'text-emerald-500',
    amber: 'text-amber-500',
    rose: 'text-rose-500',
  }[tone]
  return (
    <div className="rounded-md border border-border/40 bg-muted/20 px-2 py-1.5" title={LATENCY_LEGEND[label]}>
      <p className="text-[9.5px] uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className={cn('text-[14px] font-semibold tabular-nums', toneClass)}>
        {value}
        <span className="text-[10px] text-muted-foreground/80 ml-0.5">ms</span>
      </p>
    </div>
  )
}
