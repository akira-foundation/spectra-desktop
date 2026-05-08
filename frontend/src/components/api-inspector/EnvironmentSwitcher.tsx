import { useEffect, useMemo, useState } from 'react'
import {
  Layers,
  Check,
  Plus,
  Pencil,
  Trash2,
  ChevronsUpDown,
  X,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useProjectStore } from '@/store/projectStore'
import { useEnvironmentStore } from '@/store/environmentStore'
import { useUIStore } from '@/store/uiStore'
import type { EnvironmentDTO } from '@/services/environmentService'
import { cn } from '@/lib/utils'

const EMPTY_ENVS: EnvironmentDTO[] = []

export function EnvironmentSwitcher() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const project = useProjectStore((s) =>
    s.projects.find((p) => p.id === s.activeProjectId),
  )
  const setActiveEnv = useProjectStore((s) => s.setActiveEnvironment)

  const envs = useEnvironmentStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_ENVS : EMPTY_ENVS,
  )
  const loadEnvs = useEnvironmentStore((s) => s.load)
  const removeEnv = useEnvironmentStore((s) => s.remove)
  const setActiveBackend = useEnvironmentStore((s) => s.setActive)

  useEffect(() => {
    if (activeProjectId) void loadEnvs(activeProjectId)
  }, [activeProjectId, loadEnvs])

  const [pickerOpen, setPickerOpen] = useState(false)
  const [editing, setEditing] = useState<EnvironmentDTO | null>(null)
  const [creating, setCreating] = useState(false)

  const editingEnvId = useUIStore((s) => s.editingEnvId)
  const setEditingEnvId = useUIStore((s) => s.setEditingEnvId)
  useEffect(() => {
    if (!editingEnvId) return
    const env = envs.find((e) => e.id === editingEnvId)
    if (env) {
      setEditing(env)
      setEditingEnvId(null)
    }
  }, [editingEnvId, envs, setEditingEnvId])

  const active = useMemo(
    () => envs.find((e) => e.id === project?.activeEnvironmentId) ?? null,
    [envs, project?.activeEnvironmentId],
  )

  if (!activeProjectId) return null

  const handlePick = async (envId: string) => {
    if (!activeProjectId) return
    await setActiveBackend(activeProjectId, envId)
    setActiveEnv(activeProjectId, envId)
    setPickerOpen(false)
  }

  const handleClearActive = async () => {
    if (!activeProjectId) return
    await setActiveBackend(activeProjectId, '')
    setActiveEnv(activeProjectId, '')
    setPickerOpen(false)
  }

  return (
    <>
      <Popover open={pickerOpen} onOpenChange={setPickerOpen}>
        <PopoverTrigger asChild>
          <button
            type="button"
            className={cn(
              'h-7 inline-flex items-center gap-1.5 px-2 rounded-md text-[11px] border border-border/50 hover:bg-accent/40 transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground',
            )}
            title="Switch environment"
          >
            <Layers className="w-3 h-3" />
            <span className="truncate max-w-[120px]">
              {active ? active.name : 'No environment'}
            </span>
            <ChevronsUpDown className="w-3 h-3 opacity-60" />
          </button>
        </PopoverTrigger>
        <PopoverContent align="end" className="w-[240px] p-1">
          <div className="space-y-px">
            <button
              type="button"
              onClick={handleClearActive}
              className={cn(
                'w-full flex items-center gap-2 px-2 py-1.5 rounded text-[11.5px] hover:bg-accent/40',
                !active && 'bg-accent/60',
              )}
            >
              <span className="italic text-muted-foreground flex-1 text-left">— None —</span>
              {!active && <Check className="w-3 h-3 text-primary" />}
            </button>
            {envs.map((env) => (
              <div
                key={env.id}
                className={cn(
                  'group flex items-center gap-1 rounded',
                  env.id === active?.id && 'bg-accent/60',
                )}
              >
                <button
                  type="button"
                  onClick={() => handlePick(env.id)}
                  className="flex-1 flex items-center gap-2 px-2 py-1.5 text-[11.5px] hover:bg-accent/40 rounded text-left"
                >
                  <span className="truncate flex-1">{env.name}</span>
                  <span className="text-[10px] text-muted-foreground/70 font-mono">
                    {Object.keys(env.vars ?? {}).length}
                  </span>
                  {env.id === active?.id && <Check className="w-3 h-3 text-primary" />}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setEditing(env)
                    setPickerOpen(false)
                  }}
                  aria-label="Edit"
                  className="opacity-0 group-hover:opacity-100 inline-flex h-6 w-6 items-center justify-center text-muted-foreground hover:text-foreground"
                >
                  <Pencil className="w-3 h-3" />
                </button>
              </div>
            ))}
          </div>
          <div className="border-t border-border/40 mt-1 pt-1">
            <button
              type="button"
              onClick={() => {
                setCreating(true)
                setPickerOpen(false)
              }}
              className="w-full flex items-center gap-2 px-2 py-1.5 text-[11.5px] text-muted-foreground hover:text-foreground hover:bg-accent/40 rounded"
            >
              <Plus className="w-3 h-3" />
              New environment
            </button>
          </div>
        </PopoverContent>
      </Popover>

      <EnvironmentEditor
        open={creating}
        projectId={activeProjectId}
        env={null}
        onClose={() => setCreating(false)}
        onSaved={async () => {
          setCreating(false)
        }}
      />
      <EnvironmentEditor
        open={Boolean(editing)}
        projectId={activeProjectId}
        env={editing}
        onClose={() => setEditing(null)}
        onSaved={() => setEditing(null)}
        onDelete={async () => {
          if (editing && activeProjectId) {
            await removeEnv(activeProjectId, editing.id)
            if (project?.activeEnvironmentId === editing.id) {
              setActiveEnv(activeProjectId, '')
            }
            setEditing(null)
          }
        }}
      />

    </>
  )
}

