import { useEffect, useMemo, useState } from 'react'
import { PieChart, Pie, Cell, ResponsiveContainer, AreaChart, Area, Tooltip as RechartsTooltip, XAxis } from 'recharts'
import { metricsService, type DashboardMetrics, type EndpointMetricDTO } from '@/services/metricsService'
import { useProjectStore } from '@/store/projectStore'
import { useStatsStore } from '@/store/statsStore'
import { useAuthStore } from '@/store/authStore'
import { useEnvironmentStore } from '@/store/environmentStore'
import { useHistoryStore } from '@/store/historyStore'
import { useChangelogStore } from '@/store/changelogStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useUIStore } from '@/store/uiStore'
import {
  Navigation,
  Code,
  Database,
  Cpu,
  AlertCircle,
  FileCheck,
  Briefcase,
  Mail,
  Wrench,
  KeyRound,
  Layers,
  Star,
  History,
  GitCompare,
  ArrowUpRight,
  Plus,
  Minus,
  Pencil,
  Activity,
  Timer,
  AlertTriangle,
  Zap,
  TrendingUp,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Welcome } from '@/components/pages/Welcome'
import { Skeleton } from '@/components/ui/skeleton'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import type { StatCard, ScannedEndpoint } from '@/services/scannerService'
import { cn } from '@/lib/utils'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []

const KIND_ICON: Record<string, LucideIcon> = {
  routes: Navigation,
  controllers: Code,
  middleware: Cpu,
  models: Database,
  form_requests: FileCheck,
  jobs: Briefcase,
  mailers: Mail,
  services: Wrench,
  errors: AlertCircle,
}

export function Dashboard() {
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const projects = useProjectStore((state) => state.projects)
  const activeProject = projects.find((p) => p.id === activeProjectId)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)

  const loadReport = useStatsStore((s) => s.loadReport)
  const report = useStatsStore((s) =>
    activeProjectId ? s.reportByProject[activeProjectId] ?? null : null,
  )
  const auth = useAuthStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const loadAuth = useAuthStore((s) => s.load)

  const envs = useEnvironmentStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const loadEnvs = useEnvironmentStore((s) => s.load)
  const activeEnv = envs?.find((e) => e.id === activeProject?.activeEnvironmentId) ?? null

  const history = useHistoryStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const loadHistory = useHistoryStore((s) => s.load)

  const snapshots = useChangelogStore((s) =>
    activeProjectId ? s.snapshotsByProject[activeProjectId] ?? null : null,
  )
  const loadSnapshots = useChangelogStore((s) => s.load)

  const allEndpoints = useEndpointsStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
  )

  const pinnedKeys = useUIStore((s) =>
    activeProjectId ? s.pinnedEndpointsByProject[activeProjectId] ?? null : null,
  )
  const setSelectedEndpoint = useUIStore((s) => s.setSelectedEndpoint)

  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)

  useEffect(() => {
    if (!activeProjectId) return
    void loadReport(activeProjectId)
    void loadAuth(activeProjectId)
    void loadEnvs(activeProjectId)
    void loadHistory(activeProjectId)
    void loadSnapshots(activeProjectId)
    void metricsService.get(activeProjectId).then((m) => setMetrics(m))
  }, [activeProjectId, loadReport, loadAuth, loadEnvs, loadHistory, loadSnapshots])

  const pinnedEndpoints = useMemo(() => {
    const set = new Set(pinnedKeys ?? [])
    return allEndpoints.filter((e) => set.has(`${e.method} ${e.path}`)).slice(0, 8)
  }, [allEndpoints, pinnedKeys])

  if (!activeProject) return <Welcome />

  const cards = report?.cards ?? []
  const recent = (history ?? []).slice(0, 5)
  const latestSnapshot = snapshots?.[0]

  const goToInspector = (endpointId?: string) => {
    if (endpointId && activeProjectId) {
      setSelectedEndpoint(activeProjectId, endpointId)
    }
    setCurrentPage('inspector')
  }

  return (
    <div className="space-y-5 p-6">
      <header className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-md bg-primary/10 flex items-center justify-center text-primary font-semibold text-base">
          {activeProject.name[0]?.toUpperCase()}
        </div>
        <div className="flex-1">
          <h1 className="text-xl font-semibold tracking-tight capitalize">{activeProject.name}</h1>
          <p className="text-foreground/60 text-[12px]">
            {activeProject.framework}
            {activeProject.frameworkVersion ? ` · ${activeProject.frameworkVersion}` : ''}
            {activeProject.baseUrl ? ` · ${activeProject.baseUrl}` : ''}
          </p>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <StatusCard metrics={metrics} />
        <LatencyCard metrics={metrics} />
        <VolumeCard metrics={metrics} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <EndpointTopCard
          title="Slowest"
          icon={Timer}
          entries={metrics?.topSlow ?? []}
          metric="ms"
          onOpen={(id) => goToInspector(id)}
        />
        <EndpointTopCard
          title="Failing"
          icon={AlertTriangle}
          entries={metrics?.topFailing ?? []}
          metric="errorRate"
          onOpen={(id) => goToInspector(id)}
        />
        <EndpointTopCard
          title="Most used"
          icon={Zap}
          entries={metrics?.topUsed ?? []}
          metric="count"
          onOpen={(id) => goToInspector(id)}
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div className="md:col-span-2">
          <RecentActivityCard
            entries={recent}
            onOpen={(endpointID) => goToInspector(endpointID)}
          />
        </div>
        <SnapshotCard
          snapshot={latestSnapshot}
          onOpen={() => setCurrentPage('changelog')}
        />
      </div>
    </div>
  )
}

