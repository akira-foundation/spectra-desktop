import type { CSSProperties } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { Search, RefreshCw, Lock, User, Key, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ProjectSwitcher } from '@/components/projects/ProjectSwitcher'
import { ThemeSwitcher } from './ThemeSwitcher'

const drag = { '--wails-draggable': 'drag' } as CSSProperties
const noDrag = { '--wails-draggable': 'no-drag' } as CSSProperties

export function Topbar() {
  const setCommandPaletteOpen = useUIStore((state) => state.setCommandPaletteOpen)
  const setAuthDrawerOpen = useUIStore((state) => state.setAuthDrawerOpen)
  const setAddProjectOpen = useUIStore((state) => state.setAddProjectOpen)
  const activeAuthMethod = useUIStore((state) => state.activeAuthMethod)
  const projects = useProjectStore((state) => state.projects)
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const setActiveProject = useProjectStore((state) => state.setActiveProject)
  const syncProject = useProjectStore((state) => state.syncProject)
  const isSyncing = useProjectStore((state) => state.isSyncing)

  const authMethodConfig = {
    'current-user': { label: 'Current User', icon: User },
    'impersonate': { label: 'Impersonate', icon: Key },
    'bearer-token': { label: 'Bearer Token', icon: Lock },
    'basic-auth': { label: 'Basic Auth', icon: Shield },
  }

  const authConfig =
    authMethodConfig[activeAuthMethod as keyof typeof authMethodConfig] ?? { label: 'Auth', icon: Lock }
  const AuthIcon = authConfig.icon

  const activeProject = projects.find((p) => p.id === activeProjectId)
  const openAddProject = () => setAddProjectOpen(true)

  const handleSync = async () => {
    if (activeProjectId) {
      await syncProject(activeProjectId)
    }
  }

  return (
    <div
      className="h-10 border-b border-border/60 bg-card/40 backdrop-blur-md flex items-center justify-between gap-3 pr-3 select-none"
      style={{ ...drag, paddingLeft: 80 }}
    >
      <div className="flex items-center gap-2 min-w-0" style={noDrag}>
        <ProjectSwitcher
          projects={projects}
          activeProject={activeProject}
          onSelect={setActiveProject}
          onAddProject={openAddProject}
        />
      </div>

      <div className="flex-1 max-w-md mx-auto" style={noDrag}>
        <button
          onClick={() => setCommandPaletteOpen(true)}
          className="w-full h-7 flex items-center gap-2 px-2.5 text-[12px] text-muted-foreground bg-muted/40 hover:bg-muted/60 border border-border/50 rounded-md transition-colors"
        >
          <Search className="w-3.5 h-3.5" />
          <span>Search...</span>
          <span className="ml-auto text-[10px] tracking-wide opacity-70">⌘K</span>
        </button>
      </div>

      <div className="flex items-center gap-0.5" style={noDrag}>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleSync}
          disabled={!activeProjectId || isSyncing === activeProjectId}
          className="h-7 px-2 text-[11px]"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isSyncing === activeProjectId ? 'animate-spin' : ''}`} />
          <span>{isSyncing === activeProjectId ? 'Syncing' : 'Sync'}</span>
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setAuthDrawerOpen(true)}
          className="h-7 px-2 text-[11px] gap-1.5"
        >
          <AuthIcon className="w-3.5 h-3.5" />
          <span>{authConfig.label}</span>
        </Button>
        <ThemeSwitcher />
      </div>
    </div>
  )
}
