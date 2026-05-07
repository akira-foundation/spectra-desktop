import { ReactNode } from 'react'
import { Topbar } from './Topbar'
import { TabBar } from './TabBar'
import { StatusBar } from './StatusBar'
import { DrawerMenu } from './DrawerMenu'
import { CommandPalette } from './CommandPalette'

interface AppShellProps {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  return (
    <div className="h-screen bg-background text-foreground flex flex-col">
      {/* Command Palette */}
      <CommandPalette />

      {/* Drawer Menu */}
      <DrawerMenu />

      {/* Topbar */}
      <Topbar />

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Main Viewport */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Tabs */}
          <TabBar />

          {/* Content Area */}
          <div className="flex-1 overflow-auto">
            {children}
          </div>
        </div>
      </div>

      {/* Status Bar */}
      <StatusBar />
    </div>
  )
}