interface StatProps {
  card: StatCard
}

function Stat({ card }: StatProps) {
  const Icon = KIND_ICON[card.kind] ?? Navigation
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3" title={card.hint ?? ''}>
      <div className="flex items-center justify-between">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          {card.label}
        </p>
        <Icon className="w-3.5 h-3.5 text-primary/70" />
      </div>
      <p className="text-2xl font-semibold mt-1.5 tabular-nums">{card.value}</p>
    </div>
  )
}

function StatSkeleton() {
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3 space-y-2">
      <div className="flex items-center justify-between">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-3.5 w-3.5 rounded" />
      </div>
      <Skeleton className="h-7 w-12 mt-1.5" />
    </div>
  )
}

interface CardProps {
  title: string
  icon: LucideIcon
  action?: { label: string; onClick: () => void }
  children: React.ReactNode
}

function Card({ title, icon: Icon, action, children }: CardProps) {
  return (
    <section className="rounded-lg border border-border/60 bg-card/40 p-3.5 space-y-2.5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <Icon className="w-3.5 h-3.5 text-muted-foreground" />
          <h2 className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            {title}
          </h2>
        </div>
        {action && (
          <button
            type="button"
            onClick={action.onClick}
            className="inline-flex items-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground transition-colors"
          >
            {action.label}
            <ArrowUpRight className="w-3 h-3" />
          </button>
        )}
      </div>
      {children}
    </section>
  )
}

function AuthCard({ auth }: { auth: ReturnType<typeof useAuthStore.getState>['byProject'][string] | null }) {
  return (
    <Card title="Authentication" icon={KeyRound}>
      {auth?.user ? (
        <div className="space-y-1">
          <p className="text-[13px] font-medium truncate">
            {auth.user.name || auth.user.username || auth.user.email || 'User'}
          </p>
          {auth.user.email && (
            <p className="text-[11px] text-muted-foreground truncate">{auth.user.email}</p>
          )}
          {auth.user.role && (
            <p className="text-[10.5px] text-muted-foreground/80">{auth.user.role}</p>
          )}
          {auth.hasToken && (
            <div className="flex items-center gap-1.5 pt-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              <span className="text-[10.5px] text-muted-foreground font-mono truncate">
                {auth.tokenPreview}
              </span>
            </div>
          )}
        </div>
      ) : auth?.hasToken ? (
        <div className="space-y-1">
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span className="text-[12px] font-medium">Token active</span>
          </div>
          <p className="text-[10.5px] text-muted-foreground font-mono truncate">
            {auth.tokenPreview}
          </p>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">No active session.</p>
      )}
    </Card>
  )
}

