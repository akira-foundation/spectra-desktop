import { useEffect, useMemo, useRef, useState } from 'react'
import { Plus, Trash2, Check, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { PathAutocomplete } from './PathAutocomplete'
import { testsService, type EndpointTest, type TestResult } from '@/services/testsService'
import { cn } from '@/lib/utils'

const KIND_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'status', label: 'Status' },
  { value: 'max_duration', label: 'Max duration' },
  { value: 'header', label: 'Header' },
  { value: 'jsonpath', label: 'JSON path' },
  { value: 'body', label: 'Body' },
]

const OP_BY_KIND: Record<string, Array<{ value: string; label: string }>> = {
  status: [],
  max_duration: [],
  header: [
    { value: 'exists', label: 'exists' },
    { value: 'not_exists', label: 'not exists' },
    { value: 'equals', label: 'equals' },
    { value: 'contains', label: 'contains' },
    { value: 'matches', label: 'matches /regex/' },
  ],
  jsonpath: [
    { value: 'exists', label: 'exists' },
    { value: 'not_exists', label: 'not exists' },
    { value: 'equals', label: 'equals' },
    { value: 'matches', label: 'matches /regex/' },
    { value: 'type', label: 'type' },
    { value: 'min_length', label: 'min length' },
  ],
  body: [
    { value: 'contains', label: 'contains' },
    { value: 'matches', label: 'matches /regex/' },
  ],
}

interface TestsEditorProps {
  projectId: string | null
  method: string | null
  path: string | null
  results?: TestResult[]
  responseBody?: string
  responseHeaders?: Record<string, string[]>
}

export function TestsEditor({
  projectId,
  method,
  path,
  results,
  responseBody,
  responseHeaders,
}: TestsEditorProps) {
  const endpointKey = method && path ? `${method.toUpperCase()} ${path}` : null
  const [tests, setTests] = useState<EndpointTest[]>([])
  const [savedTests, setSavedTests] = useState<EndpointTest[]>([])
  const [saving, setSaving] = useState(false)
  const loadKey = useRef<string | null>(null)

  useEffect(() => {
    if (!projectId || !endpointKey) {
      setTests([])
      setSavedTests([])
      return
    }
    if (loadKey.current === endpointKey) return
    loadKey.current = endpointKey
    void testsService.list(projectId, endpointKey).then((rows) => {
      const normalized = rows.map((r) => {
        const opts = OP_BY_KIND[r.kind] ?? []
        if (opts.length > 0 && !r.op) return { ...r, op: opts[0].value }
        return r
      })
      setTests(normalized)
      setSavedTests(normalized)
    })
  }, [projectId, endpointKey])

  const dirty = useMemo(() => JSON.stringify(tests) !== JSON.stringify(savedTests), [tests, savedTests])

  const update = (idx: number, patch: Partial<EndpointTest>) => {
    setTests((prev) => prev.map((t, i) => (i === idx ? { ...t, ...patch } : t)))
  }
  const addTest = () => {
    setTests((prev) => [...prev, { kind: 'status', expected: '200' } as EndpointTest])
  }
  const removeTest = (idx: number) => {
    setTests((prev) => prev.filter((_, i) => i !== idx))
  }

  const save = async () => {
    if (!projectId || !endpointKey) return
    setSaving(true)
    try {
      await testsService.save({ projectID: projectId, endpointKey, tests })
      setSavedTests(tests)
    } finally {
      setSaving(false)
    }
  }

  const resultByIndex = useMemo(() => {
    if (!results) return new Map<number, TestResult>()
    const m = new Map<number, TestResult>()
    results.forEach((r, i) => m.set(i, r))
    return m
  }, [results])

  const jsonPaths = useMemo(() => extractJSONPaths(responseBody), [responseBody])
  const headerNames = useMemo(() => Object.keys(responseHeaders ?? {}), [responseHeaders])

  if (!projectId || !endpointKey) {
    return <p className="text-[11.5px] italic text-muted-foreground p-2">Select an endpoint.</p>
  }

  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-center justify-between">
        <p className="text-[10.5px] text-muted-foreground">
          Run automatically after each Execute. Results appear inline.
        </p>
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={addTest}
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

      {tests.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground p-2">No tests yet.</p>
      ) : (
        <ul className="space-y-1.5">
          {tests.map((t, idx) => {
            const saved = savedTests[idx]
            const stale = !saved || JSON.stringify(saved) !== JSON.stringify(t)
            return (
              <TestRow
                key={idx}
                test={t}
                result={stale ? undefined : resultByIndex.get(idx)}
                stale={stale && !!resultByIndex.get(idx)}
                onChange={(patch) => update(idx, patch)}
                onRemove={() => removeTest(idx)}
                jsonPaths={jsonPaths}
                headerNames={headerNames}
              />
            )
          })}
        </ul>
      )}
    </div>
  )
}

