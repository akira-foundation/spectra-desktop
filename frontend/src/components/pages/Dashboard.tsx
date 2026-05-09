import { useEffect, useMemo, useState } from 'react'
import { metricsService, type DashboardMetrics } from '@/services/metricsService'
import { InsightsSection } from '@/components/dashboard/InsightsSection'
import { DiscoverySection } from '@/components/dashboard/DiscoverySection'
import { DashboardTabs, type DashboardTab } from '@/components/dashboard/DashboardTabs'
import { ActivityTab } from '@/components/dashboard/ActivityTab'
import { StatusCard, LatencyCard, VolumeCard } from '@/components/dashboard/cards'
import { useProjectStore } from '@/store/projectStore'
import { useStatsStore } from '@/store/statsStore'
import { useAuthStore } from '@/store/authStore'
import { useEnvironmentStore } from '@/store/environmentStore'
import { useHistoryStore } from '@/store/historyStore'
import { useChangelogStore } from '@/store/changelogStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useUIStore } from '@/store/uiStore'
import { Welcome } from '@/components/pages/Welcome'
import type { ScannedEndpoint } from '@/services/scannerService'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []

export function Dashboard() {
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const projects = useProjectStore((state) => state.projects)
  const activeProject = projects.find((p) => p.id === activeProjectId)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)

  const loadReport = useStatsStore((s) => s.loadReport)
  const loadAuth = useAuthStore((s) => s.load)
  const loadEnvs = useEnvironmentStore((s) => s.load)

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
  const [volumeDays, setVolumeDays] = useState<7 | 14 | 30>(7)
  const [activeTab, setActiveTab] = useState<DashboardTab>('overview')

  useEffect(() => {
    if (!activeProjectId) return
    void loadReport(activeProjectId)
    void loadAuth(activeProjectId)
    void loadEnvs(activeProjectId)
    void loadHistory(activeProjectId)
    void loadSnapshots(activeProjectId)
  }, [activeProjectId, loadReport, loadAuth, loadEnvs, loadHistory, loadSnapshots])

  useEffect(() => {
    if (!activeProjectId) return
    void metricsService.get(activeProjectId, volumeDays).then((m) => setMetrics(m))
  }, [activeProjectId, volumeDays, history?.length])

  // pinnedEndpoints unused on overview/activity tabs but kept for downstream extensibility
  useMemo(() => {
    const set = new Set(pinnedKeys ?? [])
    return allEndpoints.filter((e) => set.has(`${e.method} ${e.path}`)).slice(0, 8)
  }, [allEndpoints, pinnedKeys])

  if (!activeProject) return <Welcome />

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

      <DashboardTabs active={activeTab} onChange={setActiveTab} />

      {activeTab === 'overview' && (
        <div className="space-y-3">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <StatusCard metrics={metrics} />
            <LatencyCard metrics={metrics} />
            <VolumeCard metrics={metrics} days={volumeDays} onChangeDays={setVolumeDays} />
          </div>
          <InsightsSection
            projectId={activeProjectId ?? null}
            days={volumeDays}
            refreshKey={history?.length ?? 0}
          />
        </div>
      )}

      {activeTab === 'discovery' && (
        <DiscoverySection projectId={activeProjectId ?? null} refreshKey={history?.length ?? 0} />
      )}

      {activeTab === 'activity' && (
        <ActivityTab
          history={history ?? []}
          recent={recent}
          latestSnapshot={latestSnapshot}
          onOpen={goToInspector}
          onOpenSnapshots={() => setCurrentPage('changelog')}
          onOpenCollections={() => setCurrentPage('collections')}
        />
      )}
    </div>
  )
}