function applyRename(input: string, from: string, to: string): string {
  const escaped = from.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const re = new RegExp(`\\{\\{\\s*${escaped}\\s*\\}\\}`, 'g')
  return input.replace(re, `{{${to}}}`)
}

async function applyRenames(
  projectId: string,
  renames: Array<{ from: string; to: string }>,
  baseUrl: string,
  updateBaseURL: (id: string, url: string) => Promise<void>,
) {
  const { useUIStore } = await import('@/store/uiStore')
  const ui = useUIStore.getState()

  let nextBase = baseUrl
  for (const r of renames) {
    nextBase = applyRename(nextBase, r.from, r.to)
  }
  if (nextBase !== baseUrl) {
    try {
      await updateBaseURL(projectId, nextBase)
    } catch (err) {
      console.error('rename baseUrl failed:', err)
    }
  }

  const bodies = { ...ui.requestBodyByEndpoint }
  let bodiesChanged = false
  for (const [k, body] of Object.entries(bodies)) {
    if (!k.startsWith(projectId + '#')) continue
    let next = body
    for (const r of renames) {
      next = applyRename(next, r.from, r.to)
    }
    if (next !== body) {
      bodies[k] = next
      bodiesChanged = true
    }
  }

  const headers = { ...ui.requestHeadersByEndpoint }
  let headersChanged = false
  for (const [k, list] of Object.entries(headers)) {
    if (!k.startsWith(projectId + '#')) continue
    const nextList = list.map((h) => {
      let key = h.key
      let value = h.value
      for (const r of renames) {
        key = applyRename(key, r.from, r.to)
        value = applyRename(value, r.from, r.to)
      }
      return { ...h, key, value }
    })
    if (nextList.some((h, i) => h.key !== list[i].key || h.value !== list[i].value)) {
      headers[k] = nextList
      headersChanged = true
    }
  }

  if (bodiesChanged || headersChanged) {
    useUIStore.setState({
      requestBodyByEndpoint: bodies,
      requestHeadersByEndpoint: headers,
    })
  }
}

interface EnvironmentEditorProps {
  open: boolean
  projectId: string
  env: EnvironmentDTO | null
  onClose: () => void
  onSaved: () => void | Promise<void>
  onDelete?: () => void | Promise<void>
}

