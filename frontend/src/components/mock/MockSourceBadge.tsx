import { cn } from '@/lib/utils'
import type { MockSource } from '@/services/mockService'

const TONES: Record<MockSource, string> = {
  auto: 'bg-muted text-muted-foreground border-border/50',
  history: 'bg-emerald-500/15 text-emerald-500 border-emerald-500/30',
  custom: 'bg-purple-500/15 text-purple-500 border-purple-500/30',
  generated: 'bg-blue-500/15 text-blue-500 border-blue-500/30',
  'no-match': 'bg-amber-500/15 text-amber-500 border-amber-500/30',
}

const LABELS: Record<MockSource, string> = {
  auto: 'Auto',
  history: 'History',
  custom: 'Custom',
  generated: 'Generated',
  'no-match': 'No match',
}

export function MockSourceBadge({ source, className }: { source: MockSource; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded border px-1.5 py-px text-[10px] font-medium uppercase tracking-wider',
        TONES[source] ?? TONES.auto,
        className,
      )}
    >
      {LABELS[source] ?? source}
    </span>
  )
}
