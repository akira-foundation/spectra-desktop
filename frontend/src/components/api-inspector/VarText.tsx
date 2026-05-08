import { cn } from '@/lib/utils'

interface VarTextProps {
  value: string
  variables?: Record<string, string>
  className?: string
}

const VAR_REGEX = /\{\{([A-Za-z0-9_.\-]+)\}\}/g

export function VarText({ value, variables, className }: VarTextProps) {
  if (!value) return <span className={className}>{value}</span>

  const parts: Array<{ kind: 'text' | 'var'; content: string; resolved?: string }> = []
  let last = 0
  let m: RegExpExecArray | null
  VAR_REGEX.lastIndex = 0
  while ((m = VAR_REGEX.exec(value)) !== null) {
    if (m.index > last) {
      parts.push({ kind: 'text', content: value.slice(last, m.index) })
    }
    const name = m[1]
    parts.push({ kind: 'var', content: name, resolved: variables?.[name] })
    last = m.index + m[0].length
  }
  if (last < value.length) {
    parts.push({ kind: 'text', content: value.slice(last) })
  }

  return (
    <span className={cn('inline', className)}>
      {parts.map((p, i) =>
        p.kind === 'text' ? (
          <span key={i}>{p.content}</span>
        ) : (
          <VarPill key={i} name={p.content} value={p.resolved} />
        ),
      )}
    </span>
  )
}

interface VarPillProps {
  name: string
  value?: string
}

function VarPill({ name, value }: VarPillProps) {
  const hasValue = value !== undefined && value !== ''
  const truncated = hasValue && value!.length > 24 ? value!.slice(0, 24) + '…' : value
  const borderVar = hasValue ? 'var(--var-pill-border)' : 'var(--var-pill-unset-border)'
  const nameBg = hasValue ? 'var(--var-pill-name-bg)' : 'var(--var-pill-unset-name-bg)'
  const nameFg = hasValue ? 'var(--var-pill-name-fg)' : 'var(--var-pill-unset-name-fg)'
  const nameBorder = hasValue ? 'var(--var-pill-name-border)' : 'var(--var-pill-unset-name-border)'
  const valBg = hasValue ? 'var(--var-pill-value-bg)' : 'var(--var-pill-unset-value-bg)'
  const valFg = hasValue ? 'var(--var-pill-value-fg)' : 'var(--var-pill-unset-value-fg)'
  return (
    <span
      title={hasValue ? `${name} = ${value}` : `${name} (unset)`}
      style={{ borderColor: borderVar }}
      className={cn(
        'inline-flex items-stretch h-[18px] mx-0.5 rounded overflow-hidden border align-middle font-mono text-[10.5px] font-semibold leading-[18px]',
      )}
    >
      <span
        style={{ background: nameBg, color: nameFg, borderRight: `1px solid ${nameBorder}` }}
        className="px-1.5"
      >
        {name}
      </span>
      <span style={{ background: valBg, color: valFg }} className="px-1.5">
        {hasValue ? truncated : '∅'}
      </span>
    </span>
  )
}
