import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import {Search, RefreshCw, ChevronDown, PanelRightClose, Folder, Lock, User, Key, Shield} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ThemeSwitcher } from './ThemeSwitcher'

export function Topbar() {
  const setCommandPaletteOpen = useUIStore((state) => state.setCommandPaletteOpen)
  const toggleSidebar = useUIStore((state) => state.toggleSidebar)
  const setAuthDrawerOpen = useUIStore((state) => state.setAuthDrawerOpen)
  const activeAuthMethod = useUIStore((state) => state.activeAuthMethod)
  const projects = useProjectStore((state) => state.projects)
  const activeProjectId = useProjectStore((state) => state.activeProjectId)
  const syncProject = useProjectStore((state) => state.syncProject)
  const isSyncing = useProjectStore((state) => state.isSyncing)

  const authMethodConfig = {
    'current-user': { label: 'Current User', icon: User },
    'impersonate': { label: 'Impersonate', icon: Key },
    'bearer-token': { label: 'Bearer Token', icon: Lock },
    'basic-auth': { label: 'Basic Auth', icon: Shield },
  }

  const authConfig = authMethodConfig[activeAuthMethod as keyof typeof authMethodConfig] || { label: 'Auth', icon: Lock }
  const AuthIcon = authConfig.icon

  const activeProject = projects.find((p) => p.id === activeProjectId)

  const handleSync = async () => {
    if (activeProjectId) {
      await syncProject(activeProjectId)
    }
  }

  return (
    <div className="h-14 border-b border-border/50 bg-background flex items-center justify-between px-4 gap-4">
      {/* Left: Menu & Project Switcher */}
      <div className="flex items-center gap-4 flex-1">
        <Button variant="ghost" size="icon" onClick={toggleSidebar} title="Toggle sidebar" className="h-8 w-8">
          <PanelRightClose className="w-4 h-4" />
        </Button>

        {/* Project Switcher */}
        {activeProject && (
          <button className="flex items-center gap-2 hover:bg-muted/50 px-3 py-1.5 rounded-md transition-colors">
            <Folder className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            <span className="text-sm font-medium">{activeProject.name}</span>
            <ChevronDown className="w-3 h-3 text-muted-foreground flex-shrink-0" />
          </button>
        )}
        {!activeProject && (
          <div className="flex items-center gap-2 px-3 py-1.5">
            <Folder className="w-4 h-4 text-muted-foreground flex-shrink-0" />
            <span className="text-sm text-muted-foreground">No project selected</span>
          </div>
        )}
      </div>

      {/* Center: Search */}
      <div className="flex-1 max-w-md mx-auto">
        <button
          onClick={() => setCommandPaletteOpen(true)}
          className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-muted-foreground bg-muted/30 hover:bg-muted/50 border border-border/50 rounded-md transition-colors"
        >
          <Search className="w-4 h-4" />
          <span>Search...</span>
          <span className="ml-auto text-xs">⌘K</span>
        </button>
      </div>

      {/* Right: Actions */}
      <div className="flex items-center gap-2 flex-1 justify-end">

        <Button
          variant="ghost"
          size="sm"
          onClick={handleSync}
          disabled={!activeProjectId || isSyncing === activeProjectId}
          className="h-8"
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${isSyncing === activeProjectId ? 'animate-spin' : ''}`} />
          <span className="text-xs">{isSyncing === activeProjectId ? 'Syncing...' : 'Sync'}</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            localStorage.removeItem('spectra-onboarded')
            window.location.reload()
          }}
          className="h-8"
        >
          <span className="text-xs">Reset Onboarding</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setAuthDrawerOpen(true)}
          title="Change authentication method"
          className="h-8 gap-1.5"
        >
          <AuthIcon className="w-4 h-4" />
          <span className="text-xs">{authConfig.label}</span>
        </Button>
        <ThemeSwitcher />
      </div>
    </div>
  )
}
