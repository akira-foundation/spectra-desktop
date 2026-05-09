import { useEffect, useMemo, useRef, useState } from 'react'
import { Plus, Trash2, Check } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { PathAutocomplete } from './PathAutocomplete'
import { capturesService, type EndpointCapture, type CapturedValue } from '@/services/capturesService'
import { cn } from '@/lib/utils'

interface CapturesEditorProps {
  projectId: string | null
  method: string | null
  path: string | null
  responseBody?: string
  responseHeaders?: Record<string, string[]>
  capturedValues?: CapturedValue[]
  onCapturedChange?: (values: CapturedValue[]) => void
}

export function CapturesEditor({
  projectId,
  method,
  path,
  responseBody,
  responseHeaders,
  capturedValues,
  onCapturedChange,
}: CapturesEditorProps) {
  const endpointKey = method && path ? `${method.toUpperCase()} ${path}` : null
  const [captures, setCaptures] = useState<EndpointCapture[]>([])
  const [savedCaptures, setSavedCaptures] = useState<EndpointCapture[]>([])
  const [saving, setSaving] = useState(false)
  const loadKey = useRef<string | null>(null)

  useEffect(() => {
    if (!projectId || !endpointKey) {
      setCaptures([])
      setSavedCaptures([])
      return
    }
    if (loadKey.current === endpointKey) return
    loadKey.current = endpointKey
    void capturesService.list(projectId, endpointKey).then((rows) => {
      const normalized = rows.map((r) => ({ ...r, source: r.source || 'body' }))
      setCaptures(normalized)
      setSavedCaptures(normalized)
    })
  }, [projectId, endpointKey])

  const dirty = useMemo(
    () => JSON.stringify(captures) !== JSON.stringify(savedCaptures),
    [captures, savedCaptures],
  )

  const update = (idx: number, patch: Partial<EndpointCapture>) => {
    setCaptures((prev) => prev.map((c, i) => (i === idx ? { ...c, ...patch } : c)))
  }
  const add = () => {
    setCaptures((prev) => [...prev, { name: '', source: 'body', path: '' } as EndpointCapture])
  }
  const remove = (idx: number) => {
    setCaptures((prev) => prev.filter((_, i) => i !== idx))
  }

  const save = async () => {
    if (!projectId || !endpointKey) return
    setSaving(true)
    try {
      await capturesService.save({ projectID: projectId, endpointKey, captures })
      setSavedCaptures(captures)
      try {
        const vals = await capturesService.listValues(projectId)
        onCapturedChange?.(vals)
      } catch {}
    } finally {
      setSaving(false)
    }
  }

  const jsonPaths = useMemo(() => extractJSONPaths(responseBody), [responseBody])
  const headerNames = useMemo(() => Object.keys(responseHeaders ?? {}), [responseHeaders])

  const valueByName = useMemo(() => {
    const m = new Map<string, CapturedValue>()
    for (const v of capturedValues ?? []) m.set(v.name, v)
    return m
  }, [capturedValues])

  if (!projectId || !endpointKey) {
    return <p className="text-[11.5px] italic text-muted-foreground p-2">Select an endpoint.</p>
  }

  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-center justify-between">
        <p className="text-[10.5px] text-muted-foreground">
          Extract values into <code className="font-mono">{'{{name}}'}</code> after Execute.
        </p>
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={add}
            className="inline-flex items-center gap-1 text-[10.5px] font-medium text-muted-foreground hover:text-foreground"
          >
            <Plus className="w-3 h-3" />
            Add
          </button>
          <Button
            size="sm"
            variant="outline"
            className="h-6 px-2 text-[10.5px]"
            onClick={save}
            disabled={!dirty || saving}
          >
            {saving ? 'Saving…' : dirty ? 'Save' : 'Saved'}
          </Button>
        </div>
      </div>

      {captures.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground p-2">No captures yet.</p>
      ) : (
        <ul className="space-y-1.5">
          {captures.map((c, idx) => {
            const captured = c.name ? valueByName.get(c.name) : undefined
            return (
              <li key={idx} className="grid grid-cols-[minmax(0,1fr)_90px_minmax(0,1.6fr)_18px_28px] gap-1.5 items-center min-w-0">
                <Input
                  value={c.name ?? ''}
                  onChange={(e) => update(idx, { name: e.target.value })}
                  placeholder="var_name"
                  className="h-7 text-[11.5px] font-mono"
                />
                <select
                  value={c.source || 'body'}
                  onChange={(e) => update(idx, { source: e.target.value })}
                  className="h-7 text-[11.5px] bg-input/40 border border-border/50 rounded-md px-2"
                >
                  <option value="body">body</option>
                  <option value="header">header</option>
                </select>
                <PathAutocomplete
                  value={c.path ?? ''}
                  onChange={(v) => update(idx, { path: v })}
                  placeholder={c.source === 'header' ? 'Content-Type' : '$.data.token'}
                  suggestions={c.source === 'header' ? headerNames : jsonPaths}
                  className="h-7 text-[11.5px]"
                />
                <span
                  className={cn(
                    'inline-flex h-7 w-[18px] items-center justify-center',
                    captured ? 'text-emerald-500' : 'text-transparent',
                  )}
                  title={captured ? `Captured: ${truncate(captured.value, 80)}` : ''}
                >
                  <Check className="w-3 h-3" />
                </span>
                <button
                  type="button"
                  onClick={() => remove(idx)}
                  aria-label="Remove"
                  className="inline-flex h-7 w-7 items-center justify-center text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

function truncate(s: string, n: number) {
  if (s.length <= n) return s
  return s.slice(0, n) + '…'
}

function extractJSONPaths(raw?: string): string[] {
  if (!raw) return []
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    return []
  }
  const out: string[] = []
  const walk = (node: unknown, prefix: string) => {
    if (out.length > 200) return
    if (node && typeof node === 'object' && !Array.isArray(node)) {
      for (const key of Object.keys(node as Record<string, unknown>)) {
        const path = prefix ? `${prefix}.${key}` : `$.${key}`
        out.push(path)
        walk((node as Record<string, unknown>)[key], path)
      }
    } else if (Array.isArray(node) && node.length > 0) {
      const path = `${prefix}[0]`
      out.push(path)
      walk(node[0], path)
    }
  }
  walk(parsed, '')
  return out
}
