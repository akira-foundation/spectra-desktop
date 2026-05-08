import { useEffect, useMemo } from 'react'
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
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Welcome } from '@/components/pages/Welcome'
import { Skeleton } from '@/components/ui/skeleton'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import type { StatCard } from '@/services/scannerService'
import { cn } from '@/lib/utils'

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
    activeProjectId ? s.byProject[activeProjectId] ?? [] : [],
  )

  const pinnedKeys = useUIStore((s) =>
    activeProjectId ? s.pinnedEndpointsByProject[activeProjectId] ?? null : null,
  )
  const setSelectedEndpoint = useUIStore((s) => s.setSelectedEndpoint)

  useEffect(() => {
    if (!activeProjectId) return
    void loadReport(activeProjectId)
    void loadAuth(activeProjectId)
    void loadEnvs(activeProjectId)
    void loadHistory(activeProjectId)
    void loadSnapshots(activeProjectId)
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
          {activeProject.name[0]}
        </div>
        <div className="flex-1">
          <h1 className="text-xl font-semibold tracking-tight">{activeProject.name}</h1>
          <p className="text-foreground/60 text-[12px]">
            {activeProject.framework}
            {activeProject.frameworkVersion ? ` · ${activeProject.frameworkVersion}` : ''}
            {activeProject.baseUrl ? ` · ${activeProject.baseUrl}` : ''}
          </p>
        </div>
      </header>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2.5">
        {cards.length > 0
          ? cards.map((card) => <Stat key={card.key} card={card} />)
          : Array.from({ length: 5 }).map((_, i) => <StatSkeleton key={i} />)}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div className="space-y-3 lg:col-span-1">
          <AuthCard auth={auth} />
          <EnvCard env={activeEnv} count={envs?.length ?? 0} onOpen={() => goToInspector()} />
          <PinnedCard
            endpoints={pinnedEndpoints}
            onOpen={(id) => goToInspector(id)}
          />
        </div>
        <div className="space-y-3 lg:col-span-2">
          <RecentActivityCard
            entries={recent}
            onOpen={(endpointID) => goToInspector(endpointID)}
          />
          <SnapshotCard
            snapshot={latestSnapshot}
            onOpen={() => setCurrentPage('changelog')}
          />
        </div>
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
