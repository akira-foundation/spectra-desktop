import { useEffect, useState } from 'react'
import { Trash2, Eye, EyeOff, Copy, Crosshair } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { capturesService, type CapturedValue } from '@/services/capturesService'
import { cn } from '@/lib/utils'

interface CapturedValuesPopoverProps {
  projectId: string | null
  values: CapturedValue[]
  onChange?: (values: CapturedValue[]) => void
}

export function CapturedValuesPopover({ projectId, values, onChange }: CapturedValuesPopoverProps) {
  const [open, setOpen] = useState(false)
  const [revealed, setRevealed] = useState<Record<string, boolean>>({})

  useEffect(() => {
    if (!open) setRevealed({})
  }, [open])

  const refresh = async () => {
    if (!projectId) return
    const vals = await capturesService.listValues(projectId)
    onChange?.(vals)
  }

  const clearAll = async () => {
    if (!projectId) return
    await capturesService.clearValues(projectId)
    onChange?.([])
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            'inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-border/50 bg-card hover:bg-accent/60 text-[11px] text-muted-foreground hover:text-foreground transition-colors',
            values.length > 0 && 'text-foreground',
          )}
          onClick={() => void refresh()}
          title="Captured runtime variables"
        >
          <Crosshair className="w-3 h-3" />
          <span className="text-[10.5px]">Captures</span>
          <span className="font-mono text-[10.5px] tabular-nums text-muted-foreground">{values.length}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[420px] p-0">
        <div className="px-3 py-2 border-b border-border/40 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
              Captured variables
            </span>
            <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{values.length}</span>
          </div>
          {values.length > 0 && (
            <button
              type="button"
              onClick={() => void clearAll()}
              className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="w-3 h-3" />
              Clear
            </button>
          )}
        </div>
        {values.length === 0 ? (
          <p className="px-3 py-5 text-[11px] italic text-muted-foreground/80 text-center">
            No captures yet. Define them in the Captures tab and re-execute.
          </p>
        ) : (
          <ul className="max-h-80 overflow-auto divide-y divide-border/30">
            {values.map((v) => {
              const show = !!revealed[v.name]
              return (
                <li key={v.name} className="px-3 py-2 hover:bg-accent/30 group">
                  <div className="flex items-baseline gap-2 min-w-0">
                    <code className="text-[11.5px] font-mono text-foreground shrink-0">
                      {v.name}
                    </code>
                    <span className="text-muted-foreground/40 text-[11px]">=</span>
                    <code className="flex-1 text-[10.5px] font-mono text-muted-foreground truncate">
                      {show ? v.value : mask(v.value)}
                    </code>
                    <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onClick={() => setRevealed((r) => ({ ...r, [v.name]: !r[v.name] }))}
                        className="inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/60"
                        title={show ? 'Hide' : 'Show'}
                      >
                        {show ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                      </button>
                      <button
                        type="button"
                        onClick={() => void navigator.clipboard.writeText(v.value)}
                        className="inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/60"
                        title="Copy value"
                      >
                        <Copy className="w-3 h-3" />
                      </button>
                    </div>
                  </div>
                  {v.endpointKey && (
                    <p className="text-[9.5px] font-mono text-muted-foreground/50 truncate mt-0.5">
                      {v.endpointKey}
                    </p>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </PopoverContent>
    </Popover>
  )
}

function mask(v: string): string {
  if (!v) return ''
  if (v.length <= 6) return '•'.repeat(v.length)
  return v.slice(0, 3) + '•'.repeat(Math.min(v.length - 6, 12)) + v.slice(-3)
}
