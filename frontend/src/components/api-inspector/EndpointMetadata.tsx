import { Code2, FileText, Shield, KeyRound, Hash } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface EndpointMetadataProps {
  controller?: string
  sourceFile?: string
  sourceLine?: number
  middleware?: string[]
  authRequired?: boolean
  className?: string
}

export function EndpointMetadata({
  controller,
  sourceFile,
  sourceLine,
  middleware,
  authRequired,
  className,
}: EndpointMetadataProps) {
  const hasAny =
    controller ||
    sourceFile ||
    (middleware && middleware.length > 0) ||
    authRequired !== undefined
  if (!hasAny) return null

  return (
    <div
      className={cn(
        'border-b border-border/60 bg-transparent px-3.5 py-2 flex flex-col gap-1.5 text-[11.5px]',
        className,
      )}
    >
      {controller && <Field icon={Code2} label="Controller" value={controller} mono />}
      {sourceFile && (
        <Field
          icon={FileText}
          label="Source"
          value={sourceLine !== undefined ? `${sourceFile}:${sourceLine}` : sourceFile}
          mono
        />
      )}
      {middleware && middleware.length > 0 && <MiddlewareField middleware={middleware} />}
      {authRequired !== undefined && (
        <Field
          icon={KeyRound}
          label="Auth"
          value={authRequired ? 'Required' : 'Public'}
          tone={authRequired ? 'primary' : 'muted'}
        />
      )}
    </div>
  )
}

interface FieldProps {
  icon: LucideIcon
  label: string
  value: string
  mono?: boolean
  tone?: 'default' | 'muted' | 'primary'
}

function Field({ icon: Icon, label, value, mono, tone = 'default' }: FieldProps) {
  return (
    <div className="grid grid-cols-[110px_1fr] items-center gap-x-3 min-w-0">
      <div className="flex items-center gap-1.5 min-w-0">
        <Icon className="w-3 h-3 text-muted-foreground/70 shrink-0" />
        <span className="text-muted-foreground uppercase tracking-wider text-[10px] font-medium">
          {label}
        </span>
      </div>
      <span
        className={cn(
          'truncate min-w-0',
          mono && 'font-mono text-[11px]',
          tone === 'muted' && 'text-muted-foreground',
          tone === 'primary' && 'text-primary',
          tone === 'default' && 'text-foreground/85',
        )}
      >
        {value}
      </span>
    </div>
  )
}

function MiddlewareField({ middleware }: { middleware: string[] }) {
  return (
    <div className="grid grid-cols-[110px_1fr] items-center gap-x-3 min-w-0">
      <div className="flex items-center gap-1.5 min-w-0">
        <Shield className="w-3 h-3 text-muted-foreground/70 shrink-0" />
        <span className="text-muted-foreground uppercase tracking-wider text-[10px] font-medium">
          Middleware
        </span>
      </div>
      <div className="flex items-center gap-1.5 flex-wrap min-w-0">
        {middleware.map((m) => (
          <span
            key={m}
            className="inline-flex items-center h-[18px] text-[10px] font-mono px-1.5 rounded border border-border/60 bg-muted/40 text-foreground/80"
          >
            <Hash className="w-2.5 h-2.5 mr-0.5 text-muted-foreground/70" />
            {m}
          </span>
        ))}
      </div>
    </div>
  )
}
