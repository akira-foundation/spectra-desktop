import { useEffect, useRef, useState } from 'react'
import { Globe, Check, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useProjectStore } from '@/store/projectStore'

export function BaseURLBar() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const projects = useProjectStore((s) => s.projects)
  const updateBaseURL = useProjectStore((s) => s.updateBaseURL)
  const project = projects.find((p) => p.id === activeProjectId)

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
    </div>
  )
}
