import { Plus, X, FileUp, Type } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export interface MultipartPart {
  name: string
  value?: string
  filePath?: string
}

interface Props {
  parts: MultipartPart[]
  onChange: (parts: MultipartPart[]) => void
}

export function MultipartEditor({ parts, onChange }: Props) {
  const update = (idx: number, patch: Partial<MultipartPart>) => {
    onChange(parts.map((p, i) => (i === idx ? { ...p, ...patch } : p)))
  }
  const remove = (idx: number) => onChange(parts.filter((_, i) => i !== idx))
  const add = () => onChange([...parts, { name: '', value: '' }])

  const pickFile = async (idx: number) => {
    try {
      const { PickFile } = await import('../../../wailsjs/go/app/App')
      const path = await PickFile()
      if (path) update(idx, { filePath: path, value: undefined })
    } catch {}
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          Multipart parts
        </span>
        <Button size="sm" variant="outline" className="h-6 px-2 text-[10.5px] gap-1" onClick={add}>
          <Plus className="w-3 h-3" />
          Add part
        </Button>
      </div>
      {parts.length === 0 ? (
        <p className="px-1 py-4 text-[11px] italic text-muted-foreground/70 text-center">
          No parts. Click Add to build a multipart/form-data body.
        </p>
      ) : (
        <ul className="m-0 p-0 list-none space-y-1.5">
          {parts.map((p, idx) => {
            const isFile = !!p.filePath
            return (
              <li key={idx} className="flex items-center gap-1.5">
                <Input
                  value={p.name}
                  onChange={(e) => update(idx, { name: e.target.value })}
                  placeholder="field name"
                  className="h-7 text-[11.5px] font-mono w-32 shrink-0"
                />
                <button
                  type="button"
                  onClick={() => {
                    if (isFile) update(idx, { filePath: undefined, value: '' })
                    else void pickFile(idx)
                  }}
                  title={isFile ? 'Switch to text' : 'Switch to file'}
                  className={cn(
                    'inline-flex h-7 items-center justify-center rounded border border-border/50 shrink-0 px-2 gap-1 text-[10px] font-mono uppercase tracking-wider',
                    isFile
                      ? 'bg-primary/15 text-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent/60',
                  )}
                >
                  {isFile ? <FileUp className="w-3 h-3" /> : <Type className="w-3 h-3" />}
                  {isFile ? 'File' : 'Text'}
                </button>
                {isFile ? (
                  <button
                    type="button"
                    onClick={() => void pickFile(idx)}
                    className="flex-1 h-7 px-2 text-left text-[11px] font-mono rounded border border-border/50 bg-input/30 text-foreground/85 truncate hover:bg-accent/40"
                    title={p.filePath}
                  >
                    {p.filePath?.split('/').pop()}
                  </button>
                ) : (
                  <Input
                    value={p.value ?? ''}
                    onChange={(e) => update(idx, { value: e.target.value })}
                    placeholder="value"
                    className="h-7 text-[11.5px] font-mono flex-1"
                  />
                )}
                <button
                  type="button"
                  onClick={() => remove(idx)}
                  aria-label="Remove part"
                  className="inline-flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 shrink-0"
                >
                  <X className="w-3 h-3" />
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
