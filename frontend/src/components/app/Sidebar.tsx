import { Compass, FolderKanban, FileText, Settings, HelpCircle } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { useUIStore, type PageType } from '@/store/uiStore'
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
  { id: 'inspector', label: 'API Inspector', icon: Compass, page: 'inspector' },
  { id: 'dashboard', label: 'Dashboard', icon: FolderKanban, page: 'dashboard' },
  { id: 'changelog', label: 'Snapshots', icon: FileText, page: 'changelog' },
]

const secondary: NavItem[] = [
  { id: 'settings', label: 'Settings', icon: Settings, page: 'settings' },
  { id: 'help', label: 'Help', icon: HelpCircle, disabled: true },
]

export function Sidebar() {
  const currentPage = useUIStore((s) => s.currentPage)
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)

  return (
    <aside className="flex h-full w-12 shrink-0 flex-col items-center justify-between py-2">
      <div className="flex flex-col items-center gap-0.5 pt-1">
        <div className="mb-2 flex h-8 w-8 items-center justify-center rounded-md bg-primary/15 text-primary text-[13px] font-semibold tracking-tight">
          S
        </div>
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
}

function Item({ item, active, onClick }: ItemProps) {
  const Icon = item.icon
  return (
    <Tooltip delayDuration={150}>
      <TooltipTrigger asChild>
        <button
          onClick={item.disabled ? undefined : onClick}
          aria-disabled={item.disabled}
          className={cn(
            'flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors',
            'hover:bg-accent/60 hover:text-foreground',
            active && 'bg-accent text-foreground',
            item.disabled && 'opacity-40 cursor-not-allowed hover:bg-transparent hover:text-muted-foreground',
          )}
        >
          <Icon className="h-[15px] w-[15px]" strokeWidth={1.75} />
        </button>
      </TooltipTrigger>
      <TooltipContent side="right" className="text-[11px]">
        {item.label}
        {item.disabled && <span className="ml-1 opacity-60">· soon</span>}
      </TooltipContent>
    </Tooltip>
  )
}
