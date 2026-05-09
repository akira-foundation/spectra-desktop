import { useEffect, useMemo, useState } from 'react'
import { Trash2, Eye, EyeOff, Copy, Crosshair, Search } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { capturesService, type CapturedValue } from '@/services/capturesService'
import { useUIStore } from '@/store/uiStore'
import { cn } from '@/lib/utils'

interface CapturedValuesPopoverProps {
  projectId: string | null
  values: CapturedValue[]
  onChange?: (values: CapturedValue[]) => void
}

export function CapturedValuesPopover({ projectId, values, onChange }: CapturedValuesPopoverProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [revealed, setRevealed] = useState<Record<string, boolean>>({})
  const compact = useUIStore((s) => s.compactToolbar)

  useEffect(() => {
    if (!open) {
      setQuery('')
      setRevealed({})
    }
  }, [open])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return values
    return values.filter(
      (v) =>
        v.name.toLowerCase().includes(q) ||
        v.value.toLowerCase().includes(q) ||
        (v.endpointKey ?? '').toLowerCase().includes(q),
    )
  }, [values, query])

  const groups = useMemo(() => {
    const map = new Map<string, CapturedValue[]>()
    for (const v of filtered) {
      const key = v.endpointKey || 'Unassigned'
      const list = map.get(key) ?? []
      list.push(v)
      map.set(key, list)
    }
    return Array.from(map.entries())
  }, [filtered])

  const toggleReveal = (name: string) =>
    setRevealed((r) => ({ ...r, [name]: !r[name] }))

  const refresh = async () => {
    if (!projectId) return
    const v = await capturesService.listValues(projectId)
    onChange?.(v)
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
          onClick={() => void refresh()}
          title="Captured runtime variables"
          className={cn(
            'inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-border/50 bg-card text-[11px] text-muted-foreground hover:text-foreground hover:bg-accent/60 transition-colors',
            values.length > 0 && 'text-foreground',
          )}
        >
          <Crosshair className="w-3 h-3" />
          {!compact && <span>Captures</span>}
          <span className="font-mono text-[10.5px] tabular-nums text-muted-foreground">
            {values.length}
          </span>
        </button>
      </PopoverTrigger>

      <PopoverContent
        align="end"
        className="w-110 p-0 overflow-y-auto"
        style={{ maxHeight: 'min(560px, 70vh)' }}
      >
        <div className={cn('sticky top-0 z-30 bg-popover', values.length > 0 ? 'h-19.5' : 'h-9')}>
          <Header
            count={values.length}
            filteredCount={filtered.length}
            query={query}
            onClear={() => void clearAll()}
          />
          {values.length > 0 && (
            <SearchBar value={query} onChange={setQuery} />
          )}
        </div>
        {values.length === 0 ? (
          <EmptyState message="No captures yet. Define them in the Captures tab and re-execute." />
        ) : filtered.length === 0 ? (
          <EmptyState message="No matches." />
        ) : (
          groups.map(([key, items]) => (
            <Group
              key={key}
              endpoint={key}
              items={items}
              revealed={revealed}
              onToggle={toggleReveal}
            />
          ))
        )}
      </PopoverContent>
    </Popover>
  )
}

function Header({
  count,
  filteredCount,
  query,
  onClear,
}: {
  count: number
  filteredCount: number
  query: string
  onClear: () => void
}) {
  return (
    <div className="h-9 px-3 border-b border-border/40 flex items-center justify-between shrink-0">
      <div className="flex items-center gap-1.5">
        <Crosshair className="w-3 h-3 text-muted-foreground" />
        <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          Captured variables
        </span>
        <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">
          {query ? `${filteredCount}/${count}` : count}
        </span>
      </div>
      {count > 0 && (
        <button
          type="button"
          onClick={onClear}
          className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="w-3 h-3" />
          Clear all
        </button>
      )}
    </div>
  )
}

function SearchBar({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div className="px-2 py-1.5 border-b border-border/40 shrink-0">
      <div className="relative">
        <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground/60 pointer-events-none" />
        <input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Filter by name, value, or endpoint…"
          className="w-full h-7 pl-7 pr-2 rounded-md bg-input/30 border border-border/40 text-[11px] outline-none focus:border-border placeholder:text-muted-foreground/50"
        />
      </div>
    </div>
  )
}

function EmptyState({ message }: { message: string }) {
  return (
    <p className="px-4 py-10 text-[11px] italic text-muted-foreground/70 text-center">
      {message}
    </p>
  )
}

function Group({
  endpoint,
  items,
  revealed,
  onToggle,
}: {
  endpoint: string
  items: CapturedValue[]
  revealed: Record<string, boolean>
  onToggle: (name: string) => void
}) {
  const [method, ...rest] = endpoint.split(' ')
  const path = rest.join(' ')
  return (
    <section>
      <header className="sticky top-19.5 z-20 h-7 px-3 bg-popover border-b border-border/40 flex items-center gap-2">
        <span className="text-[9.5px] font-mono font-semibold text-primary/80 shrink-0">
          {method}
        </span>
        <span className="text-[10.5px] font-mono text-foreground/70 truncate">{path}</span>
        <span className="ml-auto text-[10px] font-mono text-muted-foreground/50 tabular-nums shrink-0">
          {items.length}
        </span>
      </header>
      <ul className="m-0 p-0 list-none">
        {items.map((v) => (
          <Row key={v.name} value={v} show={!!revealed[v.name]} onToggle={() => onToggle(v.name)} />
        ))}
      </ul>
    </section>
  )
}

function Row({
  value: v,
  show,
  onToggle,
}: {
  value: CapturedValue
  show: boolean
  onToggle: () => void
}) {
  return (
    <li className="group h-7 flex items-center gap-2 px-3 hover:bg-accent/30 min-w-0">
      <code className="text-[11.5px] font-mono text-foreground shrink-0 max-w-[42%] truncate">
        {v.name}
      </code>
      <span className="text-muted-foreground/30 text-[10.5px] shrink-0">=</span>
      <code className="flex-1 text-[10.5px] font-mono text-muted-foreground truncate">
        {show ? v.value : mask(v.value)}
      </code>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
        <IconButton title={show ? 'Hide' : 'Show'} onClick={onToggle}>
          {show ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
        </IconButton>
        <IconButton title="Copy" onClick={() => void navigator.clipboard.writeText(v.value)}>
          <Copy className="w-3 h-3" />
        </IconButton>
      </div>
    </li>
  )
}

function IconButton({
  children,
  onClick,
  title,
}: {
  children: React.ReactNode
  onClick: () => void
  title: string
}) {
  return (
    <button
      type="button"
      title={title}
      onClick={onClick}
      className="inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/60"
    >
      {children}
    </button>
  )
}

function mask(v: string): string {
  if (!v) return ''
  if (v.length <= 6) return '•'.repeat(v.length)
  return v.slice(0, 3) + '•'.repeat(Math.min(v.length - 6, 12)) + v.slice(-3)
}
