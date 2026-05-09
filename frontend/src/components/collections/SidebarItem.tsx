import { useEffect, useState } from 'react'
import { Trash2 } from 'lucide-react'
import { collectionsService, type Collection, type CollectionRun } from '@/services/collectionsService'
import { cn } from '@/lib/utils'

interface Props {
  collection: Collection
  active: boolean
  run?: CollectionRun | null
  onSelect: () => void
  onChanged: () => void
}

export function SidebarItem({ collection: c, active, run, onSelect, onChanged }: Props) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(c.name)

  useEffect(() => setName(c.name), [c.id, c.name])

  useEffect(() => {
    if (!editing) return
    const trimmed = name.trim()
    if (!trimmed || trimmed === c.name) return
    const t = setTimeout(async () => {
      await collectionsService.save({
        id: c.id,
        projectID: c.projectID,
        name: trimmed,
        description: c.description ?? '',
        items: c.items ?? [],
      })
      onChanged()
    }, 500)
    return () => clearTimeout(t)
  }, [name, editing])

  const commit = () => {
    const trimmed = name.trim()
    if (!trimmed) setName(c.name)
    setEditing(false)
  }

  const remove = async () => {
    await collectionsService.remove(c.id)
    onChanged()
  }

  return (
    <li>
      <div
        className={cn(
          'group flex items-center gap-1.5 px-2 py-1.5 text-[11.5px] hover:bg-accent/40',
          active && 'bg-accent/60',
        )}
      >
        {editing ? (
          <input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            onBlur={() => void commit()}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void commit()
              if (e.key === 'Escape') {
                setName(c.name)
                setEditing(false)
              }
            }}
            className="flex-1 min-w-0 h-5 px-1 rounded bg-input/30 border border-border/50 outline-none focus:border-border text-[11.5px]"
          />
        ) : (
          <button
            type="button"
            onClick={onSelect}
            onDoubleClick={() => setEditing(true)}
            className="flex-1 min-w-0 truncate text-left"
            title="Double-click to rename"
          >
            {c.name}
          </button>
        )}
        {run && !editing && (
          <span className="text-[9.5px] font-mono text-muted-foreground tabular-nums shrink-0">
            {run.passCount}/{(run.passCount ?? 0) + (run.failCount ?? 0) + (run.skipCount ?? 0)}
          </span>
        )}
        {!editing && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              void remove()
            }}
            title="Delete collection"
            className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive shrink-0"
          >
            <Trash2 className="w-3 h-3" />
          </button>
        )}
      </div>
    </li>
  )
}
