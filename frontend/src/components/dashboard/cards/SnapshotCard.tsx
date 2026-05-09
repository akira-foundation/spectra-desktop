import { GitCompare, Plus, Minus, Pencil } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/format'
import { Card } from './Card'

interface Snapshot {
  id: string
  scannedAt: string | Date
  endpointCount: number
  added: number
  removed: number
  changed: number
}

export function SnapshotCard({ snapshot, onOpen }: { snapshot: Snapshot | undefined; onOpen: () => void }) {
  return (
    <Card
      title="Latest snapshot"
      icon={GitCompare}
      action={snapshot ? { label: 'View all', onClick: onOpen } : undefined}
    >
      {snapshot ? (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[12px] font-medium">{formatDate(new Date(snapshot.scannedAt))}</span>
            <span className="text-[10.5px] font-mono text-muted-foreground">
              {snapshot.endpointCount} endpoints
            </span>
          </div>
          <div className="flex items-center gap-3 text-[11px] font-mono">
            <Delta icon={Plus} count={snapshot.added} tone="emerald" />
            <Delta icon={Minus} count={snapshot.removed} tone="rose" />
            <Delta icon={Pencil} count={snapshot.changed} tone="amber" />
          </div>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">No snapshots yet. Run Sync to capture one.</p>
      )}
    </Card>
  )
}

function Delta({ icon: Icon, count, tone }: { icon: LucideIcon; count: number; tone: 'emerald' | 'rose' | 'amber' }) {
  const toneClass = {
    emerald: 'text-emerald-500',
    rose: 'text-rose-500',
    amber: 'text-amber-500',
  }[tone]
  return (
    <span className={cn('inline-flex items-center gap-1', count === 0 ? 'text-muted-foreground/60' : toneClass)}>
      <Icon className="w-3 h-3" />
      {count}
    </span>
  )
}
