import { ResponsiveContainer, AreaChart, Area, Tooltip as RechartsTooltip, XAxis } from 'recharts'
import { TrendingUp, CalendarDays } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import type { DashboardMetrics } from '@/services/metricsService'
import { cn } from '@/lib/utils'
import { Card } from './Card'

interface Props {
  metrics: DashboardMetrics | null
  days: 7 | 14 | 30
  onChangeDays: (d: 7 | 14 | 30) => void
}

export function VolumeCard({ metrics, days, onChangeDays }: Props) {
  const volume = metrics?.volume ?? []
  const total = volume.reduce((s, v) => s + v.count, 0)
  return (
    <Card
      title={`Volume · ${days}d`}
      icon={TrendingUp}
      headerExtra={<VolumeRangePicker value={days} onChange={onChangeDays} />}
    >
      {total === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">No requests in last 7 days.</p>
      ) : (
        <div className="space-y-2">
          <div className="flex items-baseline gap-2">
            <span className="text-[20px] font-semibold tabular-nums">{total}</span>
            <span className="text-[10.5px] text-muted-foreground">total runs</span>
          </div>
          <div className="h-16">
            <ResponsiveContainer>
              <AreaChart data={volume} margin={{ top: 4, right: 0, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="volumeGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="rgb(168, 85, 247)" stopOpacity={0.5} />
                    <stop offset="100%" stopColor="rgb(168, 85, 247)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <RechartsTooltip
                  cursor={{ stroke: 'rgba(168, 85, 247, 0.3)', strokeWidth: 1 }}
                  contentStyle={{
                    background: 'rgba(24, 24, 27, 0.95)',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    fontSize: '11px',
                    borderRadius: 6,
                    padding: '4px 8px',
                    color: '#fafafa',
                  }}
                  labelStyle={{ color: 'rgba(255, 255, 255, 0.6)', fontSize: '10px' }}
                  formatter={(v) => [`${v} runs`, '']}
                />
                <XAxis dataKey="day" hide />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke="rgb(168, 85, 247)"
                  strokeWidth={2}
                  fill="url(#volumeGradient)"
                  dot={{ fill: 'rgb(168, 85, 247)', r: 2 }}
                  activeDot={{ fill: 'rgb(168, 85, 247)', r: 4 }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </Card>
  )
}

function VolumeRangePicker({ value, onChange }: { value: 7 | 14 | 30; onChange: (d: 7 | 14 | 30) => void }) {
  const options: Array<7 | 14 | 30> = [7, 14, 30]
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/40"
          aria-label="Volume range"
        >
          <CalendarDays className="w-3 h-3" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-32 p-1">
        {options.map((d) => (
          <button
            key={d}
            type="button"
            onClick={() => onChange(d)}
            className={cn(
              'w-full text-left px-2 py-1 rounded text-[11.5px] hover:bg-accent/40 transition-colors',
              value === d && 'bg-accent text-foreground',
            )}
          >
            Last {d} days
          </button>
        ))}
      </PopoverContent>
    </Popover>
  )
}
