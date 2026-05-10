import { forwardRef } from 'react'
import { cn } from '@/lib/utils'

interface IslandProps extends React.HTMLAttributes<HTMLDivElement> {
  as?: 'div' | 'aside' | 'section' | 'main'
}

export const Island = forwardRef<HTMLDivElement, IslandProps>(function Island(
  { as: Tag = 'div', className, children, ...rest },
  ref,
) {
  const Component = Tag as React.ElementType
  return (
    <Component
      ref={ref}
      className={cn(
        'rounded-md border border-border/40 bg-card/30 flex flex-col overflow-hidden min-h-0',
        className,
      )}
      {...rest}
    >
      {children}
    </Component>
  )
})

interface IslandHeaderProps {
  children: React.ReactNode
  className?: string
}

export function IslandHeader({ children, className }: IslandHeaderProps) {
  return (
    <header
      className={cn(
        'h-9 px-3 shrink-0 border-b border-border/40 flex items-center gap-2 text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground',
        className,
      )}
    >
      {children}
    </header>
  )
}

interface IslandFooterProps {
  children: React.ReactNode
  className?: string
}

export function IslandFooter({ children, className }: IslandFooterProps) {
  return (
    <footer className={cn('shrink-0 border-t border-border/40 px-3 py-2', className)}>
      {children}
    </footer>
  )
}

export function IslandBody({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return <div className={cn('flex-1 min-h-0 overflow-auto', className)}>{children}</div>
}
