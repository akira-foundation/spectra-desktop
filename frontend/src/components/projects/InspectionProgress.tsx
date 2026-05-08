import { Check, Loader2, AlertCircle, Circle } from 'lucide-react'
import type { PipelineStepState } from '@/hooks/useInspectionPipeline'
import { cn } from '@/lib/utils'

interface InspectionProgressProps {
  steps: PipelineStepState[]
  title?: string
}

export function InspectionProgress({ steps, title = 'Inspecting project' }: InspectionProgressProps) {
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-4 space-y-3">
      <div className="flex items-center gap-2">
        <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
        <h3 className="text-[12.5px] font-semibold tracking-tight">{title}</h3>
      </div>
      <ol className="space-y-1.5">
        {steps.map((step) => (
          <Step key={step.id} step={step} />
        ))}
      </ol>
    </div>
  )
}

function Step({ step }: { step: PipelineStepState }) {
  return (
    <li
      className={cn(
        'flex items-center gap-2.5 text-[12px] transition-colors duration-200',
        step.status === 'pending' && 'text-muted-foreground/60',
        step.status === 'active' && 'text-foreground',
        step.status === 'done' && 'text-foreground/80',
        step.status === 'error' && 'text-destructive',
      )}
    >
      <StepIcon status={step.status} />
      <span className="font-medium">{step.label}</span>
      {step.error && <span className="text-[11px] text-destructive/80 ml-auto">{step.error}</span>}
    </li>
  )
}

function StepIcon({ status }: { status: PipelineStepState['status'] }) {
  switch (status) {
    case 'done':
      return <Check className="w-3.5 h-3.5 text-emerald-500 animate-in fade-in zoom-in-90 duration-200" />
    case 'active':
      return <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
    case 'error':
      return <AlertCircle className="w-3.5 h-3.5 text-destructive" />
    default:
      return <Circle className="w-3.5 h-3.5 text-muted-foreground/40" />
  }
}
