import { cn } from '@/lib/utils'

interface Props {
  count: number
  active: number
}

export function StepDots({ count, active }: Props) {
  return (
    <div className="flex items-center gap-1.5">
      {Array.from({ length: count }).map((_, i) => (
        <span
          key={i}
          className={cn(
            'h-1 rounded-full transition-all duration-300',
            i === active ? 'w-6 bg-foreground/80' : 'w-1 bg-foreground/20',
          )}
        />
      ))}
    </div>
  )
}
