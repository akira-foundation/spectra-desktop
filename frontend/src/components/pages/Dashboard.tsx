import { useProjectStore } from '@/store/projectStore'
import { Navigation, Code, Database, Cpu, AlertCircle, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {Welcome} from "@/components/pages/Welcome";

export function Dashboard() {
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const projects = useProjectStore((state) => state.projects)

  const activeProject = projects.find((p) => p.id === activeProjectId)

  if (!activeProject) {
    return (
     <Welcome/>
    )
  }

  const StatCard = ({ icon: Icon, label, value }: { icon: any; label: string; value: number }) => (
    <div className="glass-card p-6 space-y-2">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-foreground/60">{label}</p>
        <Icon className="w-5 h-5 text-primary/60" />
      </div>
      <p className="text-3xl font-bold">{value}</p>
    </div>
  )

  return (
    <div className="space-y-8 p-8">
      {/* Header */}
      <div className="space-y-2">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary font-bold text-lg">
            {activeProject.name[0]}
          </div>
          <div>
            <h1 className="text-3xl font-bold">{activeProject.name}</h1>
            <p className="text-foreground/60 text-sm">{activeProject.framework} • {activeProject.frameworkVersion}</p>
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
        <StatCard icon={Navigation} label="Routes" value={activeProject.stats.routes} />
        <StatCard icon={Database} label="Models" value={activeProject.stats.models} />
        <StatCard icon={Cpu} label="Middleware" value={activeProject.stats.middleware} />
        <StatCard icon={Code} label="Controllers" value={activeProject.stats.controllers} />
        <StatCard icon={AlertCircle} label="Errors" value={activeProject.stats.errors} />
      </div>

      {/* SDK Status */}
      <div className="glass-card p-6 space-y-4">
        <h2 className="text-lg font-semibold">SDK Status</h2>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm">Status:</span>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-sm font-medium">{activeProject.status}</span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm">SDK Version:</span>
            <span className="text-sm font-medium">{activeProject.sdkVersion}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm">Last Sync:</span>
            <span className="text-sm font-medium">
              {activeProject.lastSyncTime
                ? new Date(activeProject.lastSyncTime).toLocaleDateString()
                : 'Never'}
            </span>
          </div>
        </div>
      </div>

      {/* Quick Navigation */}
      <div className="space-y-3">
        <h2 className="text-lg font-semibold">Quick Navigation</h2>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
          {[
            { label: 'Endpoints', icon: Navigation },
            { label: 'Models', icon: Database },
            { label: 'Controllers', icon: Code },
            { label: 'Middleware', icon: Cpu },
            { label: 'Changelog', icon: Navigation },
            { label: 'Settings', icon: Cpu },
          ].map((item) => {
            const Icon = item.icon
            return (
              <Button
                key={item.label}
                variant="ghost"
                className="glass-card h-auto flex flex-col items-center gap-2 p-4"
              >
                <Icon className="w-6 h-6 text-primary" />
                <span className="text-sm font-medium">{item.label}</span>
              </Button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
