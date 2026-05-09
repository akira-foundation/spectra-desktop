import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  text: string
  title?: string
  className?: string
  size?: 'sm' | 'md'
}

export function CopyButton({ text, title = 'Copy', className, size = 'md' }: Props) {
  const [copied, setCopied] = useState(false)
  const handle = async () => {
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  const dim = size === 'sm' ? 'h-5 w-5' : 'h-6 w-6'
  const icon = size === 'sm' ? 'w-3 h-3' : 'w-3 h-3'
  return (
    <button
      type="button"
      onClick={handle}
      disabled={!text}
      className={cn(
        'inline-flex items-center justify-center rounded transition-colors',
        dim,
        copied
          ? 'text-emerald-500 bg-emerald-500/10'
          : 'text-muted-foreground hover:text-foreground hover:bg-accent/60',
        'disabled:opacity-40 disabled:hover:bg-transparent',
        className,
      )}
      title={copied ? 'Copied!' : title}
    >
      {copied ? <Check className={icon} /> : <Copy className={icon} />}
    </button>
  )
}