function EnvironmentEditor({ open, projectId, env, onClose, onSaved, onDelete }: EnvironmentEditorProps) {
  const saveEnv = useEnvironmentStore((s) => s.save)
  const updateBaseURL = useProjectStore((s) => s.updateBaseURL)
  const project = useProjectStore((s) => s.projects.find((p) => p.id === projectId))
  const [name, setName] = useState(env?.name ?? '')
  const [rows, setRows] = useState<Array<{ key: string; value: string; originalKey?: string }>>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setName(env?.name ?? '')
    const entries = Object.entries(env?.vars ?? {})
    setRows(
      entries.length > 0
        ? entries.map(([key, value]) => ({ key, value, originalKey: key }))
        : [{ key: '', value: '' }],
    )
  }, [open, env])

  const updateRow = (idx: number, patch: Partial<{ key: string; value: string }>) => {
    setRows((prev) => prev.map((r, i) => (i === idx ? { ...r, ...patch } : r)))
  }
  const addRow = () => setRows((prev) => [...prev, { key: '', value: '' }])
  const removeRow = (idx: number) => setRows((prev) => prev.filter((_, i) => i !== idx))

  const handleSave = async () => {
    const trimmedName = name.trim()
    if (!trimmedName) return
    const vars: Record<string, string> = {}
    const renames: Array<{ from: string; to: string }> = []
    for (const r of rows) {
      const k = r.key.trim()
      if (!k) continue
      vars[k] = r.value
      if (r.originalKey && r.originalKey !== k) {
        renames.push({ from: r.originalKey, to: k })
      }
    }
    setSaving(true)
    try {
      await saveEnv({
        id: env?.id,
        projectID: projectId,
        name: trimmedName,
        vars,
        sortOrder: env?.sortOrder ?? 0,
      })
      if (renames.length > 0) {
        await applyRenames(projectId, renames, project?.baseUrl ?? '', updateBaseURL)
      }
      await onSaved()
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="sm:max-w-md max-h-[80vh] flex flex-col gap-0 p-0">
        <DialogHeader className="px-6 pt-6 pb-3 border-b border-border/40">
          <DialogTitle className="text-base">
            {env ? 'Edit environment' : 'New environment'}
          </DialogTitle>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4 space-y-4">
          <div className="space-y-1.5">
            <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              Name
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="dev / staging / prod"
              className="h-8 text-[12px]"
            />
          </div>

          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                Variables
              </label>
              <button
                type="button"
                onClick={addRow}
                className="inline-flex items-center gap-1 text-[10.5px] font-medium text-muted-foreground hover:text-foreground"
              >
                <Plus className="w-3 h-3" />
                Add
              </button>
            </div>
            <p className="text-[10.5px] text-muted-foreground/70">
              Reference with <code className="font-mono">{'{{key}}'}</code> in path, headers, or body.
            </p>
            <div className="space-y-1.5">
              {rows.map((row, idx) => (
                <div key={idx} className="grid grid-cols-[1fr_1.4fr_28px] gap-2 items-center">
                  <Input
                    value={row.key}
                    onChange={(e) => updateRow(idx, { key: e.target.value })}
                    placeholder="key"
                    className="h-7 text-[12px] font-mono"
                  />
                  <Input
                    value={row.value}
                    onChange={(e) => updateRow(idx, { value: e.target.value })}
                    placeholder="value"
                    className="h-7 text-[12px] font-mono"
                  />
                  <button
                    type="button"
                    onClick={() => removeRow(idx)}
                    aria-label="Remove"
                    className="inline-flex h-7 w-7 items-center justify-center text-muted-foreground hover:text-destructive"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        <DialogFooter className="px-6 py-3 border-t border-border/40 gap-2 justify-between sm:justify-between">
          {env && onDelete && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => void onDelete()}
              className="text-destructive hover:text-destructive"
            >
              <Trash2 className="w-3.5 h-3.5" />
              Delete
            </Button>
          )}
          <div className="flex items-center gap-2 ml-auto">
            <Button variant="outline" size="sm" onClick={onClose} disabled={saving}>
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={handleSave}
              disabled={saving || !name.trim()}
              className="min-w-[100px]"
            >
              {saving ? 'Saving...' : env ? 'Save' : 'Create'}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
