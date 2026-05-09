import {
  Navigation,
  Code,
  Database,
  Cpu,
  AlertCircle,
  FileCheck,
  Briefcase,
  Mail,
  Wrench,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import type { StatCard } from '@/services/scannerService'

const KIND_ICON: Record<string, LucideIcon> = {
  routes: Navigation,
  controllers: Code,
  middleware: Cpu,
  models: Database,
  form_requests: FileCheck,
  jobs: Briefcase,
  mailers: Mail,
  services: Wrench,
  errors: AlertCircle,
}

export function Stat({ card }: { card: StatCard }) {
  const Icon = KIND_ICON[card.kind] ?? Navigation
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3" title={card.hint ?? ''}>
      <div className="flex items-center justify-between">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          {card.label}
        </p>
        <Icon className="w-3.5 h-3.5 text-primary/70" />
      </div>
      <p className="text-2xl font-semibold mt-1.5 tabular-nums">{card.value}</p>
    </div>
  )
}

export function StatSkeleton() {
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3 space-y-2">
      <div className="flex items-center justify-between">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-3.5 w-3.5 rounded" />
      </div>
      <Skeleton className="h-7 w-12 mt-1.5" />
    </div>
  )
}
