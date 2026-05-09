import { useEffect, useMemo, useState } from 'react'
import { TrendingUp, Calendar, AlertTriangle, Zap, XCircle } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { LineChart, Line, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { insightsService, type Insights, type FlakyEndpoint, type HourlyCell } from '@/services/insightsService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

const SERIES_COLORS = ['#a855f7', '#22d3ee', '#10b981', '#f59e0b', '#ef4444']
const DOW_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

interface Props {
  projectId: string | null
  days: number
  refreshKey?: number
}

export function InsightsSection({ projectId, days, refreshKey }: Props) {
  const [data, setData] = useState<Insights | null>(null)

  useEffect(() => {
    if (!projectId) return
    void insightsService.get(projectId, days).then(setData)
  }, [projectId, days, refreshKey])

  if (!projectId) return null

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <SeriesCard
          title="Slowest endpoints"
          icon={TrendingUp}
          unit="ms"
          series={(data?.latencyOverTime ?? []).map((s) => ({
            id: s.endpointID,
            method: s.method,
            path: s.path,
            value: s.avgMs,
            points: s.points.map((p) => ({ day: p.day, value: p.avgMs })),
          }))}
        />
        <SeriesCard
          title="Most used"
          icon={Zap}
          unit=""
          series={(data?.usageOverTime ?? []).map((s) => ({
            id: s.endpointID,
            method: s.method,
            path: s.path,
            value: s.total,
            points: s.points.map((p) => ({ day: p.day, value: p.count })),
          }))}
        />
        <SeriesCard
          title="Failures"
          icon={XCircle}
          unit=""
          series={(data?.failuresOverTime ?? []).map((s) => ({
            id: s.endpointID,
            method: s.method,
            path: s.path,
            value: s.failures,
            points: s.points.map((p) => ({ day: p.day, value: p.count })),
          }))}
          colors={['#ef4444', '#f97316', '#f59e0b', '#eab308', '#a855f7']}
        />
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div className="lg:col-span-2">
          <HeatmapCard cells={data?.hourlyHeatmap ?? []} />
        </div>
        <FlakyCard items={data?.flaky ?? []} />
      </div>
    </div>
  )
}

interface UnifiedSeries {
  id: string
  method: string
  path: string
  value: number
  points: { day: string; value: number }[]
}

function CardShell({
  title,
  icon: Icon,
  children,
  empty,
  emptyMessage,
}: {
  title: string
  icon: LucideIcon
  children: React.ReactNode
  empty?: boolean
  emptyMessage?: string
}) {
  return (
    <div className="rounded-md border border-border/40 bg-card/30 p-4 flex flex-col min-h-0">
      <div className="flex items-center gap-1.5 mb-3">
        <Icon className="w-3 h-3 text-muted-foreground" />
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          {title}
        </h3>
      </div>
      {empty ? (
        <p className="text-[11px] italic text-muted-foreground/60 text-center py-8">
          {emptyMessage || 'Not enough data yet.'}
        </p>
      ) : (
        children
      )}
    </div>
  )
}

