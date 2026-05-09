import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

interface Props {
  endpoints: { id: string; method: string; path: string }[]
  onPick: (ids: string[]) => void
  onClose: () => void
}

export function EndpointPicker({ endpoints, onPick, onClose }: Props) {
  const { getMethodColor } = useHttpMethod()
  const [query, setQuery] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const filtered = endpoints
    .filter((e) => `${e.method} ${e.path}`.toLowerCase().includes(query.toLowerCase()))
    .slice(0, 500)

  const toggle = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const allFilteredSelected = filtered.length > 0 && filtered.every((e) => selected.has(e.id))

  const toggleAll = () => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (allFilteredSelected) {
        for (const e of filtered) next.delete(e.id)
      } else {
        for (const e of filtered) next.add(e.id)
      }
      return next
    })
  }

  const confirm = () => onPick(Array.from(selected))

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent
        className="p-0 gap-0 sm:max-w-xl w-[640px] flex flex-col overflow-hidden"
        style={{ height: '70vh' }}
      >
        <DialogTitle className="sr-only">Add requests</DialogTitle>
        <div className="px-3 py-2 border-b border-border/40 flex items-center gap-2 shrink-0">
          <Input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search endpoint…"
            className="h-8 text-[12px] border-0 bg-transparent focus-visible:ring-0 shadow-none"
          />
        </div>
        <div className="px-4 py-1.5 border-b border-border/30 flex items-center gap-2 shrink-0 bg-muted/20">
          <input
            type="checkbox"
            checked={allFilteredSelected}
            onChange={toggleAll}
            className="h-3.5 w-3.5 accent-primary cursor-pointer"
            aria-label="Select all"
          />
          <span className="text-[10px] text-muted-foreground">
            {allFilteredSelected ? 'Deselect all' : 'Select all'}
          </span>
          <span className="ml-auto text-[10px] font-mono text-muted-foreground/60 tabular-nums">
            {selected.size} selected · {filtered.length} of {endpoints.length}
          </span>
        </div>
        <ul className="m-0 p-0 list-none flex-1 overflow-y-auto">
          {filtered.length === 0 ? (
            <li className="px-3 py-8 text-[11px] italic text-muted-foreground/70 text-center">No matches.</li>
          ) : (
            filtered.map((e) => {
              const isSelected = selected.has(e.id)
              return (
                <li key={e.id}>
                  <label className="w-full flex items-center gap-2.5 px-4 py-1.5 text-[11.5px] hover:bg-accent/40 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggle(e.id)}
                      className="h-3.5 w-3.5 accent-primary cursor-pointer"
                    />
                    <span
                      className={cn(
                        'inline-flex w-12 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                        getMethodColor(e.method),
                      )}
                    >
                      {e.method}
                    </span>
                    <span className="font-mono text-foreground/85 truncate">{e.path}</span>
                  </label>
                </li>
              )
            })
          )}
        </ul>
        <div className="px-3 py-2 border-t border-border/40 flex items-center justify-end gap-2 shrink-0">
          <Button size="sm" variant="ghost" className="h-7 text-[11px]" onClick={onClose}>
            Cancel
          </Button>
          <Button size="sm" className="h-7 text-[11px]" onClick={confirm} disabled={selected.size === 0}>
            Add {selected.size > 0 ? `${selected.size} ` : ''}request{selected.size === 1 ? '' : 's'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