interface EnvCardProps {
  env: { name: string; vars?: Record<string, string> } | null
  count: number
  onOpen: () => void
}

function EnvCard({ env, count }: EnvCardProps) {
  return (
    <Card title="Environment" icon={Layers}>
      {env ? (
        <div className="space-y-1">
          <p className="text-[13px] font-medium truncate">{env.name}</p>
          <p className="text-[10.5px] text-muted-foreground">
            {Object.keys(env.vars ?? {}).length} variable
            {Object.keys(env.vars ?? {}).length === 1 ? '' : 's'}
          </p>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">
          {count > 0 ? 'No active environment selected.' : 'No environments yet.'}
        </p>
      )}
    </Card>
  )
}

interface PinnedCardProps {
  endpoints: Array<{ id: string; method: string; path: string }>
  onOpen: (id: string) => void
}

function PinnedCard({ endpoints, onOpen }: PinnedCardProps) {
  const { getMethodColor } = useHttpMethod()
  return (
    <Card title="Pinned" icon={Star}>
      {endpoints.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">
          Pin endpoints from the inspector for quick access.
        </p>
      ) : (
        <ul className="space-y-px">
          {endpoints.map((ep) => (
            <li key={ep.id}>
              <button
                type="button"
                onClick={() => onOpen(ep.id)}
                className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 transition-colors"
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(ep.method),
                  )}
                >
                  {ep.method}
                </span>
                <span className="text-[11.5px] font-mono truncate flex-1 text-left text-foreground/85">
                  {ep.path}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}

interface RecentActivityCardProps {
  entries: Array<{
    id: string
    method: string
    url: string
    responseStatus: number
    durationMs: number
    error?: string
    createdAt: string | Date
    endpointID?: string
  }>
  onOpen: (endpointID?: string) => void
}

function RecentActivityCard({ entries, onOpen }: RecentActivityCardProps) {
  const { getMethodColor } = useHttpMethod()
  return (
    <Card title="Recent activity" icon={History}>
      {entries.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">
          No requests run yet.
        </p>
      ) : (
        <ul className="space-y-px">
          {entries.map((entry) => {
            const tone = entry.error
              ? 'text-destructive'
              : entry.responseStatus >= 500
                ? 'text-destructive'
                : entry.responseStatus >= 400
                  ? 'text-amber-500'
                  : entry.responseStatus >= 200
                    ? 'text-emerald-500'
                    : 'text-muted-foreground'
            return (
              <li key={entry.id}>
                <button
                  type="button"
                  onClick={() => onOpen(entry.endpointID)}
                  className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 transition-colors"
                >
                  <span
                    className={cn(
                      'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                      getMethodColor(entry.method),
                    )}
                  >
                    {entry.method}
                  </span>
                  <span className={cn('text-[11px] font-mono tabular-nums w-9 text-right shrink-0', tone)}>
                    {entry.error ? 'ERR' : entry.responseStatus}
                  </span>
                  <span className="text-[11.5px] font-mono truncate flex-1 text-left text-foreground/85">
                    {shortUrl(entry.url)}
                  </span>
                  <span className="text-[10px] text-muted-foreground tabular-nums shrink-0">
                    {entry.durationMs}ms
                  </span>
                  <span className="text-[10px] text-muted-foreground/70 shrink-0 w-12 text-right">
                    {timeAgo(new Date(entry.createdAt))}
                  </span>
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </Card>
  )
}

interface SnapshotCardProps {
  snapshot: {
    id: string
    scannedAt: string | Date
    endpointCount: number
    added: number
    removed: number
    changed: number
  } | undefined
  onOpen: () => void
}

function SnapshotCard({ snapshot, onOpen }: SnapshotCardProps) {
  return (
    <Card
      title="Latest snapshot"
      icon={GitCompare}
      action={snapshot ? { label: 'View all', onClick: onOpen } : undefined}
    >
      {snapshot ? (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[12px] font-medium">
              {formatDate(new Date(snapshot.scannedAt))}
            </span>
            <span className="text-[10.5px] font-mono text-muted-foreground">
              {snapshot.endpointCount} endpoints
            </span>
          </div>
          <div className="flex items-center gap-3 text-[11px] font-mono">
            <SnapshotDelta icon={Plus} count={snapshot.added} tone="emerald" />
            <SnapshotDelta icon={Minus} count={snapshot.removed} tone="rose" />
            <SnapshotDelta icon={Pencil} count={snapshot.changed} tone="amber" />
          </div>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">
          No snapshots yet. Run Sync to capture one.
        </p>
      )}
    </Card>
  )
}

function SnapshotDelta({
  icon: Icon,
  count,
  tone,
}: {
  icon: LucideIcon
  count: number
  tone: 'emerald' | 'rose' | 'amber'
}) {
  const toneClass = {
    emerald: 'text-emerald-500',
    rose: 'text-rose-500',
    amber: 'text-amber-500',
  }[tone]
  return (
    <span className={cn('inline-flex items-center gap-1', count === 0 ? 'text-muted-foreground/60' : toneClass)}>
      <Icon className="w-3 h-3" />
      {count}
    </span>
  )
}

function shortUrl(url: string): string {
  try {
    const u = new URL(url)
    return u.pathname + u.search
  } catch {
    return url
  }
}

function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}

function formatDate(date: Date): string {
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return `Today · ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  }
  return date.toLocaleString([], {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const STATUS_COLORS: Record<string, string> = {
  '2xx': '#10b981',
  '3xx': '#3b82f6',
  '4xx': '#f59e0b',
  '5xx': '#ef4444',
  err: '#ef4444',
}

function StatusCard({ metrics }: { metrics: DashboardMetrics | null }) {
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
              <span className={cn('font-mono tabular-nums', errorRate > 10 ? 'text-rose-500' : errorRate > 0 ? 'text-amber-500' : 'text-emerald-500')}>
                {errorRate.toFixed(1)}%
              </span>
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}

function LatencyCard({ metrics }: { metrics: DashboardMetrics | null }) {
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

const LATENCY_LEGEND: Record<string, string> = {
  p50: 'Median — half of requests faster',
  p95: '95% of requests faster',
  p99: '99% of requests faster (worst-case tail)',
}

function LatencyStat({ label, value, tone }: { label: string; value: number; tone: 'emerald' | 'amber' | 'rose' }) {
  const toneClass = {
    emerald: 'text-emerald-500',
    amber: 'text-amber-500',
    rose: 'text-rose-500',
  }[tone]
  return (
    <div
      className="rounded-md border border-border/40 bg-muted/20 px-2 py-1.5"
      title={LATENCY_LEGEND[label]}
    >
      <p className="text-[9.5px] uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className={cn('text-[14px] font-semibold tabular-nums', toneClass)}>
        {value}
        <span className="text-[10px] text-muted-foreground/80 ml-0.5">ms</span>
      </p>
    </div>
  )
}

function VolumeCard({ metrics }: { metrics: DashboardMetrics | null }) {
  const volume = metrics?.volume ?? []
  const total = volume.reduce((s, v) => s + v.count, 0)
  return (
    <Card title="Volume · 7d" icon={TrendingUp}>
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

interface EndpointTopCardProps {
  title: string
  icon: LucideIcon
  entries: EndpointMetricDTO[]
  metric: 'ms' | 'count' | 'errorRate'
  onOpen: (id: string) => void
}

function EndpointTopCard({ title, icon, entries, metric, onOpen }: EndpointTopCardProps) {
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
