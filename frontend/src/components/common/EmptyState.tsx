import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  hint?: ReactNode
  action?: ReactNode
  className?: string
  size?: 'sm' | 'md'
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  hint,
  action,
  className,
  size = 'md',
}: EmptyStateProps) {
  return (
    <div className={cn('flex h-full w-full items-center justify-center p-6', className)}>
      <div
        className={cn(
          'w-full text-center space-y-2.5',
          size === 'sm' ? 'max-w-[260px]' : 'max-w-sm',
        )}
      >
        {Icon && (
          <div className="inline-flex items-center justify-center w-9 h-9 rounded-lg bg-muted/50 text-muted-foreground">
            <Icon className="w-4 h-4" strokeWidth={1.75} />
          </div>
        )}
        <div className="space-y-1">
          <h3 className="text-[13px] font-semibold tracking-tight text-foreground">{title}</h3>
          {description && (
            <p className="text-[12px] text-muted-foreground leading-relaxed">{description}</p>
          )}
        </div>
        {hint && <div className="pt-1 text-[11px] text-muted-foreground/80">{hint}</div>}
        {action && <div className="pt-1">{action}</div>}
      </div>
    </div>
  )
}

interface KbdProps {
  children: ReactNode
}

export function Kbd({ children }: KbdProps) {
  return (
    <kbd className="inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded border border-border/60 bg-muted/50 text-[10.5px] font-sans leading-none text-foreground/80">
      {children}
    </kbd>
  )
}
