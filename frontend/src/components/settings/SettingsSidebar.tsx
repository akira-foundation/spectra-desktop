import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SettingsNavGroup {
  heading: string
  items: SettingsNavItem[]
}

export interface SettingsNavItem {
  id: string
  label: string
  icon: LucideIcon
  badge?: string
  disabled?: boolean
}

interface Props {
  groups: SettingsNavGroup[]
  activeId: string
  onSelect: (id: string) => void
}

export function SettingsSidebar({ groups, activeId, onSelect }: Props) {
  return (
    <aside className="w-60 h-full shrink-0 py-5 px-3 overflow-y-auto">
      {groups.map((group) => (
        <div key={group.heading} className="mb-5">
          <h3 className="px-2 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.08em] text-muted-foreground/70">
            {group.heading}
          </h3>
          <ul className="space-y-0.5">
            {group.items.map((item) => {
              const Icon = item.icon
              const active = item.id === activeId
              return (
                <li key={item.id}>
                  <button
                    type="button"
                    onClick={() => !item.disabled && onSelect(item.id)}
                    disabled={item.disabled}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-2.5 h-8 rounded-md text-[12.5px] transition-colors text-left',
                      active
                        ? 'bg-accent/60 text-foreground'
                        : 'text-muted-foreground hover:bg-accent/40 hover:text-foreground',
                      item.disabled && 'opacity-50 cursor-not-allowed hover:bg-transparent hover:text-muted-foreground',
                    )}
                  >
                    <Icon className="h-4 w-4 shrink-0" strokeWidth={1.75} />
                    <span className="flex-1 truncate">{item.label}</span>
                    {item.badge && (
                      <span className="text-[9px] uppercase tracking-wider text-muted-foreground/60 border border-border/40 rounded px-1 py-px">
                        {item.badge}
                      </span>
                    )}
                  </button>
                </li>
              )
            })}
          </ul>
        </div>
      ))}
    </aside>
  )
}
