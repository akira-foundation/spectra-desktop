import { Plus, X } from 'lucide-react'
import type { ReactNode } from 'react'
import { Input } from '@/components/ui/input'
import type { QueryParam } from '@/lib/route-params'
import { VarInput } from './VarInput'

interface ParamsEditorProps {
  routeParams: string[]
  routeValues: string[]
  onRouteValueChange: (index: number, value: string) => void
  queryParams: QueryParam[]
  onQueryAdd: () => void
  onQueryChange: (index: number, patch: Partial<QueryParam>) => void
  onQueryRemove: (index: number) => void
  variables?: Record<string, string>
}

export function ParamsEditor({
  routeParams,
  routeValues,
  onRouteValueChange,
  queryParams,
  onQueryAdd,
  onQueryChange,
  onQueryRemove,
  variables,
}: ParamsEditorProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-3 min-w-0">
      <Section title="Route Params" count={routeParams.length}>
        {routeParams.length === 0 ? (
          <EmptyHint label="No route params" />
        ) : (
          <div className="space-y-1.5">
            {routeParams.map((name, idx) => (
              <RowGrid key={`${name}-${idx}`}>
                <KeyCell value={name} />
                <VarInput
                  value={routeValues[idx] ?? ''}
                  onChange={(value) => onRouteValueChange(idx, value)}
                  placeholder={`Enter ${name}`}
                  className="h-7 text-[12px]"
                  variables={variables}
                />
              </RowGrid>
            ))}
          </div>
        )}
      </Section>

      <Section
        title="Query Params"
        count={queryParams.length}
        action={
          <button
            type="button"
            onClick={onQueryAdd}
            className="inline-flex items-center gap-1 text-[10.5px] font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            <Plus className="w-3 h-3" />
            Add
          </button>
        }
      >
        {queryParams.length === 0 ? (
          <EmptyHint label="No query params" />
        ) : (
          <div className="space-y-1.5">
            {queryParams.map((row, idx) => (
              <RowGrid key={idx}>
                <VarInput
                  value={row.key}
                  onChange={(value) => onQueryChange(idx, { key: value })}
                  placeholder="key"
                  className="h-7 text-[12px] font-mono"
                  variables={variables}
                />
                <div className="flex items-center gap-1">
                  <VarInput
                    value={row.value}
                    onChange={(value) => onQueryChange(idx, { value })}
                    placeholder="value"
                    className="h-7 text-[12px] font-mono"
                    variables={variables}
                  />
                  <button
                    type="button"
                    onClick={() => onQueryRemove(idx)}
                    aria-label="Remove query param"
                    className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
              </RowGrid>
            ))}
          </div>
        )}
      </Section>
    </div>
  )
}

interface SectionProps {
  title: string
  count: number
  children: ReactNode
  action?: ReactNode
}

function Section({ title, count, children, action }: SectionProps) {
  return (
    <div className="min-w-0 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            {title}
          </span>
          <span className="text-[10px] font-mono text-muted-foreground/70">{count}</span>
        </div>
        {action}
      </div>
      {children}
    </div>
  )
}

function RowGrid({ children }: { children: ReactNode }) {
  return <div className="grid grid-cols-[110px_1fr] gap-2 items-center min-w-0">{children}</div>
}

function KeyCell({ value }: { value: string }) {
  return (
    <span className="inline-flex h-7 items-center px-2 rounded-md border border-border/60 bg-muted/40 font-mono text-[11px] text-foreground/85 truncate">
      {value}
    </span>
  )
}

function EmptyHint({ label }: { label: string }) {
  return <p className="text-[11px] text-muted-foreground/70 italic px-1">{label}</p>
}
