import { useProjectStore } from '@/store/projectStore'
import { Navigation, Code, Database, Cpu, AlertCircle } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Welcome } from '@/components/pages/Welcome'

export function Dashboard() {
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const projects = useProjectStore((state) => state.projects)
  const activeProject = projects.find((p) => p.id === activeProjectId)

  if (!activeProject) {
    return <Welcome />
  }

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

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2.5">
        <Stat icon={Navigation} label="Routes" value={activeProject.stats.routes} />
        <Stat icon={Database} label="Models" value={activeProject.stats.models} />
        <Stat icon={Cpu} label="Middleware" value={activeProject.stats.middleware} />
        <Stat icon={Code} label="Controllers" value={activeProject.stats.controllers} />
        <Stat icon={AlertCircle} label="Errors" value={activeProject.stats.errors} />
      </div>

      <section className="rounded-lg border border-border/60 bg-card/40 p-4 space-y-3">
        <h2 className="text-[12px] font-semibold uppercase tracking-wider text-muted-foreground">
          SDK Status
        </h2>
        <div className="space-y-2 text-[12.5px]">
          <Row label="Status">
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              <span className="font-medium capitalize">{activeProject.status}</span>
            </span>
          </Row>
          <Row label="SDK version">
            <span className="font-mono text-[11.5px]">{activeProject.sdkVersion}</span>
          </Row>
          <Row label="Last sync">
            <span className="font-medium">
              {activeProject.lastSyncTime
                ? new Date(activeProject.lastSyncTime).toLocaleDateString()
                : 'Never'}
            </span>
          </Row>
        </div>
      </section>

      <section className="space-y-2">
        <h2 className="text-[12px] font-semibold uppercase tracking-wider text-muted-foreground">
          Quick Navigation
        </h2>
        <div className="grid grid-cols-3 sm:grid-cols-6 gap-2">
          {[
            { label: 'Endpoints', icon: Navigation },
            { label: 'Models', icon: Database },
            { label: 'Controllers', icon: Code },
            { label: 'Middleware', icon: Cpu },
            { label: 'Changelog', icon: Navigation },
            { label: 'Settings', icon: Cpu },
          ].map(({ label, icon: Icon }) => (
            <Button
              key={label}
              variant="outline"
              className="h-auto flex-col gap-1.5 py-3 border-border/60 bg-card/30 hover:bg-card/60"
            >
              <Icon className="w-4 h-4 text-primary" />
              <span className="text-[11px] font-medium">{label}</span>
            </Button>
          ))}
        </div>
      </section>
    </div>
  )
}

interface StatProps {
  icon: LucideIcon
  label: string
  value: number
}

function Stat({ icon: Icon, label, value }: StatProps) {
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3">
      <div className="flex items-center justify-between">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          {label}
        </p>
        <Icon className="w-3.5 h-3.5 text-primary/70" />
      </div>
      <p className="text-2xl font-semibold mt-1.5 tabular-nums">{value}</p>
    </div>
  )
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      {children}
    </div>
  )
}
