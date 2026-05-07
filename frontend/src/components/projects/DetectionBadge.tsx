import { Check, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

interface DetectionBadgeProps {
  framework: string
  detected: boolean
  confidence?: number
  className?: string
}

export function DetectionBadge({ framework, detected, confidence, className }: DetectionBadgeProps) {
  const Icon = detected ? Check : AlertCircle
  const label = detected ? framework : 'Unknown framework'
  const pct = confidence != null ? Math.round(confidence * 100) : null

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-medium',
        detected
          ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-500'
          : 'border-amber-500/30 bg-amber-500/10 text-amber-500',
        className,
      )}
    >
      <Icon className="w-3 h-3" />
      <span className="capitalize">{label}</span>
      {pct != null && detected && (
        <span className="text-muted-foreground tabular-nums">· {pct}%</span>
      )}
    </span>
  )
}