function SeriesCard({
  title,
  icon,
  unit,
  series,
  colors = SERIES_COLORS,
}: {
  title: string
  icon: LucideIcon
  unit: string
  series: UnifiedSeries[]
  colors?: string[]
}) {
  const { getMethodColor } = useHttpMethod()

  const chartData = useMemo(() => {
    if (series.length === 0) return []
    const days = series[0]?.points ?? []
    return days.map((_, i) => {
      const row: Record<string, any> = { day: days[i].day.slice(5) }
      for (const s of series) {
        row[s.id] = s.points[i]?.value ?? 0
      }
      return row
    })
  }, [series])

  return (
    <CardShell title={title} icon={icon} empty={series.length === 0}>
      <div className="h-32">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <XAxis dataKey="day" stroke="var(--muted-foreground)" tick={{ fontSize: 9 }} />
            <YAxis stroke="var(--muted-foreground)" tick={{ fontSize: 9 }} width={32} unit={unit} />
            <Tooltip
              contentStyle={{
                background: 'var(--popover)',
                border: '1px solid var(--border)',
                borderRadius: 6,
                fontSize: 10,
              }}
            />
            {series.map((s, i) => (
              <Line
                key={s.id}
                type="monotone"
                dataKey={s.id}
                stroke={colors[i % colors.length]}
                strokeWidth={1.5}
                dot={false}
                name={`${s.method} ${s.path}`}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>
      <ul className="m-0 p-0 list-none mt-2 space-y-1">
        {series.map((s, i) => (
          <li key={s.id} className="flex items-center gap-2 text-[10.5px] min-w-0">
            <span
              className="inline-block w-2 h-2 rounded-full shrink-0"
              style={{ background: colors[i % colors.length] }}
            />
            <span
              className={cn(
                'inline-flex w-10 shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
                getMethodColor(s.method),
              )}
            >
              {s.method}
            </span>
            <code className="font-mono truncate flex-1 text-foreground/80">{s.path}</code>
            <span className="font-mono tabular-nums text-muted-foreground shrink-0">
              {s.value}
              {unit}
            </span>
          </li>
        ))}
      </ul>
    </CardShell>
  )
}

function HeatmapCard({ cells }: { cells: HourlyCell[] }) {
  const max = useMemo(() => {
    let m = 0
    for (const c of cells) if (c.count > m) m = c.count
    return m
  }, [cells])

  const grid = useMemo(() => {
    const g: number[][] = Array.from({ length: 7 }, () => Array(24).fill(0))
    for (const c of cells) g[c.day][c.hour] = c.count
    return g
  }, [cells])

  const opacity = (n: number) => (max > 0 ? Math.min(1, n / max) : 0)

  return (
    <CardShell title="Usage by hour" icon={Calendar} empty={max === 0}>
      <div className="overflow-x-auto">
        <table className="w-full border-separate border-spacing-[1px]">
          <thead>
            <tr>
              <th className="w-6"></th>
              {[0, 6, 12, 18].map((h) => (
                <th key={h} colSpan={6} className="text-[8.5px] font-mono text-muted-foreground/60 text-left">
                  {h.toString().padStart(2, '0')}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {grid.map((row, d) => (
              <tr key={d}>
                <td className="text-[8.5px] font-mono text-muted-foreground/60 pr-1 align-middle">{DOW_LABELS[d]}</td>
                {row.map((count, h) => (
                  <td
                    key={h}
                    title={`${DOW_LABELS[d]} ${h.toString().padStart(2, '0')}:00 — ${count} request${count === 1 ? '' : 's'}`}
                    className="rounded-sm"
                    style={{
                      width: 12,
                      height: 12,
                      background:
                        count > 0
                          ? `rgba(168, 85, 247, ${0.15 + 0.85 * opacity(count)})`
                          : 'rgba(255, 255, 255, 0.04)',
                    }}
                  />
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="text-[10px] text-muted-foreground/60 mt-2 text-right">peak {max}/hr</p>
    </CardShell>
  )
}

function FlakyCard({ items }: { items: FlakyEndpoint[] }) {
  const { getMethodColor } = useHttpMethod()
  return (
    <CardShell title="Flaky endpoints" icon={AlertTriangle} empty={items.length === 0} emptyMessage="No flakiness detected.">
      <ul className="m-0 p-0 list-none space-y-1.5">
        {items.slice(0, 6).map((f) => (
          <li key={f.endpointID} className="flex items-center gap-2 text-[10.5px] min-w-0">
            <div className="flex-1 min-w-0 flex items-center gap-2">
              <span
                className={cn(
                  'inline-flex w-10 shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
                  getMethodColor(f.method),
                )}
              >
                {f.method}
              </span>
              <code className="font-mono truncate text-foreground/85">{f.path}</code>
            </div>
            <div className="flex items-center gap-1.5 shrink-0">
              <span className="font-mono text-emerald-500/80 tabular-nums">{f.successes}</span>
              <span className="text-muted-foreground/40">/</span>
              <span className="font-mono text-rose-500/80 tabular-nums">{f.failures}</span>
              <FlakeBar score={f.flakeScore} />
            </div>
          </li>
        ))}
      </ul>
    </CardShell>
  )
}

function FlakeBar({ score }: { score: number }) {
  const pct = Math.round(score * 100)
  return (
    <div className="w-12 h-1.5 rounded-full bg-muted/50 overflow-hidden" title={`${pct}% flaky`}>
      <div className="h-full bg-gradient-to-r from-amber-500 to-rose-500" style={{ width: `${pct}%` }} />
    </div>
  )
}
