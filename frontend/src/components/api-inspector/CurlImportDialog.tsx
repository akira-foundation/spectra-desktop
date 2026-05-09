import { useState } from 'react'
import { Terminal, Loader2 } from 'lucide-react'
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
  onImport: (parsed: ParsedCurl) => void
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export interface ParsedCurl {
  method: string
  url: string
  baseURL?: string
  path?: string
  headers: Record<string, string>
  body?: string
  query?: Record<string, string>
}

export function CurlImportDialog({ onImport, open: openProp, onOpenChange }: Props) {
  const [internalOpen, setInternalOpen] = useState(false)
  const open = openProp ?? internalOpen
  const setOpen = onOpenChange ?? setInternalOpen
  const controlled = openProp !== undefined
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleImport = async () => {
    if (!text.trim()) return
    setBusy(true)
    setError(null)
    try {
      const { ImportCurl } = await import('../../../wailsjs/go/app/App')
      const parsed = await ImportCurl(text)
      if (!parsed) {
        setError('Could not parse the curl command.')
        return
      }
      onImport(parsed as ParsedCurl)
      setOpen(false)
      setText('')
    } catch (err) {
      setError((err as Error).message ?? 'Failed to parse curl.')
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
            className="inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-border/50 bg-card text-[11px] text-muted-foreground hover:text-foreground hover:bg-accent/60 transition-colors"
            title="Import from curl"
          >
            <Terminal className="w-3 h-3" />
            <span>Import curl</span>
          </button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-2xl max-h-[80vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
          <DialogTitle className="text-base">Import from curl</DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Paste a curl command (e.g. from Chrome DevTools → Network → Copy as curl).
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 min-h-0 overflow-auto px-6 py-4 space-y-3">
          <textarea
            autoFocus
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder={`curl 'https://api.example.com/v1/users' \\\n  -H 'Authorization: Bearer ...' \\\n  --data-raw '{"name":"alice"}'`}
            spellCheck={false}
            rows={10}
            className="w-full bg-input/30 border border-border/50 rounded-md p-3 text-[11.5px] font-mono outline-none focus:border-border resize-y"
          />
          {error && (
            <p className="text-[11px] text-rose-500/90 font-mono">{error}</p>
          )}
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
