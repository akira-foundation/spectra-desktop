import { ReactNode, useEffect } from 'react'
import { Topbar } from './Topbar'
import { Sidebar } from './Sidebar'
import { TabBar } from './TabBar'
import { StatusBar } from './StatusBar'
import { DrawerMenu } from './DrawerMenu'
import { CommandPalette } from './CommandPalette'
import { AddProjectDialog } from '@/components/projects/AddProjectDialog'
import { AuthenticationDrawer } from '@/components/api-inspector/AuthenticationDrawer'
import { useUIStore } from '@/store/uiStore'
import { TooltipProvider } from '@/components/ui/tooltip'

interface AppShellProps {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  const activeAuthMethod = useUIStore((s) => s.activeAuthMethod)
  const setActiveAuthMethod = useUIStore((s) => s.setActiveAuthMethod)
  const goBack = useUIStore((s) => s.goBack)
  const goForward = useUIStore((s) => s.goForward)

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (!(e.metaKey || e.ctrlKey)) return
      if (e.key !== '[' && e.key !== ']') return
      const target = e.target as HTMLElement | null
      if (target) {
        const tag = target.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || target.isContentEditable) return
      }
      e.preventDefault()
      if (e.key === '[') goBack()
      else goForward()
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [goBack, goForward])

  return (
    <TooltipProvider delayDuration={200}>
      <div className="h-screen w-screen flex flex-col bg-[#e5e5e5] dark:bg-transparent text-foreground overflow-hidden">
        <CommandPalette />
        <AddProjectDialog />
        <AuthenticationDrawer
          activeMethod={activeAuthMethod}
          onMethodChange={setActiveAuthMethod}
        />
        <DrawerMenu />

        <Topbar />

        <div className="flex flex-1 min-h-0 gap-2 px-2 pb-2 pt-0">
          <div className="rounded-xl border border-border bg-sidebar/70 backdrop-blur-xl backdrop-saturate-150 overflow-hidden shadow-sm">
            <Sidebar />
          </div>
          <div className="flex-1 flex flex-col min-w-0 min-h-0 rounded-xl border border-border bg-sidebar overflow-hidden shadow-sm">
            <TabBar />
            <div className="flex-1 overflow-auto min-h-0">{children}</div>
          </div>
        </div>

        <StatusBar />
      </div>
    </TooltipProvider>
  )
}
