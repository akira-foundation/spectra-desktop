import { Home, Search, Package, FileText, GitCompare, Settings, HelpCircle, Zap } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { Drawer, DrawerContent, DrawerClose } from '@/components/ui/drawer'
import { Button } from '@/components/ui/button'

export function DrawerMenu() {
  const sidebarOpen = useUIStore((state) => state.sidebarOpen)
  const setSidebarOpen = useUIStore((state) => state.setSidebarOpen)
  const setCurrentPage = useUIStore((state) => state.setCurrentPage)

  const navItems = [
    { icon: Home, label: 'Dashboard', shortcut: '⌘1', page: 'dashboard' as const },
    { icon: Search, label: 'API Explorer', shortcut: '⌘2', page: 'inspector' as const },
    { icon: Package, label: 'Models', shortcut: '⌘3' },
    { icon: FileText, label: 'Changelog', shortcut: '⌘4' },
    { icon: GitCompare, label: 'Diff Viewer', shortcut: '⌘5' },
    { icon: Settings, label: 'Settings', shortcut: '⌘,', page: 'settings' as const },
    { icon: HelpCircle, label: 'Help', shortcut: '?' },
  ]

  const handleNavClick = (item: typeof navItems[0]) => {
    setSidebarOpen(false)
    if (item.page) {
      setCurrentPage(item.page)
    }
  }

  return (
    <Drawer open={sidebarOpen} onOpenChange={setSidebarOpen} direction="left">
      <DrawerContent className="w-44">
        {/* Logo */}
        <div className="flex items-center gap-4 p-4 border-b border-border">
          <div className="w-10 h-10 rounded-xl gradient-primary flex items-center justify-center flex-shrink-0">
            <Zap className="w-7 h-7 text-white" />
          </div>
          <div>
            <p className="font-bold text-base tracking-wide text-foreground">Spectra</p>
            <p className="text-xs text-foreground/50">Professional API Inspector</p>
          </div>
        </div>

        {/* Navigation */}
        <nav className="overflow-y-auto px-2 py-4 space-y-1 flex-1">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <DrawerClose asChild key={item.label}>
                <Button
                  variant="ghost"
                  className="w-full justify-start gap-3 px-4"
                  onClick={() => handleNavClick(item)}
                >
                  <Icon className="w-5 h-5 flex-shrink-0" />
                  <span className="flex-1 text-left">{item.label}</span>
                  <span className="text-xs text-foreground/40">{item.shortcut}</span>
                </Button>
              </DrawerClose>
            )
          })}
        </nav>

        {/* Footer */}
        <div className="border-t border-border p-2">
          <DrawerClose asChild>
            <Button variant="ghost" className="w-full justify-start gap-3 px-4">
              <HelpCircle className="w-5 h-5 flex-shrink-0" />
              <span>Support</span>
            </Button>
          </DrawerClose>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
