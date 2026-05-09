import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts'
import { Activity } from 'lucide-react'
import type { DashboardMetrics } from '@/services/metricsService'
import { cn } from '@/lib/utils'
import { Card } from './Card'

const STATUS_COLORS: Record<string, string> = {
  '2xx': '#10b981',
  '3xx': '#3b82f6',
  '4xx': '#f59e0b',
  '5xx': '#ef4444',
  err: '#ef4444',
}

export function StatusCard({ metrics }: { metrics: DashboardMetrics | null }) {
  const buckets = (metrics?.statusBuckets ?? []).filter((b) => b.count > 0)
  const total = buckets.reduce((sum, b) => sum + b.count, 0)
  const errorRate = metrics ? metrics.errorRate * 100 : 0
  return (
    <Card title="Response status" icon={Activity}>
      {total === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">No requests yet.</p>
      ) : (
        <div className="flex items-center gap-3">
          <div className="w-24 h-24 shrink-0 relative">
            <ResponsiveContainer>
              <PieChart>
                <Pie
                  data={buckets}
                  dataKey="count"
                  innerRadius={28}
                  outerRadius={42}
                  paddingAngle={2}
                  stroke="none"
                >
                  {buckets.map((b) => (
                    <Cell key={b.bucket} fill={STATUS_COLORS[b.bucket] ?? '#6b7280'} />
                  ))}
                </Pie>
              </PieChart>
            </ResponsiveContainer>
            <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
              <span className="text-[14px] font-semibold tabular-nums">{total}</span>
              <span className="text-[9px] text-muted-foreground uppercase tracking-wider">runs</span>
            </div>
          </div>
          <div className="flex-1 space-y-1">
            {buckets.map((b) => (
              <div key={b.bucket} className="flex items-center gap-2 text-[11px]">
                <span
                  className="w-2 h-2 rounded-full"
                  style={{ background: STATUS_COLORS[b.bucket] ?? '#6b7280' }}
                />
                <span className="font-mono uppercase">{b.bucket}</span>
                <span className="ml-auto tabular-nums text-muted-foreground">{b.count}</span>
              </div>
            ))}
            <div className="pt-1 mt-1 border-t border-border/40 flex items-center justify-between text-[10.5px]">
              <span className="text-muted-foreground">Error rate</span>
              <span
                className={cn(
                  'font-mono tabular-nums',
                  errorRate > 10 ? 'text-rose-500' : errorRate > 0 ? 'text-amber-500' : 'text-emerald-500',
                )}
              >
                {errorRate.toFixed(1)}%
              </span>
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}
