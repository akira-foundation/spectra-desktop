import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { historyService } from '@/services/historyService'
import { cn } from '@/lib/utils'

export function CopyHistoryButton({ entryId }: { entryId: string }) {
  const [copied, setCopied] = useState(false)
  const handle = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      const detail = await historyService.get(entryId)
      const body = detail?.responseBody ?? ''
      if (!body) return
      await navigator.clipboard.writeText(body)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  return (
    <button
      type="button"
      onClick={handle}
      title={copied ? 'Copied!' : 'Copy response'}
      className={cn(
        'inline-flex h-5 w-5 items-center justify-center rounded transition-all',
        copied
          ? 'text-emerald-500 bg-emerald-500/10 opacity-100'
          : 'text-muted-foreground/40 opacity-0 group-hover:opacity-100 hover:text-foreground hover:bg-accent/60',
      )}
    >
      {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
    </button>
  )
}
