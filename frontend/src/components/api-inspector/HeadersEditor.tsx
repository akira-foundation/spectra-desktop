import { useState } from 'react'
import { Plus, X, Lock, KeyRound } from 'lucide-react'
import { VarInput } from './VarInput'
import { cn } from '@/lib/utils'

const COMMON_HEADERS = [
  'Accept',
  'Accept-Encoding',
  'Accept-Language',
  'Authorization',
  'Cache-Control',
  'Content-Type',
  'Content-Length',
  'Cookie',
  'Host',
  'Origin',
  'Referer',
  'User-Agent',
  'X-Requested-With',
  'X-CSRF-Token',
  'X-Forwarded-For',
  'X-API-Key',
  'If-Match',
  'If-None-Match',
  'If-Modified-Since',
  'Range',
]

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
  autoAuth?: { scheme?: string; tokenPreview?: string } | null
  onOpenAuth?: () => void
}

const isAuthHeader = (k: string) => k.trim().toLowerCase() === 'authorization'

export function HeadersEditor({ headers, onAdd, onChange, onRemove, variables, autoAuth, onOpenAuth }: HeadersEditorProps) {
  const [overrideMode, setOverrideMode] = useState(false)
  const hasAuto = true
  const showAutoBadge = hasAuto && !overrideMode
  const authIndices = headers.map((h, i) => (isAuthHeader(h.key) ? i : -1)).filter((i) => i !== -1)
  const visibleHeaders = headers
    .map((h, i) => ({ row: h, idx: i }))
    .filter(({ row }) => !hasAuto || overrideMode || !isAuthHeader(row.key))

  const enableOverride = () => {
    setOverrideMode(true)
    if (authIndices.length === 0) {
      onAdd()
      setTimeout(() => {
        onChange(headers.length, { key: 'Authorization', value: '', enabled: true })
      }, 0)
    }
  }

  const cancelOverride = () => {
    setOverrideMode(false)
    for (const i of [...authIndices].reverse()) onRemove(i)
  }

  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Headers
          </span>
          <span className="text-[10px] font-mono text-muted-foreground/70">{headers.length}</span>
        </div>
        <div className="flex items-center gap-3">
          {hasAuto && overrideMode && (
            <button
              type="button"
              onClick={cancelOverride}
              className="inline-flex items-center gap-1 text-[10px] font-medium text-emerald-500 hover:text-emerald-400"
              title="Use auto-injected token"
            >
              <Lock className="w-3 h-3" />
              Use auto
            </button>
          )}
          <button
            type="button"
            onClick={onAdd}
            className="inline-flex items-center gap-1 text-[10.5px] font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            <Plus className="w-3 h-3" />
            Add
          </button>
        </div>
      </div>

      {showAutoBadge && (
        <div className="flex items-center justify-between gap-2 px-2 py-1.5 rounded-md border border-border/40 bg-muted/30">
          <button
            type="button"
            onClick={onOpenAuth}
            className="flex items-center gap-1.5 min-w-0 text-left hover:opacity-80"
            title="Manage authentication"
          >
            <Lock className={cn(
              'w-3 h-3 shrink-0',
              autoAuth?.tokenPreview ? 'text-emerald-500' : 'text-muted-foreground',
            )} />
            <span className="text-[10.5px] font-medium text-foreground/80 shrink-0">Authorization</span>
            {autoAuth?.tokenPreview ? (
              <>
                <span className="text-[10.5px] font-mono text-muted-foreground truncate">
                  {autoAuth?.scheme || 'Bearer'} {autoAuth.tokenPreview}
                </span>
                <span className="text-[9.5px] uppercase tracking-wider text-emerald-500/80 shrink-0">auto</span>
              </>
            ) : (
              <span className="text-[10.5px] text-muted-foreground italic">Not authenticated · click to login</span>
            )}
          </button>
          <button
            type="button"
            onClick={enableOverride}
            className="inline-flex items-center gap-1 text-[10px] font-medium text-muted-foreground hover:text-foreground shrink-0"
            title="Override with manual value"
          >
            <KeyRound className="w-3 h-3" />
            Override
          </button>
        </div>
      )}

      {visibleHeaders.length === 0 && !showAutoBadge ? (
        <p className="text-[11px] text-muted-foreground/70 italic px-1">
          No headers. Default <code className="font-mono">Content-Type: application/json</code> is sent with body.
        </p>
      ) : visibleHeaders.length === 0 ? null : (
        <div className="space-y-1.5">
          {visibleHeaders.map(({ row, idx }) => (
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
                suggestions={COMMON_HEADERS}
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
