import { ReactNode } from 'react'
import { Topbar } from './Topbar'
import { Sidebar } from './Sidebar'
import { TabBar } from './TabBar'
import { StatusBar } from './StatusBar'
import { DrawerMenu } from './DrawerMenu'
import { CommandPalette } from './CommandPalette'
import { AddProjectDialog } from '@/components/projects/AddProjectDialog'
import { TooltipProvider } from '@/components/ui/tooltip'

interface AppShellProps {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  return (
    <TooltipProvider delayDuration={200}>
      <div className="h-screen bg-background text-foreground flex flex-col overflow-hidden">
        <CommandPalette />
        <AddProjectDialog />
        <DrawerMenu />
        <Topbar />
        <div className="flex flex-1 min-h-0">
          <Sidebar />
          <div className="flex-1 flex flex-col min-w-0 min-h-0">
            <TabBar />
            <div className="flex-1 overflow-auto min-h-0">{children}</div>
          </div>
        </div>
        <StatusBar />
      </div>
    </TooltipProvider>
  )
}
