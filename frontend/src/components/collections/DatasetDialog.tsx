import { useEffect, useState } from 'react'
import { Loader2, Sparkles, Trash2, Database } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogHeader,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { datasetsService } from '@/services/datasetsService'
import { cn } from '@/lib/utils'
import { Switch } from './Switch'
import { DatasetRow } from './DatasetRow'

interface Props {
  projectId: string
  endpointId: string
  method: string
  path: string
  iterating: boolean
  onToggleIterate: () => void
  onClose: () => void
}

export function DatasetDialog({
  projectId,
  endpointId,
  method,
  path,
  iterating,
  onToggleIterate,
  onClose,
}: Props) {
  const endpointKey = `${method.toUpperCase()} ${path}`
  const [rows, setRows] = useState<unknown[]>([])
  const [count, setCount] = useState(10)
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)
  const [expanded, setExpanded] = useState<number | null>(null)

  useEffect(() => {
    setLoading(true)
    void datasetsService.get(projectId, endpointKey).then((r) => {
      setRows(r)
      if (r.length > 0) setCount(r.length)
      setLoading(false)
    })
  }, [projectId, endpointKey])

  const persist = async (next: unknown[]) => {
    setRows(next)
    await datasetsService.save(projectId, endpointKey, next)
  }

  const generate = async () => {
    setGenerating(true)
    try {
      const next = await datasetsService.generate(endpointId, count)
      await persist(next)
    } finally {
      setGenerating(false)
    }
  }

  const removeRow = async (i: number) => {
    await persist(rows.filter((_, idx) => idx !== i))
  }

  const updateRow = async (i: number, value: string) => {
    try {
      const parsed = JSON.parse(value)
      await persist(rows.map((r, idx) => (idx === i ? parsed : r)))
    } catch {}
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
          <DialogTitle className="text-base">Dataset</DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Run this request multiple times with different payloads generated from the schema.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4 space-y-3">
          <div className="rounded-md border border-border/60 bg-card/40 p-3 flex items-center gap-3">
            <div className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-emerald-500/10 text-emerald-500 shrink-0">
              <Database className="w-4 h-4" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
                Endpoint
              </p>
              <code className="text-[11.5px] font-mono text-foreground/85 truncate block">{endpointKey}</code>
            </div>
            <div className="flex flex-col items-end gap-1 shrink-0">
              <Switch checked={iterating} onCheckedChange={onToggleIterate} />
              <span
                className={cn(
                  'text-[10px] font-medium',
                  iterating ? 'text-emerald-500' : 'text-muted-foreground',
                )}
              >
                {iterating ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>

          <div className="rounded-md border border-border/60 bg-card/40">
            <div className="px-3 py-2 border-b border-border/40 flex items-center gap-2">
              <Sparkles className="w-3 h-3 text-muted-foreground" />
              <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
                Payloads
              </span>
              <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">{rows.length}</span>
              <div className="ml-auto flex items-center gap-1.5">
                <Input
                  type="number"
                  value={count}
                  min={1}
                  max={500}
                  onChange={(e) =>
                    setCount(Math.max(1, Math.min(500, Number(e.target.value) || 1)))
                  }
                  className="h-7 w-14 text-[11px] font-mono text-center"
                />
                <Button
                  size="sm"
                  variant="outline"
                  className="h-7 px-2.5 text-[10.5px] gap-1.5"
                  onClick={generate}
                  disabled={generating}
                >
                  {generating ? (
                    <Loader2 className="w-3 h-3 animate-spin" />
                  ) : (
                    <Sparkles className="w-3 h-3 text-emerald-500" />
                  )}
                  Generate
                </Button>
                {rows.length > 0 && (
                  <button
                    type="button"
                    onClick={() => void persist([])}
                    className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-destructive ml-1"
                  >
                    <Trash2 className="w-3 h-3" />
                    Clear
                  </button>
                )}
              </div>
            </div>
            {loading ? (
              <p className="px-4 py-8 text-[11px] italic text-muted-foreground/70 text-center">Loading…</p>
            ) : rows.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 px-6 py-8 text-center">
                <Sparkles className="w-5 h-5 text-muted-foreground/40" />
                <p className="text-[11.5px] text-muted-foreground/70">
                  No payloads yet. Set a count and click Generate.
                </p>
              </div>
            ) : (
              <ul className="m-0 p-0 list-none divide-y divide-border/20 max-h-96 overflow-y-auto">
                {rows.map((row, i) => (
                  <DatasetRow
                    key={i}
                    index={i}
                    row={row}
                    expanded={expanded === i}
                    onToggle={() => setExpanded((e) => (e === i ? null : i))}
                    onUpdate={(v) => void updateRow(i, v)}
                    onRemove={() => void removeRow(i)}
                  />
                ))}
              </ul>
            )}
          </div>
        </div>

        <DialogFooter className="px-6 py-3 shrink-0 border-t border-border/40">
          <Button variant="outline" size="sm" onClick={onClose}>
            Done
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
