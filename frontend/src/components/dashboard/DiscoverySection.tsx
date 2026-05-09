import { useEffect, useState } from 'react'
import { Compass, Clock, GitBranch, ShieldCheck, FileCheck, Hash } from 'lucide-react'
import { discoveryService, type Discovery } from '@/services/discoveryService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { cn } from '@/lib/utils'

interface Props {
  projectId: string | null
  refreshKey?: number
}

export function DiscoverySection({ projectId, refreshKey }: Props) {
  const [data, setData] = useState<Discovery | null>(null)
  useEffect(() => {
    if (!projectId) return
    void discoveryService.get(projectId, 30).then(setData)
  }, [projectId, refreshKey])
  if (!projectId) return null
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <CoverageCard data={data} />
        <TestCoverageCard data={data} />
        <CompositionCard data={data} />
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <UnusedCard data={data} />
        <StaleCard data={data} />
      </div>
    </div>
  )
}

function TestCoverageCard({ data }: { data: Discovery | null }) {
  const pct = data ? Math.round(data.testCoverage * 100) : 0
  const tested = data?.testedEndpoints ?? 0
  const total = data?.totalEndpoints ?? 0
  return (
    <div className="rounded-md border border-border/40 bg-card/30 p-4 flex flex-col">
      <div className="flex items-center gap-1.5 mb-3">
        <FileCheck className="w-3 h-3 text-muted-foreground" />
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          Test coverage
        </h3>
      </div>
      <div className="flex items-baseline gap-2">
        <div className="text-[28px] font-semibold tabular-nums leading-none">{pct}%</div>
        <div className="text-[10.5px] text-muted-foreground">
          {tested} of {total} have tests/captures
        </div>
      </div>
      <div className="mt-3 flex h-2 rounded-full overflow-hidden bg-muted/40">
        <div className="bg-emerald-500/80" style={{ width: `${pct}%` }} />
      </div>
      {total - tested > 0 && (
        <p className="mt-2 text-[10.5px] text-muted-foreground">
          {total - tested} endpoint{total - tested === 1 ? '' : 's'} without assertions.
        </p>
      )}
    </div>
  )
}

function CompositionCard({ data }: { data: Discovery | null }) {
  const total = data?.totalEndpoints ?? 0
  const reads = data?.readEndpoints ?? 0
  const writes = data?.writeEndpoints ?? 0
  const auth = data?.authRequired ?? 0
  const pub = data?.authPublic ?? 0
  const readPct = total > 0 ? (reads / total) * 100 : 0
  const writePct = total > 0 ? (writes / total) * 100 : 0
  const authPct = total > 0 ? (auth / total) * 100 : 0
  return (
    <div className="rounded-md border border-border/40 bg-card/30 p-4 flex flex-col gap-3">
      <div className="flex items-center gap-1.5">
        <Hash className="w-3 h-3 text-muted-foreground" />
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          Composition
        </h3>
      </div>
      <div>
        <div className="flex items-baseline justify-between mb-1">
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground/70">Reads / writes</span>
          <span className="text-[10px] font-mono tabular-nums text-muted-foreground">
            <span className="text-emerald-500/90">{reads}</span>
            <span className="text-muted-foreground/40 mx-1">/</span>
            <span className="text-amber-500/90">{writes}</span>
          </span>
        </div>
        <div className="flex h-1.5 rounded-full overflow-hidden bg-muted/40">
          <div className="bg-emerald-500/70" style={{ width: `${readPct}%` }} />
          <div className="bg-amber-500/70" style={{ width: `${writePct}%` }} />
        </div>
      </div>
      <div>
        <div className="flex items-baseline justify-between mb-1">
          <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground/70">Auth / public</span>
          <span className="text-[10px] font-mono tabular-nums">
            <ShieldCheck className="w-3 h-3 inline mr-0.5 text-primary/80 align-text-top" />
            <span className="text-primary/90">{auth}</span>
            <span className="text-muted-foreground/40 mx-1">/</span>
            <span className="text-muted-foreground/80">{pub}</span>
          </span>
        </div>
        <div className="flex h-1.5 rounded-full overflow-hidden bg-muted/40">
          <div className="bg-primary/70" style={{ width: `${authPct}%` }} />
          <div className="bg-muted-foreground/30" style={{ width: `${100 - authPct}%` }} />
        </div>
      </div>
    </div>
  )
}

