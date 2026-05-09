import { useMemo } from 'react'
import { ChevronRight, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  index: number
  row: any
  expanded: boolean
  onToggle: () => void
  onUpdate: (v: string) => void
  onRemove: () => void
}

export function DatasetRow({ index, row, expanded, onToggle, onUpdate, onRemove }: Props) {
  const summary = useMemo(() => {
    if (!row || typeof row !== 'object') return JSON.stringify(row)
    const entries = Object.entries(row).slice(0, 3)
    return entries.map(([k, v]) => `${k}: ${typeof v === 'string' ? v : JSON.stringify(v)}`).join(' · ')
  }, [row])

  return (
    <li className="group">
      <div
        className="flex items-center gap-2 px-3 py-1.5 hover:bg-accent/20 cursor-pointer"
        onClick={onToggle}
      >
        <ChevronRight
          className={cn(
            'w-3 h-3 text-muted-foreground/60 shrink-0 transition-transform',
            expanded && 'rotate-90',
          )}
        />
        <span className="text-[10px] font-mono text-muted-foreground tabular-nums w-6 shrink-0">
          #{index + 1}
        </span>
        <span className="text-[11px] font-mono text-foreground/75 truncate flex-1">{summary}</span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive shrink-0"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>
      {expanded && (
        <textarea
          defaultValue={JSON.stringify(row, null, 2)}
          onBlur={(e) => onUpdate(e.target.value)}
          spellCheck={false}
          className="w-full bg-input/20 text-[11px] font-mono px-3 py-2 outline-none resize-y min-h-[120px] max-h-72 border-t border-border/20"
        />
      )}
    </li>
  )
}
