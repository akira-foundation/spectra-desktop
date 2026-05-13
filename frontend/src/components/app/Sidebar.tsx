import { Compass, LayoutDashboard, FileText, Layers, Settings, Terminal, KeyRound, Server } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { useUIStore, type PageType } from '@/store/uiStore'
import { useUpdatesStore, isUpdateActionable } from '@/store/updatesStore'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

interface NavItem {
  id: string
  label: string
  icon: LucideIcon
  page?: PageType
  disabled?: boolean
}

const primary: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, page: 'dashboard' },
  { id: 'inspector', label: 'API Inspector', icon: Compass, page: 'inspector' },
  { id: 'collections', label: 'Collections', icon: Layers, page: 'collections' },
  { id: 'scratch', label: 'Scratch', icon: Terminal, page: 'scratch' },
  { id: 'accounts', label: 'Accounts', icon: KeyRound, page: 'accounts' },
  { id: 'mock', label: 'Mock server', icon: Server, page: 'mock' },
  { id: 'changelog', label: 'Snapshots', icon: FileText, page: 'changelog' },
]

const secondary: NavItem[] = [
  { id: 'settings', label: 'Settings', icon: Settings, page: 'settings' },
]

export function Sidebar() {
  const currentPage = useUIStore((s) => s.currentPage)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)
  const updatePhase = useUpdatesStore((s) => s.phase)
  const updateActionable = isUpdateActionable(updatePhase)

  return (
    <aside className="flex h-full w-12 shrink-0 flex-col items-center justify-between py-2">
      <div className="flex flex-col items-center gap-0.5 pt-1">
        <span className="mb-2 inline-block h-8 w-8">
          <img
            src="/favicon-light.svg"
            alt="Spectra"
            className="h-8 w-8 rounded-md dark:hidden"
          />
          <img
            src="/favicon.svg"
            alt="Spectra"
            className="h-8 w-8 rounded-md hidden dark:block"
          />
        </span>
        {primary.map((item) => (
          <Item
            key={item.id}
            item={item}
            active={!!item.page && currentPage === item.page}
            onClick={() => item.page && setCurrentPage(item.page)}
          />
        ))}
      </div>
      <div className="flex flex-col items-center gap-0.5">
        {secondary.map((item) => (
          <Item
            key={item.id}
            item={item}
            active={!!item.page && currentPage === item.page}
            onClick={() => item.page && setCurrentPage(item.page)}
            badge={item.id === 'settings' && updateActionable}
          />
        ))}
      </div>
    </aside>
  )
}

interface ItemProps {
  item: NavItem
  active: boolean
  onClick: () => void
  badge?: boolean
}

function Item({ item, active, onClick, badge }: ItemProps) {
  const Icon = item.icon
  return (
    <Tooltip delayDuration={150}>
      <TooltipTrigger asChild>
        <button
          onClick={item.disabled ? undefined : onClick}
          aria-disabled={item.disabled}
          className={cn(
            'relative flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors',
            'hover:bg-accent/60 hover:text-foreground',
            active && 'bg-accent text-foreground',
            item.disabled && 'opacity-40 cursor-not-allowed hover:bg-transparent hover:text-muted-foreground',
          )}
        >
          <Icon className="h-[15px] w-[15px]" strokeWidth={1.75} />
          {badge && (
            <span className="absolute top-1 right-1 h-1.5 w-1.5 rounded-full bg-primary ring-2 ring-sidebar" />
          )}
        </button>
      </TooltipTrigger>
      <TooltipContent side="right" className="text-[11px]">
        {item.label}
        {badge && <span className="ml-1 text-primary">· update</span>}
        {item.disabled && <span className="ml-1 opacity-60">· soon</span>}
      </TooltipContent>
    </Tooltip>
  )
}