function CoverageCard({ data }: { data: Discovery | null }) {
  const pct = data ? Math.round(data.coverage * 100) : 0
  const total = data?.totalEndpoints ?? 0
  const used = data?.usedEndpoints ?? 0
  return (
    <div className="rounded-md border border-border/40 bg-card/30 p-4 flex flex-col">
      <div className="flex items-center gap-1.5 mb-3">
        <Compass className="w-3 h-3 text-muted-foreground" />
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          Coverage
        </h3>
      </div>
      <div className="flex items-baseline gap-2">
        <div className="text-[28px] font-semibold tabular-nums leading-none">{pct}%</div>
        <div className="text-[10.5px] text-muted-foreground">
          {used} of {total} endpoints called
        </div>
      </div>
      <div className="mt-3 flex h-2 rounded-full overflow-hidden bg-muted/40">
        <div className="bg-primary/80" style={{ width: `${pct}%` }} />
      </div>
      {total - used > 0 && (
        <p className="mt-2 text-[10.5px] text-muted-foreground">
          {total - used} endpoint{total - used === 1 ? '' : 's'} never called.
        </p>
      )}
    </div>
  )
}

function UnusedCard({ data }: { data: Discovery | null }) {
  const items = data?.unused ?? []
  return (
    <DiscoveryListCard
      title="Unused endpoints"
      icon={GitBranch}
      items={items}
      emptyMessage="All endpoints have been called."
      hint="Possible dead routes — never called."
    />
  )
}

function StaleCard({ data }: { data: Discovery | null }) {
  const items = data?.stale ?? []
  return (
    <DiscoveryListCard
      title="Stale endpoints"
      icon={Clock}
      items={items}
      emptyMessage="No stale endpoints in last 30 days."
      hint="Not called in 30+ days."
      showAge
    />
  )
}

function DiscoveryListCard({
  title,
  icon: Icon,
  items,
  emptyMessage,
  hint,
  showAge,
}: {
  title: string
  icon: typeof Compass
  items: Discovery['unused' | 'stale']
  emptyMessage: string
  hint?: string
  showAge?: boolean
}) {
  const { getMethodColor } = useHttpMethod()
  const setSelectedEndpoint = useUIStore((s) => s.setSelectedEndpoint)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)
  const projectId = useProjectStore((s) => s.activeProjectId)
  const open = (id: string) => {
    if (!projectId) return
    setSelectedEndpoint(projectId, id)
    setCurrentPage('inspector')
  }

  return (
    <div className="rounded-md border border-border/40 bg-card/30 p-4 flex flex-col min-h-0">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-1.5">
          <Icon className="w-3 h-3 text-muted-foreground" />
          <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            {title}
          </h3>
          <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{items.length}</span>
        </div>
      </div>
      {hint && <p className="text-[10px] text-muted-foreground/70 mb-2">{hint}</p>}
      {items.length === 0 ? (
        <p className="text-[11px] italic text-muted-foreground/60 text-center py-6">{emptyMessage}</p>
      ) : (
        <ul className="m-0 p-0 list-none space-y-1 max-h-48 overflow-y-auto">
          {items.slice(0, 30).map((e) => (
            <li key={e.endpointID}>
              <button
                type="button"
                onClick={() => open(e.endpointID)}
                className="w-full flex items-center gap-2 px-1.5 py-1 rounded hover:bg-accent/40 text-left"
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
                    getMethodColor(e.method),
                  )}
                >
                  {e.method}
                </span>
                <code className="text-[11px] font-mono truncate flex-1 text-foreground/85">{e.path}</code>
                {showAge && e.daysAgo !== undefined && e.daysAgo > 0 && (
                  <span className="text-[10px] font-mono text-muted-foreground/70 tabular-nums shrink-0">
                    {e.daysAgo}d
                  </span>
                )}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
