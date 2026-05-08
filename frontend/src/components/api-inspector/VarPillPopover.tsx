import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { Pencil } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface VarPillPopoverProps {
  name: string
  value?: string
  variables: Record<string, string>
  rect: { left: number; top: number; bottom: number }
  onClose: () => void
  onSwap: (newName: string) => void
}

export function VarPillPopover({
  name,
  value,
  variables,
  rect,
  onClose,
  onSwap,
}: VarPillPopoverProps) {
  const ref = useRef<HTMLDivElement>(null)
  const setEditingEnvId = useUIStore((s) => s.setEditingEnvId)
  const activeEnvId = useProjectStore(
    (s) => s.projects.find((p) => p.id === s.activeProjectId)?.activeEnvironmentId ?? null,
  )

  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose()
      }
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('mousedown', onDoc)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDoc)
      document.removeEventListener('keydown', onKey)
    }
  }, [onClose])

  const names = Object.keys(variables)
  const hasValue = value !== undefined && value !== ''

  const top = rect.bottom + 4
  const left = rect.left

  return createPortal(
    <div
      ref={ref}
      style={{ position: 'fixed', top, left, zIndex: 60 }}
      className="w-[320px] rounded-md border border-border bg-popover text-popover-foreground shadow-md p-2.5 space-y-2"
    >
      <div className="flex items-center gap-2">
        <label className="text-[10.5px] uppercase tracking-wider text-muted-foreground shrink-0">
          Variable
        </label>
        <select
          value={name}
          onChange={(e) => onSwap(e.target.value)}
          className="flex-1 h-7 rounded-md border border-border/50 bg-muted/40 text-[12px] px-2"
        >
          {!names.includes(name) && <option value={name}>{name}</option>}
          {names.map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
        <Button
          size="sm"
          variant="outline"
          className="h-7 px-2 text-[11px]"
          disabled={!activeEnvId}
          onClick={() => {
            if (activeEnvId) setEditingEnvId(activeEnvId)
            onClose()
          }}
          title="Open environment editor"
        >
          <Pencil className="w-3 h-3" />
          Edit
        </Button>
      </div>
      <div
        className={cn(
          'rounded-md border px-2 py-1.5 text-[11.5px] font-mono break-all',
          hasValue
            ? 'border-amber-500/30 bg-amber-500/5 text-amber-200'
            : 'border-rose-500/30 bg-rose-500/5 text-rose-300 italic',
        )}
      >
        {hasValue ? value : 'unset'}
      </div>
    </div>,
    document.body,
  )
}
