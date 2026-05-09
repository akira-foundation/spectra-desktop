import { ArrowUpRight } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

interface Props {
  title: string
  icon: LucideIcon
  action?: { label: string; onClick: () => void }
  headerExtra?: React.ReactNode
  children: React.ReactNode
}

export function Card({ title, icon: Icon, action, headerExtra, children }: Props) {
  return (
    <section className="rounded-lg border border-border/60 bg-card/40 p-3.5 space-y-2.5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <Icon className="w-3.5 h-3.5 text-muted-foreground" />
          <h2 className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            {title}
          </h2>
        </div>
        <div className="flex items-center gap-1.5">
          {headerExtra}
          {action && (
            <button
              type="button"
              onClick={action.onClick}
              className="inline-flex items-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground transition-colors"
            >
              {action.label}
              <ArrowUpRight className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
      {children}
    </section>
  )
}
