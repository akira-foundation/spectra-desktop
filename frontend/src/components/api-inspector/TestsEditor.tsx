import { useEffect, useMemo, useRef, useState } from 'react'
import { Plus, Trash2, Check, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
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
}

export function TestsEditor({ projectId, method, path, results }: TestsEditorProps) {
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
      setTests(rows)
      setSavedTests(rows)
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
          {tests.map((t, idx) => (
            <TestRow
              key={idx}
              test={t}
              result={resultByIndex.get(idx)}
              onChange={(patch) => update(idx, patch)}
              onRemove={() => removeTest(idx)}
            />
          ))}
        </ul>
      )}
    </div>
  )
}

interface TestRowProps {
  test: EndpointTest
  result?: TestResult
  onChange: (patch: Partial<EndpointTest>) => void
  onRemove: () => void
}

function TestRow({ test, result, onChange, onRemove }: TestRowProps) {
  const ops = OP_BY_KIND[test.kind] ?? []
  const showPath = test.kind === 'header' || test.kind === 'jsonpath'
  const showOp = ops.length > 0
  const hideExpected =
    test.kind === 'jsonpath' && (test.op === 'exists' || test.op === 'not_exists')

  return (
    <li className="flex items-start gap-1.5">
      <ResultBadge result={result} />
      <div className="flex-1 grid grid-cols-[110px_1fr_110px_1fr_28px] gap-1.5 items-center min-w-0">
        <select
          value={test.kind}
          onChange={(e) => onChange({ kind: e.target.value, op: '', expected: defaultExpectedFor(e.target.value) })}
          className="h-7 text-[11.5px] bg-input/40 border border-border/50 rounded-md px-2"
        >
          {KIND_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        {showPath ? (
          <Input
            value={test.jsonPath ?? ''}
            onChange={(e) => onChange({ jsonPath: e.target.value })}
            placeholder={test.kind === 'header' ? 'Content-Type' : '$.data.token'}
            className="h-7 text-[11.5px] font-mono"
          />
        ) : (
          <span className="text-[11px] text-muted-foreground/70 px-2">—</span>
        )}
        {showOp ? (
          <select
            value={test.op ?? ''}
            onChange={(e) => onChange({ op: e.target.value })}
            className="h-7 text-[11.5px] bg-input/40 border border-border/50 rounded-md px-2"
          >
            <option value="">op…</option>
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
    </li>
  )
}

function ResultBadge({ result }: { result?: TestResult }) {
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
