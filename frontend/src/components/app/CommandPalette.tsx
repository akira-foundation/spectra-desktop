import { useEffect, useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { Compass, LayoutDashboard, Layers, FileText, Settings, HelpCircle } from 'lucide-react'

export function CommandPalette() {
  const [open, setOpen] = useState(false)
  const isCommandPaletteOpen = useUIStore((state) => state.isCommandPaletteOpen)
  const setCommandPaletteOpen = useUIStore((state) => state.setCommandPaletteOpen)
  const setCurrentPage = useUIStore((state) => state.setCurrentPage)

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
        setCommandPaletteOpen(!isCommandPaletteOpen)
      }
      // Navigation shortcuts
      if (e.metaKey || e.ctrlKey) {
        switch (e.key) {
          case '1':
            e.preventDefault()
            setCurrentPage('dashboard')
            break
          case '2':
            e.preventDefault()
            setCurrentPage('inspector')
            break
          case '3':
            e.preventDefault()
            setCurrentPage('collections')
            break
          case '4':
            e.preventDefault()
            setCurrentPage('changelog')
            break
          case ',':
            e.preventDefault()
            setCurrentPage('settings')
            break
        }
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [isCommandPaletteOpen, setCommandPaletteOpen, setCurrentPage])

  useEffect(() => {
    setOpen(isCommandPaletteOpen)
  }, [isCommandPaletteOpen])

  const commands = [
    {
      group: 'Navigation',
      items: [
        { label: 'Dashboard', icon: LayoutDashboard, shortcut: '⌘1', page: 'dashboard' as const },
        { label: 'API Inspector', icon: Compass, shortcut: '⌘2', page: 'inspector' as const },
        { label: 'Collections', icon: Layers, shortcut: '⌘3', page: 'collections' as const },
        { label: 'Snapshots', icon: FileText, shortcut: '⌘4', page: 'changelog' as const },
      ],
    },
    {
      group: 'Settings',
      items: [
        { label: 'Settings', icon: Settings, shortcut: '⌘,', page: 'settings' as const },
        { label: 'Help', icon: HelpCircle, shortcut: '?' },
      ],
    },
  ]

  const handleSelectCommand = (item: any) => {
    setOpen(false)
    setCommandPaletteOpen(false)
    if (item.page) {
      setCurrentPage(item.page)
    }
  }

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Search commands..." />
      <CommandList>
        <CommandEmpty>No commands found.</CommandEmpty>
        {commands.map((group) => (
          <CommandGroup key={group.group} heading={group.group}>
            {group.items.map((item) => {
              const Icon = item.icon
              return (
                <CommandItem
                  key={item.label}
                  onSelect={() => handleSelectCommand(item)}
                >
                  <Icon className="w-4 h-4" />
                  <span>{item.label}</span>
                  <span className="ml-auto text-xs text-muted-foreground">{item.shortcut}</span>
                </CommandItem>
              )
            })}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  )
}
