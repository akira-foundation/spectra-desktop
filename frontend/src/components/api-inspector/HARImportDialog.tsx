import { useState } from 'react'
import { FileText, Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
  DialogFooter,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface Props {
  onImport: (entries: any[]) => void
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export function HARImportDialog({ onImport, open: openProp, onOpenChange }: Props) {
  const [internalOpen, setInternalOpen] = useState(false)
  const open = openProp ?? internalOpen
  const setOpen = onOpenChange ?? setInternalOpen
  const controlled = openProp !== undefined
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleFile = async (file: File) => {
    const content = await file.text()
    setText(content)
  }

  const handleImport = async () => {
    if (!text.trim()) return
    setBusy(true)
    setError(null)
    try {
      const { ImportHAR } = await import('../../../wailsjs/go/app/App')
      const entries = await ImportHAR(text)
      if (!entries || entries.length === 0) {
        setError('No entries found.')
        return
      }
      onImport(entries)
      setOpen(false)
      setText('')
    } catch (err) {
      setError((err as Error).message ?? 'Failed to parse HAR.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {!controlled && (
        <DialogTrigger asChild>
          <button
            type="button"
            className="inline-flex items-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground"
            title="Import from HAR"
          >
            <FileText className="w-3 h-3" />
            HAR
          </button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-2xl max-h-[80vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
          <DialogTitle className="text-base">Import from HAR</DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Paste HAR JSON or upload a .har file (DevTools → Network → right-click → Save all as HAR).
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 min-h-0 overflow-auto px-6 py-4 space-y-3">
          <input
            type="file"
            accept=".har,application/json"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) void handleFile(f)
            }}
            className="text-[11px]"
          />
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder='{"log":{"entries":[...]}}'
            spellCheck={false}
            rows={10}
            className="w-full bg-input/30 border border-border/50 rounded-md p-3 text-[11.5px] font-mono outline-none focus:border-border resize-y"
          />
          {error && <p className="text-[11px] text-rose-500/90 font-mono">{error}</p>}
        </div>
        <DialogFooter className="px-6 py-3 shrink-0 border-t border-border/40 gap-2">
          <Button variant="ghost" size="sm" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button size="sm" onClick={handleImport} disabled={!text.trim() || busy} className="gap-1.5">
            {busy && <Loader2 className="w-3 h-3 animate-spin" />}
            Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
