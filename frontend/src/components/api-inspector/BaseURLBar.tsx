import { useEffect, useMemo, useRef, useState } from 'react'
import { Globe, Check, X, LogIn, ChevronDown } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import type { ScannedEndpoint } from '@/services/scannerService'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []

export function BaseURLBar() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const projects = useProjectStore((s) => s.projects)
  const updateBaseURL = useProjectStore((s) => s.updateBaseURL)
  const updateLoginEndpoint = useProjectStore((s) => s.updateLoginEndpoint)
  const project = projects.find((p) => p.id === activeProjectId)
  const allEndpoints = useEndpointsStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
  )

  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState(project?.baseUrl ?? '')
  const [busy, setBusy] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    setValue(project?.baseUrl ?? '')
  }, [project?.id, project?.baseUrl])

  if (!project) return null

  const beginEdit = () => {
    setEditing(true)
    setTimeout(() => inputRef.current?.focus(), 0)
  }

  const cancel = () => {
    setValue(project.baseUrl ?? '')
    setEditing(false)
  }

  const save = async () => {
    const trimmed = value.trim()
    if (!trimmed || trimmed === project.baseUrl) {
      setEditing(false)
      return
    }
    setBusy(true)
    try {
      await updateBaseURL(project.id, trimmed)
      setEditing(false)
    } finally {
      setBusy(false)
    }
  }

  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      void save()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      cancel()
    }
  }

  return (
    <div className="h-9 px-3 border-b border-border/50 flex items-center gap-2 bg-transparent">
      <Globe className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
      <span className="text-[10px] uppercase tracking-wider text-muted-foreground shrink-0">
        Base URL
      </span>
      {editing ? (
        <div className="flex items-center gap-1.5 flex-1 min-w-0">
          <Input
            ref={inputRef}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={onKey}
            placeholder="http://localhost:8000"
            className="h-7 text-[12px] font-mono"
            disabled={busy}
          />
          <Button size="sm" variant="outline" className="h-7 px-2" onClick={save} disabled={busy}>
            <Check className="w-3.5 h-3.5" />
          </Button>
          <Button size="sm" variant="ghost" className="h-7 px-2" onClick={cancel} disabled={busy}>
            <X className="w-3.5 h-3.5" />
          </Button>
        </div>
      ) : (
        <button
          type="button"
          onClick={beginEdit}
          className="flex-1 min-w-0 text-left font-mono text-[12px] text-foreground/85 hover:text-foreground truncate"
          title={project.baseUrl || 'Click to set base URL'}
        >
          {project.baseUrl || (
            <span className="text-muted-foreground italic">click to set base URL</span>
          )}
        </button>
      )}

      <LoginEndpointPicker
        projectId={project.id}
        endpoints={allEndpoints}
        loginEndpointId={project.loginEndpointId ?? ''}
        loginTokenPath={project.loginTokenPath ?? ''}
        onSave={updateLoginEndpoint}
      />
    </div>
  )
}

interface LoginEndpointPickerProps {
  projectId: string
  endpoints: Array<{ id: string; method: string; path: string }>
  loginEndpointId: string
  loginTokenPath: string
  onSave: (id: string, endpointId: string, tokenPath: string) => Promise<void>
}

function LoginEndpointPicker({
  projectId,
  endpoints,
  loginEndpointId,
  loginTokenPath,
  onSave,
}: LoginEndpointPickerProps) {
  const { getMethodColor } = useHttpMethod()
  const [open, setOpen] = useState(false)
  const [endpointDraft, setEndpointDraft] = useState(loginEndpointId)
  const [pathDraft, setPathDraft] = useState(loginTokenPath)
  const [saving, setSaving] = useState(false)
  const [query, setQuery] = useState('')

  useEffect(() => {
    setEndpointDraft(loginEndpointId)
    setPathDraft(loginTokenPath)
  }, [loginEndpointId, loginTokenPath, projectId])

  const selected = endpoints.find((e) => e.id === loginEndpointId)
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    const writes = endpoints.filter((e) => {
      const m = e.method.toUpperCase()
      return m === 'POST' || m === 'PUT' || m === 'PATCH'
    })
    if (!q) return writes.slice(0, 80)
    return writes
      .filter((e) => e.path.toLowerCase().includes(q) || e.method.toLowerCase().includes(q))
      .slice(0, 80)
  }, [endpoints, query])

  const handleSave = async () => {
    setSaving(true)
    try {
      await onSave(projectId, endpointDraft, pathDraft.trim())
      setOpen(false)
    } finally {
      setSaving(false)
    }
  }

  const handleClear = async () => {
    setSaving(true)
    try {
      await onSave(projectId, '', '')
      setEndpointDraft('')
      setPathDraft('')
      setOpen(false)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            'shrink-0 h-7 inline-flex items-center gap-1.5 px-2 rounded-md text-[11px] border border-border/50 hover:bg-accent/40 transition-colors',
            loginEndpointId ? 'text-foreground' : 'text-muted-foreground',
          )}
          title="Configure login endpoint"
        >
          <LogIn className="w-3 h-3" />
          {selected ? (
            <span className="font-mono truncate max-w-[180px]">{selected.path}</span>
          ) : (
            <span>Set login route</span>
          )}
          <ChevronDown className="w-3 h-3 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[360px] p-3 space-y-3">
        <div className="space-y-1.5">
          <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Login endpoint
          </label>
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search endpoints..."
            className="h-7 text-[12px]"
          />
          <div className="max-h-56 overflow-auto rounded-md border border-border/40 bg-muted/20">
            <button
              type="button"
              onClick={() => setEndpointDraft('')}
              className={cn(
                'w-full text-left px-2 py-1.5 text-[11.5px] hover:bg-accent/40 flex items-center gap-2',
                endpointDraft === '' && 'bg-accent/60',
              )}
            >
              <span className="italic text-muted-foreground">— None —</span>
            </button>
            {filtered.map((ep) => (
              <button
                key={ep.id}
                type="button"
                onClick={() => setEndpointDraft(ep.id)}
                className={cn(
                  'w-full text-left px-2 py-1.5 text-[11.5px] hover:bg-accent/40 flex items-center gap-2',
                  endpointDraft === ep.id && 'bg-accent/60',
                )}
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(ep.method),
                  )}
                >
                  {ep.method}
                </span>
                <span className="font-mono truncate flex-1">{ep.path}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Token JSONPath
          </label>
          <Input
            value={pathDraft}
            onChange={(e) => setPathDraft(e.target.value)}
            placeholder="data.token (auto-detected if empty)"
            className="h-7 text-[12px] font-mono"
          />
          <p className="text-[10.5px] text-muted-foreground/70">
            Leave empty to use heuristic detection.
          </p>
        </div>

        <div className="flex items-center gap-2 pt-1">
          <Button
            size="sm"
            onClick={handleSave}
            disabled={saving}
            className="flex-1 h-7 text-[11px]"
          >
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={handleClear}
            disabled={saving}
            className="h-7 text-[11px]"
          >
            Clear
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
