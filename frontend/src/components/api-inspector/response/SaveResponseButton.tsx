import { useState } from 'react'
import { Download, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  text: string
  method?: string
  path?: string
}

export function SaveResponseButton({ text, method, path }: Props) {
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const handle = async () => {
    if (!text || saving) return
    setSaving(true)
    try {
      const { SaveResponseToFile } = await import('../../../../wailsjs/go/app/App')
      const result = await SaveResponseToFile(method ?? '', path ?? '', text)
      if (result) {
        setSaved(true)
        setTimeout(() => setSaved(false), 1500)
      }
    } catch {} finally {
      setSaving(false)
    }
  }
  return (
    <button
      type="button"
      onClick={handle}
      disabled={!text || saving}
      className={cn(
        'inline-flex h-6 w-6 items-center justify-center rounded transition-colors',
        saved
          ? 'text-emerald-500 bg-emerald-500/10'
          : 'text-muted-foreground hover:text-foreground hover:bg-accent/60',
        'disabled:opacity-40 disabled:hover:bg-transparent',
      )}
      title={saved ? 'Saved!' : 'Save response to file'}
    >
      {saved ? <Check className="w-3 h-3" /> : <Download className="w-3 h-3" />}
    </button>
  )
}
