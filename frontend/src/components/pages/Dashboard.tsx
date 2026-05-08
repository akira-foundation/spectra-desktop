import { useEffect } from 'react'
import { useProjectStore } from '@/store/projectStore'
import { useStatsStore } from '@/store/statsStore'
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
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Welcome } from '@/components/pages/Welcome'
import type { StatCard } from '@/services/scannerService'

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
  const loadReport = useStatsStore((s) => s.loadReport)
  const report = useStatsStore((s) =>
    activeProjectId ? s.reportByProject[activeProjectId] ?? null : null,
  )

  useEffect(() => {
    if (activeProjectId) void loadReport(activeProjectId)
  }, [activeProjectId, loadReport])

  if (!activeProject) {
    return <Welcome />
  }

  const cards = report?.cards ?? []

  return (
    <div className="space-y-6 p-6">
      <header className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-md bg-primary/10 flex items-center justify-center text-primary font-semibold text-base">
          {activeProject.name[0]}
        </div>
        <div>
          <h1 className="text-xl font-semibold tracking-tight">{activeProject.name}</h1>
          <p className="text-foreground/60 text-[12px]">
            {activeProject.framework} · {activeProject.frameworkVersion}
          </p>
        </div>
      </header>

      {cards.length > 0 ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2.5">
          {cards.map((card) => (
            <Stat key={card.key} card={card} />
          ))}
        </div>
      ) : (
        <div className="text-[12px] text-muted-foreground italic">Loading stats…</div>
      )}
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

