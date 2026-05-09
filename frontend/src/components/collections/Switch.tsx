import { cn } from '@/lib/utils'

interface Props {
  checked: boolean
  onCheckedChange: (v: boolean) => void
  label?: string
}

export function Switch({ checked, onCheckedChange, label }: Props) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onCheckedChange(!checked)}
      className="inline-flex items-center gap-2 group"
    >
      <span
        className={cn(
          'relative inline-flex h-4 w-7 shrink-0 items-center rounded-full transition-colors',
          checked ? 'bg-emerald-500' : 'bg-muted',
        )}
      >
        <span
          className={cn(
            'inline-block h-3 w-3 transform rounded-full bg-background shadow transition-transform',
            checked ? 'translate-x-3.5' : 'translate-x-0.5',
          )}
        />
      </span>
      {label && (
        <span className={cn('text-[10.5px]', checked ? 'text-emerald-500' : 'text-muted-foreground')}>
          {label}
        </span>
      )}
    </button>
  )
}
