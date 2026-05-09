import { useState } from 'react'
import { Play, Square, Copy, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { MockStatus } from '@/services/mockService'
import toast from 'react-hot-toast'

interface Props {
  projectId: string
  status: MockStatus
  onStart: (port: number) => Promise<void>
  onStop: () => Promise<void>
  onClearLogs: () => void
}

export function MockHeader({ projectId, status, onStart, onStop, onClearLogs }: Props) {
  const [port, setPort] = useState('4001')
  const [busy, setBusy] = useState(false)

  const start = async () => {
    if (!projectId) return
    setBusy(true)
    try {
      await onStart(parseInt(port, 10) || 0)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }
  const stop = async () => {
    setBusy(true)
    try {
      await onStop()
    } finally {
      setBusy(false)
    }
  }
  const copyURL = async () => {
    if (!status.url) return
    try {
      await navigator.clipboard.writeText(status.url)
      toast.success('URL copied')
    } catch {}
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      {status.running ? (
        <>
          <div className="inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[11px]">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
            Running
          </div>
          <code className="font-mono text-[12px] text-foreground/85">{status.url}</code>
          <Button size="icon-sm" variant="ghost" onClick={copyURL} title="Copy URL">
            <Copy className="h-3.5 w-3.5" />
          </Button>
          <span className="text-[11px] text-muted-foreground">
            {status.requestCount} req
          </span>
          <Button
            size="sm"
            variant="ghost"
            onClick={onClearLogs}
            className="ml-2"
            title="Clear log"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Clear log
          </Button>
          <div className="ml-auto">
            <Button size="sm" variant="outline" onClick={stop} disabled={busy}>
              <Square className="h-3.5 w-3.5" />
              Stop
            </Button>
          </div>
        </>
      ) : (
        <>
          <Input
            value={port}
            onChange={(e) => setPort(e.target.value.replace(/[^0-9]/g, ''))}
            className="h-7 w-24 text-[12px] font-mono"
            placeholder="Port"
          />
          <span className="text-[11px] text-muted-foreground">port (0 = auto)</span>
          <div className="ml-auto">
            <Button size="sm" onClick={start} disabled={busy || !projectId}>
              <Play className="h-3.5 w-3.5" />
              Start mock server
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
