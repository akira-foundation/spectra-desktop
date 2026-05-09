import { useMemo, useState } from 'react'
import { JsonEditor } from '../JsonEditor'
import { TreeView } from './TreeView'
import { TableView } from './TableView'
import { RawView } from './RawView'
import { extractTableRows } from '@/lib/format'
import { cn } from '@/lib/utils'

type Mode = 'json' | 'tree' | 'table' | 'raw'

export function ResponseBodyView({ raw }: { raw: string }) {
  const [mode, setMode] = useState<Mode>('json')
  const parsed = useMemo(() => {
    try {
      return JSON.parse(raw)
    } catch {
      return null
    }
  }, [raw])

  const tableRows = useMemo(() => extractTableRows(parsed), [parsed])
  const isJson = parsed !== null

  return (
    <div className="h-full min-h-0 flex flex-col gap-2">
      <div className="flex items-center gap-1 shrink-0">
        {(['json', 'tree', 'table', 'raw'] as Mode[])
          .filter((m) => {
            if (m === 'json' || m === 'tree') return isJson
            if (m === 'table') return tableRows != null
            return true
          })
          .map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setMode(m)}
              className={cn(
                'h-6 px-2 text-[10.5px] rounded transition-colors',
                mode === m ? 'bg-primary/15 text-primary' : 'text-muted-foreground hover:bg-accent/40',
              )}
            >
              {m.toUpperCase()}
            </button>
          ))}
      </div>
      <div className="flex-1 min-h-0 overflow-auto">
        {mode === 'json' &&
          (isJson ? <JsonEditor value={raw} onChange={() => undefined} readOnly /> : <RawView raw={raw} />)}
        {mode === 'tree' && parsed !== null && <TreeView value={parsed} />}
        {mode === 'table' && tableRows && <TableView rows={tableRows} />}
        {mode === 'raw' && <RawView raw={raw} />}
      </div>
    </div>
  )
}
