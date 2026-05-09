import { useEffect, useMemo, useState } from 'react'
import { TrendingUp, Calendar, AlertTriangle, Zap, XCircle } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { LineChart, Line, ResponsiveContainer, Tooltip, XAxis, YAxis, RadarChart, PolarGrid, PolarAngleAxis, Radar } from 'recharts'
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
          <HeatmapCard cells={data?.hourlyHeatmap ?? []} methodShare={data?.methodShare ?? []} />
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

function HeatmapCard({ cells, methodShare }: { cells: HourlyCell[]; methodShare: { method: string; count: number; percent: number }[] }) {
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

  const HEAT_LEVELS = [
    'rgba(255, 255, 255, 0.05)',
    'rgba(168, 85, 247, 0.25)',
    'rgba(168, 85, 247, 0.50)',
    'rgba(168, 85, 247, 0.75)',
    'rgba(168, 85, 247, 1.00)',
  ]
  const levelOf = (n: number) => {
    if (n === 0 || max === 0) return 0
    const ratio = n / max
    if (ratio < 0.25) return 1
    if (ratio < 0.5) return 2
    if (ratio < 0.75) return 3
    return 4
  }

  const totalReqs = useMemo(() => cells.reduce((s, c) => s + c.count, 0), [cells])
  const busiestHour = useMemo(() => {
    let best = { hour: 0, count: 0 }
    const byHour: Record<number, number> = {}
    for (const c of cells) byHour[c.hour] = (byHour[c.hour] ?? 0) + c.count
    for (const [h, n] of Object.entries(byHour)) {
      if (n > best.count) best = { hour: Number(h), count: n }
    }
    return best
  }, [cells])
  const busiestDay = useMemo(() => {
    let best = { day: 0, count: 0 }
    const byDay: Record<number, number> = {}
    for (const c of cells) byDay[c.day] = (byDay[c.day] ?? 0) + c.count
    for (const [d, n] of Object.entries(byDay)) {
      if (n > best.count) best = { day: Number(d), count: n }
    }
    return best
  }, [cells])

  return (
    <CardShell title="Usage by hour" icon={Calendar} empty={max === 0}>
      <div className="grid grid-cols-3 gap-3 mb-4">
        <Stat label="Total" value={totalReqs.toString()} />
        <Stat label="Peak hour" value={`${busiestHour.hour.toString().padStart(2, '0')}:00`} sub={`${busiestHour.count} req`} />
        <Stat label="Peak day" value={DOW_LABELS[busiestDay.day]} sub={`${busiestDay.count} req`} />
      </div>
      <div className="grid grid-cols-[1fr_280px] gap-6 items-stretch">
      <div className="flex flex-col gap-[3px] min-w-0">
        <div className="flex items-center text-[9px] font-mono text-muted-foreground/60 pl-7">
          {Array.from({ length: 24 }).map((_, h) => (
            <span key={h} style={{ width: 13 }} className="shrink-0">
              {h % 6 === 0 ? h.toString().padStart(2, '0') : ''}
            </span>
          ))}
        </div>
        {grid.map((row, d) => (
          <div key={d} className="flex items-center gap-[3px]">
            <span className="text-[9px] font-mono text-muted-foreground/60 w-6 text-right pr-1">
              {DOW_LABELS[d]}
            </span>
            {row.map((count, h) => (
              <div
                key={h}
                title={`${DOW_LABELS[d]} ${h.toString().padStart(2, '0')}:00 — ${count} request${count === 1 ? '' : 's'}`}
                className="rounded-full shrink-0"
                style={{
                  width: 10,
                  height: 10,
                  background: HEAT_LEVELS[levelOf(count)],
                }}
              />
            ))}
          </div>
        ))}
      </div>
      <MethodRadar shares={methodShare} />
      </div>
      <div className="flex items-center justify-between mt-3 text-[10px] text-muted-foreground/60">
        <span>peak {max}/hr</span>
        <div className="flex items-center gap-1.5">
          <span>Less</span>
          <div className="flex items-center gap-[2px]">
            {HEAT_LEVELS.map((bg, i) => (
              <span
                key={i}
                className="rounded-full"
                style={{ width: 10, height: 10, background: bg, display: 'inline-block' }}
              />
            ))}
          </div>
          <span>More</span>
        </div>
      </div>
    </CardShell>
  )
}

