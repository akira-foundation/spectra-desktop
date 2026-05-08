import { useEffect, useMemo, useState } from 'react'
import type { RequestSchema, InferredField } from '@/lib/request-schema'
import { VarInput } from './VarInput'
import { cn } from '@/lib/utils'

interface FormBodyEditorProps {
  value: string
  schema: RequestSchema | null
  onChange: (value: string) => void
  variables?: Record<string, string>
}

export function FormBodyEditor({ value, schema, onChange, variables }: FormBodyEditorProps) {
  const fields = schema?.fields ?? []
  const parsed = useMemo(() => parseBody(value), [value])
  const [local, setLocal] = useState<Record<string, string>>(parsed)

  useEffect(() => {
    setLocal(parsed)
  }, [parsed])

  if (fields.length === 0) {
    return (
      <p className="text-[11.5px] text-muted-foreground italic p-2">
        No schema inferred for this endpoint. Use JSON tab.
      </p>
    )
  }

  const setField = (name: string, raw: string) => {
    const next = { ...local, [name]: raw }
    setLocal(next)
    onChange(serialize(fields, next))
  }

  return (
    <div className="space-y-2 min-w-0">
      {fields.map((f) => (
        <FieldRow
          key={f.name}
          field={f}
          value={local[f.name] ?? ''}
          onChange={(v) => setField(f.name, v)}
          variables={variables}
        />
      ))}
    </div>
  )
}

interface FieldRowProps {
  field: InferredField
  value: string
  onChange: (value: string) => void
  variables?: Record<string, string>
}

function FieldRow({ field, value, onChange, variables }: FieldRowProps) {
  return (
    <div className="grid grid-cols-[1fr_1.6fr] gap-2 items-center min-w-0">
      <div className="flex items-center gap-1 min-w-0">
        <span className="text-[11.5px] font-mono truncate text-foreground/85">{field.name}</span>
        {field.required && <span className="text-destructive text-[10px]">*</span>}
        <span className="ml-auto text-[10px] uppercase tracking-wider text-muted-foreground/70 shrink-0">
          {field.type || 'string'}
        </span>
      </div>
      {renderInput(field, value, onChange, variables)}
    </div>
  )
}

function renderInput(
  field: InferredField,
  value: string,
  onChange: (value: string) => void,
  variables?: Record<string, string>,
) {
  if (field.type === 'boolean') {
    return (
      <select
        value={value || 'false'}
        onChange={(e) => onChange(e.target.value)}
        className="h-7 text-[12px] bg-input/40 border border-border/50 rounded-md px-2 font-mono"
      >
        <option value="false">false</option>
        <option value="true">true</option>
      </select>
    )
  }
  if (field.type === 'array' || field.type === 'object') {
    return (
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={field.type === 'array' ? '[]' : '{}'}
        className={cn(
          'h-16 text-[11px] bg-input/40 border border-border/50 rounded-md px-2 py-1 font-mono resize-none',
          'focus:outline-none focus:border-border',
        )}
      />
    )
  }
  return (
    <VarInput
      value={value}
      onChange={onChange}
      placeholder={placeholderFor(field)}
      className="h-7 text-[12px] font-mono"
      variables={variables}
    />
  )
}

function placeholderFor(field: InferredField): string {
  switch (field.type) {
    case 'integer':
    case 'numeric':
      return '0'
    case 'email':
      return 'user@example.com'
    case 'url':
      return 'https://example.com'
    case 'date':
      return '2025-01-01'
    case 'uuid':
      return 'uuid'
    default:
      return field.name
  }
}

function parseBody(raw: string): Record<string, string> {
  if (!raw.trim()) return {}
  try {
    const obj = JSON.parse(raw) as Record<string, unknown>
    const out: Record<string, string> = {}
    for (const [k, v] of Object.entries(obj)) {
      if (v == null) {
        out[k] = ''
      } else if (typeof v === 'object') {
        out[k] = JSON.stringify(v)
      } else {
        out[k] = String(v)
      }
    }
    return out
  } catch {
    return {}
  }
}

function serialize(fields: InferredField[], values: Record<string, string>): string {
  const out: Record<string, unknown> = {}
  for (const f of fields) {
    const raw = values[f.name] ?? ''
    out[f.name] = coerce(f.type, raw)
  }
  return JSON.stringify(out, null, 2)
}

function coerce(type: string, raw: string): unknown {
  if (raw === '') {
    if (type === 'boolean') return false
    if (type === 'integer' || type === 'numeric') return 0
    if (type === 'array') return []
    if (type === 'object') return {}
    return ''
  }
  switch (type) {
    case 'integer': {
      const n = parseInt(raw, 10)
      return isNaN(n) ? raw : n
    }
    case 'numeric': {
      const n = parseFloat(raw)
      return isNaN(n) ? raw : n
    }
    case 'boolean':
      return raw === 'true'
    case 'array':
    case 'object':
      try {
        return JSON.parse(raw)
      } catch {
        return raw
      }
    default:
      return raw
  }
}
