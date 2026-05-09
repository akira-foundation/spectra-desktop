import { useState } from 'react'
import { cn } from '@/lib/utils'
import { formatLeaf, valueTone } from '@/lib/format'

interface Props {
  value: unknown
  depth?: number
  label?: string
}

export function TreeView({ value, depth = 0, label }: Props) {
  const [open, setOpen] = useState(depth < 2)
  const isObj = value && typeof value === 'object'
  const isArr = Array.isArray(value)
  if (!isObj) {
    return (
      <div className="flex items-baseline gap-2 py-0.5" style={{ paddingLeft: depth * 12 }}>
        {label && <code className="text-[11px] font-mono text-foreground/70">{label}:</code>}
        <code className={cn('text-[11px] font-mono', valueTone(value))}>{formatLeaf(value)}</code>
      </div>
    )
  }
  const entries = isArr
    ? (value as unknown[]).map((v, i) => [`[${i}]`, v] as [string, unknown])
    : Object.entries(value as Record<string, unknown>)
  const summary = isArr ? `Array(${entries.length})` : `{${entries.length}}`
  return (
    <div style={{ paddingLeft: depth * 12 }}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex items-baseline gap-2 py-0.5 hover:bg-accent/20 px-1 rounded"
      >
        <span className="text-[10px] text-muted-foreground/60 w-3">{open ? '▼' : '▶'}</span>
        {label && <code className="text-[11px] font-mono text-foreground/70">{label}:</code>}
        <code className="text-[10.5px] font-mono text-muted-foreground">{summary}</code>
      </button>
      {open && (
        <div>
          {entries.map(([k, v]) => (
            <TreeView key={k} value={v} depth={depth + 1} label={k} />
          ))}
        </div>
      )}
    </div>
  )
}
