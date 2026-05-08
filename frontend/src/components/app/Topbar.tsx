import type { CSSProperties } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useAuthStore } from '@/store/authStore'
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
  const rescan = useEndpointsStore((state) => state.scan)
  const scanStatus = useEndpointsStore((state) =>
    activeProjectId ? state.status[activeProjectId] ?? 'idle' : 'idle',
  )
  const isScanning = scanStatus === 'scanning' || scanStatus === 'loading'

  const authMethodConfig = {
    'current-user': { label: 'Current User', icon: User },
    'impersonate': { label: 'Impersonate', icon: Key },
    'bearer-token': { label: 'Bearer Token', icon: Lock },
    'basic-auth': { label: 'Basic Auth', icon: Shield },
  }

  const authConfig =
    authMethodConfig[activeAuthMethod as keyof typeof authMethodConfig] ?? { label: 'Auth', icon: Lock }
  const AuthIcon = authConfig.icon

  const projectAuth = useAuthStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const authLabel = (() => {
    if (projectAuth?.user) {
      const u = projectAuth.user
      return u.name || u.username || u.email || u.id || authConfig.label
    }
    return authConfig.label
  })()

  const activeProject = projects.find((p) => p.id === activeProjectId)
  const openAddProject = () => setAddProjectOpen(true)

  const handleSync = async () => {
    if (activeProjectId) {
      await rescan(activeProjectId)
    }
  }

  return (
    <div
      className="h-10 flex items-center justify-between gap-3 pr-3 select-none bg-[#e5e5e5] dark:bg-transparent text-foreground/90 dark:text-white/90"
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
          className="w-full h-7 flex items-center gap-2 px-2.5 text-[12px] text-muted-foreground dark:text-white/65 bg-foreground/5 dark:bg-white/5 hover:bg-foreground/10 dark:hover:bg-white/10 border border-foreground/10 dark:border-white/10 rounded-md transition-colors"
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
          disabled={!activeProjectId || isScanning}
          className="h-7 px-2 text-[11px] text-foreground/85 dark:text-white/85 hover:bg-foreground/10 dark:hover:bg-white/10 hover:text-foreground dark:hover:text-white"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isScanning ? 'animate-spin' : ''}`} />
          <span>{isScanning ? 'Scanning' : 'Sync'}</span>
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setAuthDrawerOpen(true)}
          className="h-7 px-2 text-[11px] gap-1.5 text-foreground/85 dark:text-white/85 hover:bg-foreground/10 dark:hover:bg-white/10 hover:text-foreground dark:hover:text-white"
        >
          <AuthIcon className="w-3.5 h-3.5" />
          <span>{authLabel}</span>
          {projectAuth?.hasToken && (
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
          )}
        </Button>
        <ThemeSwitcher />
      </div>
    </div>
  )
}
