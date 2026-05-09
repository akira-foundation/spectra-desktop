import { useEffect, useMemo, useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Compass, LayoutDashboard, Layers, FileText, Settings, Search } from 'lucide-react'
import { cn } from '@/lib/utils'

export function CommandPalette() {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const isCommandPaletteOpen = useUIStore((state) => state.isCommandPaletteOpen)
  const setCommandPaletteOpen = useUIStore((state) => state.setCommandPaletteOpen)
  const setCurrentPage = useUIStore((state) => state.setCurrentPage)
  const setSelectedEndpoint = useUIStore((state) => state.setSelectedEndpoint)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const endpoints = useEndpointsStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_EP : EMPTY_EP,
  )
  const { getMethodColor } = useHttpMethod()

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.key === 'k' || e.key === 'p') && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
        setCommandPaletteOpen(!isCommandPaletteOpen)
      }
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

  useEffect(() => {
    if (!open) setQuery('')
  }, [open])

  const navigation = [
    { label: 'Dashboard', icon: LayoutDashboard, shortcut: '⌘1', page: 'dashboard' as const },
    { label: 'API Inspector', icon: Compass, shortcut: '⌘2', page: 'inspector' as const },
    { label: 'Collections', icon: Layers, shortcut: '⌘3', page: 'collections' as const },
    { label: 'Snapshots', icon: FileText, shortcut: '⌘4', page: 'changelog' as const },
    { label: 'Settings', icon: Settings, shortcut: '⌘,', page: 'settings' as const },
  ]

  const filteredEndpoints = useMemo(() => {
    if (!query.trim()) return endpoints.slice(0, 3)
    const q = query.toLowerCase()
    return endpoints
      .filter((e) => `${e.method} ${e.path}`.toLowerCase().includes(q))
      .slice(0, 50)
  }, [endpoints, query])

  const close = () => {
    setOpen(false)
    setCommandPaletteOpen(false)
  }

  const handleNav = (page: NonNullable<(typeof navigation)[number]['page']>) => {
    close()
    setCurrentPage(page)
  }

  const handleEndpoint = (id: string) => {
    if (!activeProjectId) return
    close()
    setSelectedEndpoint(activeProjectId, id)
    setCurrentPage('inspector')
  }

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput
        placeholder="Search endpoints or commands..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList className="!min-h-[420px] !max-h-[420px]">
        <CommandEmpty>
          <div className="flex flex-col items-center justify-center gap-2 py-8 text-muted-foreground/70">
            <Search className="w-6 h-6 opacity-40" />
            <p className="text-[12px]">No matches for <code className="font-mono text-foreground/80">{query}</code></p>
            <p className="text-[10.5px] text-muted-foreground/50">Try method, path or page name</p>
          </div>
        </CommandEmpty>
        {filteredEndpoints.length > 0 && (
          <CommandGroup heading="Endpoints">
            {filteredEndpoints.map((e) => (
              <CommandItem
                key={e.id}
                value={`${e.method} ${e.path}`}
                onSelect={() => handleEndpoint(e.id)}
              >
                <span
                  className={cn(
                    'inline-flex w-12 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(e.method),
                  )}
                >
                  {e.method}
                </span>
                <span className="font-mono text-[12px] truncate">{e.path}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
        <CommandGroup heading="Navigation">
          {navigation.map((item) => {
            const Icon = item.icon
            return (
              <CommandItem
                key={item.label}
                onSelect={() => item.page && handleNav(item.page)}
              >
                <Icon className="w-4 h-4" />
                <span>{item.label}</span>
                <span className="ml-auto text-xs text-muted-foreground">{item.shortcut}</span>
              </CommandItem>
            )
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}

const EMPTY_EP: any[] = []
