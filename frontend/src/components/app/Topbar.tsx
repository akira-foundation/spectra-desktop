import type { CSSProperties } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useAuthStore } from '@/store/authStore'
import { useAccountsStore } from '@/store/accountsStore'
import { useUpdatesStore, isUpdateActionable } from '@/store/updatesStore'
import type { ProjectAccount } from '@/services/accountsService'

const EMPTY_ACCOUNTS: ProjectAccount[] = []
import { Search, RefreshCw, Lock, User, Key, Shield, FileDown, Download } from 'lucide-react'
import toast from 'react-hot-toast'
import { SaveOpenAPIToFile } from '../../../wailsjs/go/app/App'
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
  const setCurrentPage = useUIStore((state) => state.setCurrentPage)
  const updatePhase = useUpdatesStore((state) => state.phase)
  const updateVersion = useUpdatesStore((state) => state.info?.version ?? null)
  const updateActionable = isUpdateActionable(updatePhase)
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
  const accountList = useAccountsStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_ACCOUNTS : EMPTY_ACCOUNTS,
  )
  const accountActiveId = useAccountsStore((s) =>
    activeProjectId ? s.activeByProject[activeProjectId] ?? null : null,
  )
  const activeAccount =
    accountList.find((a) => a.id === accountActiveId) ??
    accountList.find((a) => a.isDefault) ??
    accountList[0] ??
    null
  const accountUser = activeAccount?.user as
    | { name?: string; username?: string; email?: string; id?: string | number }
    | undefined
  const authLabel = (() => {
    if (accountUser) {
      return (
        accountUser.name ||
        accountUser.username ||
        accountUser.email ||
        (accountUser.id !== undefined ? String(accountUser.id) : '') ||
        activeAccount?.label ||
        authConfig.label
      )
    }
    if (activeAccount) return activeAccount.label
    if (projectAuth?.user) {
      const u = projectAuth.user
      return u.name || u.username || u.email || u.id || authConfig.label
    }
    return authConfig.label
  })()
  const hasToken = !!activeAccount?.hasToken || !!projectAuth?.hasToken

  const activeProject = projects.find((p) => p.id === activeProjectId)
  const openAddProject = () => setAddProjectOpen(true)

  const handleSync = async () => {
    if (activeProjectId) {
      await rescan(activeProjectId)
    }
  }

  const handleExport = async () => {
    if (!activeProjectId) return
    try {
      const path = await SaveOpenAPIToFile(activeProjectId)
      if (path) toast.success(`OpenAPI saved: ${path}`)
    } catch (err) {
      toast.error('Export failed')
      console.error(err)
    }
  }

  return (
    <div
      className="h-10 grid grid-cols-[1fr_auto_1fr] items-center gap-3 pr-3 select-none bg-[#e5e5e5] dark:bg-transparent text-foreground/90 dark:text-white/90"
      style={{ ...drag, paddingLeft: 80 }}
    >
      <div className="flex items-center gap-2 min-w-0 justify-self-start" style={noDrag}>
        <ProjectSwitcher
          projects={projects}
          activeProject={activeProject}
          onSelect={setActiveProject}
          onAddProject={openAddProject}
        />
      </div>

      <div className="w-96 max-w-[calc(100vw-32rem)] justify-self-center" style={noDrag}>
        <button
          onClick={() => setCommandPaletteOpen(true)}
          className="w-full h-7 flex items-center gap-2 px-2.5 text-[12px] text-muted-foreground dark:text-white/65 bg-foreground/5 dark:bg-white/5 hover:bg-foreground/10 dark:hover:bg-white/10 border border-foreground/10 dark:border-white/10 rounded-md transition-colors"
        >
          <Search className="w-3.5 h-3.5" />
          <span>Search...</span>
          <span className="ml-auto text-[10px] tracking-wide opacity-70">⌘K</span>
        </button>
      </div>

      <div className="flex items-center gap-0.5 justify-self-end" style={noDrag}>
        {updateActionable && (
          <button
            type="button"
            onClick={() => setCurrentPage('settings')}
            className="h-7 px-2 mr-1 inline-flex items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 text-primary text-[11px] hover:bg-primary/15 transition-colors"
            title={updatePhase === 'ready' ? 'Update ready to install' : `Update available: ${updateVersion ?? ''}`}
          >
            <Download className="w-3.5 h-3.5" />
            <span>{updatePhase === 'ready' ? 'Install' : 'Update'}</span>
            {updateVersion && <span className="font-mono text-[10px] opacity-80">{updateVersion}</span>}
          </button>
        )}
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
          onClick={handleExport}
          disabled={!activeProjectId}
          className="h-7 px-2 text-[11px] text-foreground/85 dark:text-white/85 hover:bg-foreground/10 dark:hover:bg-white/10 hover:text-foreground dark:hover:text-white"
          title="Export OpenAPI spec"
        >
          <FileDown className="w-3.5 h-3.5" />
          <span>OpenAPI</span>
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setAuthDrawerOpen(true)}
          className="h-7 px-2 text-[11px] gap-1.5 text-foreground/85 dark:text-white/85 hover:bg-foreground/10 dark:hover:bg-white/10 hover:text-foreground dark:hover:text-white"
        >
          <AuthIcon className="w-3.5 h-3.5" />
          <span>{authLabel}</span>
          {hasToken && (
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
          )}
        </Button>
        <ThemeSwitcher />
      </div>
    </div>
  )
}
