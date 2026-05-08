import { Plus, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { VarInput } from './VarInput'

export interface HeaderRow {
  key: string
  value: string
  enabled: boolean
}

interface HeadersEditorProps {
  headers: HeaderRow[]
  onAdd: () => void
  onChange: (index: number, patch: Partial<HeaderRow>) => void
  onRemove: (index: number) => void
  variables?: Record<string, string>
}

export function HeadersEditor({ headers, onAdd, onChange, onRemove, variables }: HeadersEditorProps) {
  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Headers
          </span>
          <span className="text-[10px] font-mono text-muted-foreground/70">{headers.length}</span>
        </div>
        <button
          type="button"
          onClick={onAdd}
          className="inline-flex items-center gap-1 text-[10.5px] font-medium text-muted-foreground hover:text-foreground transition-colors"
        >
          <Plus className="w-3 h-3" />
          Add
        </button>
      </div>

      {headers.length === 0 ? (
        <p className="text-[11px] text-muted-foreground/70 italic px-1">
          No headers. Default <code className="font-mono">Content-Type: application/json</code> is sent with body.
        </p>
      ) : (
        <div className="space-y-1.5">
          {headers.map((row, idx) => (
            <div key={idx} className="grid grid-cols-[18px_1fr_1fr_28px] gap-2 items-center min-w-0">
              <input
                type="checkbox"
                checked={row.enabled}
                onChange={(e) => onChange(idx, { enabled: e.target.checked })}
                className="h-3.5 w-3.5 accent-primary"
                aria-label="Enable header"
              />
              <VarInput
                value={row.key}
                onChange={(value) => onChange(idx, { key: value })}
                placeholder="Header"
                className="h-7 text-[12px] font-mono"
                variables={variables}
              />
              <VarInput
                value={row.value}
                onChange={(value) => onChange(idx, { value })}
                placeholder="Value"
                className="h-7 text-[12px] font-mono"
                variables={variables}
              />
              <button
                type="button"
                onClick={() => onRemove(idx)}
                aria-label="Remove header"
                className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