interface TestRowProps {
  test: EndpointTest
  result?: TestResult
  stale?: boolean
  onChange: (patch: Partial<EndpointTest>) => void
  onRemove: () => void
  jsonPaths: string[]
  headerNames: string[]
}

function TestRow({ test, result, stale, onChange, onRemove, jsonPaths, headerNames }: TestRowProps) {
  const ops = OP_BY_KIND[test.kind] ?? []

  useEffect(() => {
    if (ops.length > 0 && !test.op) {
      onChange({ op: ops[0].value })
    }
  }, [ops, test.op])

  const showPath = test.kind === 'header' || test.kind === 'jsonpath'
  const showOp = ops.length > 0
  const hideExpected =
    test.kind === 'jsonpath' && (test.op === 'exists' || test.op === 'not_exists')
  const datalistId = `path-suggestions-${test.kind}`

  return (
    <li className="flex flex-col gap-0.5">
    <div className="flex items-start gap-1.5">
      <ResultBadge result={result} stale={stale} />
      <div className="flex-1 grid grid-cols-[100px_minmax(0,1.6fr)_110px_minmax(0,1fr)_28px] gap-1.5 items-center min-w-0">
        <select
          value={test.kind}
          onChange={(e) => {
            const k = e.target.value
            const firstOp = OP_BY_KIND[k]?.[0]?.value ?? ''
            onChange({ kind: k, op: firstOp, expected: defaultExpectedFor(k) })
          }}
          className="h-7 text-[11.5px] bg-input/40 border border-border/50 rounded-md px-2"
        >
          {KIND_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        {showPath ? (
          <PathAutocomplete
            value={test.jsonPath ?? ''}
            onChange={(v) => onChange({ jsonPath: v })}
            placeholder={test.kind === 'header' ? 'Content-Type' : '$.data.token'}
            suggestions={test.kind === 'header' ? headerNames : jsonPaths}
            className="h-7 text-[11.5px]"
          />
        ) : (
          <span className="text-[11px] text-muted-foreground/70 px-2">—</span>
        )}
        {showOp ? (
          <select
            value={test.op || ops[0].value}
            onChange={(e) => onChange({ op: e.target.value })}
            className="h-7 text-[11.5px] bg-input/40 border border-border/50 rounded-md px-2"
          >
            {ops.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        ) : (
          <span className="text-[11px] text-muted-foreground/70 px-2">—</span>
        )}
        {!hideExpected ? (
          <Input
            value={test.expected ?? ''}
            onChange={(e) => onChange({ expected: e.target.value })}
            placeholder={placeholderForExpected(test)}
            className="h-7 text-[11.5px] font-mono"
          />
        ) : (
          <span className="text-[11px] text-muted-foreground/70 px-2">—</span>
        )}
        <button
          type="button"
          onClick={onRemove}
          aria-label="Remove"
          className="inline-flex h-7 w-7 items-center justify-center text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>
    </div>
    {result && !stale ? (
      <p className={cn(
        'pl-7 text-[10.5px] font-mono',
        result.pass ? 'text-emerald-500/80' : 'text-rose-500/80',
      )}>
        {result.name}
        {result.message ? ` — ${result.message}` : ''}
      </p>
    ) : null}
    </li>
  )
}

function ResultBadge({ result, stale }: { result?: TestResult; stale?: boolean }) {
  if (stale) {
    return (
      <span
        title="Edited — re-execute to update"
        className="inline-flex h-7 w-5 shrink-0 items-center justify-center text-muted-foreground/60"
      >
        <span className="block w-1.5 h-1.5 rounded-full bg-amber-500/70" />
      </span>
    )
  }
  if (!result) {
    return <span className="inline-flex h-7 w-5 items-center justify-center" />
  }
  return (
    <span
      title={result.message || (result.pass ? 'pass' : 'fail')}
      className={cn(
        'inline-flex h-7 w-5 shrink-0 items-center justify-center',
        result.pass ? 'text-emerald-500' : 'text-rose-500',
      )}
    >
      {result.pass ? <Check className="w-3.5 h-3.5" /> : <X className="w-3.5 h-3.5" />}
    </span>
  )
}

function defaultExpectedFor(kind: string): string {
  switch (kind) {
    case 'status':
      return '200'
    case 'max_duration':
      return '1000'
    default:
      return ''
  }
}

function placeholderForExpected(test: EndpointTest): string {
  switch (test.kind) {
    case 'status':
      return '200 / 2xx'
    case 'max_duration':
      return 'ms'
    case 'jsonpath':
      if (test.op === 'type') return 'string / number / array / object / boolean'
      if (test.op === 'min_length') return '1'
      return 'value'
    default:
      return 'value'
  }
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