function MethodRadar({ shares }: { shares: { method: string; count: number; percent: number }[] }) {
  const data = useMemo(() => {
    const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
    const map = new Map(shares.map((s) => [s.method.toUpperCase(), s]))
    return methods.map((m) => ({
      method: m,
      value: Math.round((map.get(m)?.percent ?? 0) * 100),
    }))
  }, [shares])

  const total = data.reduce((sum, d) => sum + d.value, 0)
  if (total === 0) return null

  return (
    <div className="w-full h-full flex flex-col gap-2">
      <span className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/60">
        By method
      </span>
      <div className="flex-1 min-h-[160px] w-full">
        <ResponsiveContainer width="100%" height="100%">
          <RadarChart data={data} outerRadius="75%">
            <PolarGrid stroke="var(--border)" strokeOpacity={0.4} />
            <PolarAngleAxis
              dataKey="method"
              tick={{ fill: 'var(--muted-foreground)', fontSize: 9 }}
            />
            <Radar
              dataKey="value"
              stroke="rgb(168, 85, 247)"
              fill="rgb(168, 85, 247)"
              fillOpacity={0.35}
              strokeWidth={1.5}
            />
          </RadarChart>
        </ResponsiveContainer>
      </div>
      <ul className="m-0 p-0 list-none w-full space-y-0.5 text-[9.5px]">
        {data.map((d) => (
          <li key={d.method} className="flex items-center justify-between gap-2 font-mono">
            <span className="text-muted-foreground/80 w-12">{d.method}</span>
            <div className="flex-1 h-1 rounded-full bg-muted/40 overflow-hidden">
              <div className="h-full bg-primary/60" style={{ width: `${d.value}%` }} />
            </div>
            <span className="text-foreground/80 tabular-nums w-8 text-right">{d.value}%</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function FlakyCard({ items }: { items: FlakyEndpoint[] }) {
  const { getMethodColor } = useHttpMethod()
  const totals = useMemo(() => {
    let s = 0
    let f = 0
    for (const it of items) {
      s += it.successes
      f += it.failures
    }
    return { s, f, total: s + f }
  }, [items])
  const overallPassPct = totals.total > 0 ? (totals.s / totals.total) * 100 : 0
  return (
    <CardShell title="Flaky endpoints" icon={AlertTriangle} empty={items.length === 0} emptyMessage="No flakiness detected.">
      <div className="flex flex-col h-full gap-3">
        <ul className="m-0 p-0 list-none space-y-3 flex-1">
          {items.slice(0, 6).map((f) => {
            const total = f.successes + f.failures
            const passPct = total > 0 ? (f.successes / total) * 100 : 0
            return (
              <li key={f.endpointID} className="space-y-1.5">
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className={cn(
                      'inline-flex w-10 shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
                      getMethodColor(f.method),
                    )}
                  >
                    {f.method}
                  </span>
                  <code className="text-[10.5px] font-mono truncate flex-1 text-foreground/85">{f.path}</code>
                  <span className="text-[10px] font-mono tabular-nums shrink-0 text-muted-foreground/70">
                    {total} runs
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="flex-1 flex h-2 rounded-full overflow-hidden bg-muted/40">
                    <div className="bg-emerald-500/80" style={{ width: `${passPct}%` }} title={`${f.successes} pass`} />
                    <div className="bg-rose-500/80" style={{ width: `${100 - passPct}%` }} title={`${f.failures} fail`} />
                  </div>
                  <span className="text-[9.5px] font-mono tabular-nums shrink-0 w-14 text-right">
                    <span className="text-emerald-500/90">{f.successes}</span>
                    <span className="text-muted-foreground/40 mx-0.5">/</span>
                    <span className="text-rose-500/90">{f.failures}</span>
                  </span>
                </div>
              </li>
            )
          })}
        </ul>
        {totals.total > 0 && (
          <div className="border-t border-border/30 pt-3 mt-auto">
            <div className="flex items-baseline justify-between mb-1.5">
              <span className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/60">
                Overall · {totals.total} runs
              </span>
              <span className="text-[11px] font-mono tabular-nums">
                <span className="text-emerald-500/90">{Math.round(overallPassPct)}%</span>
                <span className="text-muted-foreground/50 mx-1">pass</span>
              </span>
            </div>
            <div className="flex h-2 rounded-full overflow-hidden bg-muted/40">
              <div className="bg-emerald-500/80" style={{ width: `${overallPassPct}%` }} />
              <div className="bg-rose-500/80" style={{ width: `${100 - overallPassPct}%` }} />
            </div>
          </div>
        )}
      </div>
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

function Stat({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="rounded-md border border-border/30 bg-muted/15 px-2.5 py-1.5">
      <div className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/60">{label}</div>
      <div className="text-[14px] font-semibold tabular-nums leading-tight">{value}</div>
      {sub && <div className="text-[9.5px] font-mono text-muted-foreground/70 tabular-nums">{sub}</div>}
    </div>
  )
}
