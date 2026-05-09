import { useEffect, useRef, useState } from 'react'
import { JsonEditor } from './JsonEditor'
import { FormBodyEditor } from './FormBodyEditor'
import { MultipartEditor, type MultipartPart } from './MultipartEditor'
import type { RequestSchema } from '@/lib/request-schema'
import { cn } from '@/lib/utils'

const NO_BODY_METHODS = new Set(['GET', 'HEAD', 'DELETE', 'OPTIONS'])

interface Props {
  method?: string | null
  body: string
  onBodyChange: (v: string) => void
  schema?: RequestSchema | null
  schemaBadge?: React.ReactNode
  variables?: Record<string, string>
  multipart?: MultipartPart[]
  onMultipartChange?: (parts: MultipartPart[]) => void
  noBodyOverride?: boolean
}

export function BodyEditor({
  method,
  body,
  onBodyChange,
  schema,
  schemaBadge,
  variables,
  multipart,
  onMultipartChange,
  noBodyOverride,
}: Props) {
  const [mode, setMode] = useState<'json' | 'form' | 'multipart'>('json')
  const isNoBody =
    !noBodyOverride && method !== undefined && method !== null && NO_BODY_METHODS.has(method.toUpperCase())

  if (isNoBody) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center text-center px-6 py-8 gap-2">
        <p className="text-[12.5px] font-medium text-foreground/85">No body needed</p>
        <p className="text-[11px] text-muted-foreground max-w-xs leading-relaxed">
          {(method ?? '').toUpperCase()} requests don&apos;t typically carry a payload.
        </p>
      </div>
    )
  }

  const formattedRef = useRef<string | null>(null)
  useEffect(() => {
    if (mode !== 'json') return
    const trimmed = body.trim()
    if (!trimmed) return
    if (trimmed.includes('\n')) return
    if (formattedRef.current === trimmed) return
    try {
      const pretty = JSON.stringify(JSON.parse(trimmed), null, 2)
      if (pretty !== body) {
        formattedRef.current = trimmed
        onBodyChange(pretty)
      }
    } catch {}
  }, [body, mode, onBodyChange])

  return (
    <>
      <div className="flex items-center gap-2 mb-2">
        <ModeButton active={mode === 'json'} onClick={() => setMode('json')}>
          JSON
        </ModeButton>
        <ModeButton
          active={mode === 'form'}
          disabled={!schema || schema.fields.length === 0}
          onClick={() => setMode('form')}
        >
          Form
        </ModeButton>
        {onMultipartChange && (
          <ModeButton active={mode === 'multipart'} onClick={() => setMode('multipart')}>
            Multipart
          </ModeButton>
        )}
        {schemaBadge}
      </div>
      <div className="flex-1 min-h-0 overflow-auto">
        {mode === 'multipart' && onMultipartChange ? (
          <MultipartEditor parts={multipart ?? []} onChange={onMultipartChange} />
        ) : mode === 'form' && schema && schema.fields.length > 0 ? (
          <FormBodyEditor value={body} schema={schema} onChange={onBodyChange} variables={variables} />
        ) : (
          <JsonEditor value={body} onChange={onBodyChange} placeholder="{}" variables={variables} />
        )}
      </div>
    </>
  )
}

function ModeButton({
  children,
  active,
  disabled,
  onClick,
}: {
  children: React.ReactNode
  active: boolean
  disabled?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={cn(
        'px-2 py-0.5 text-[10.5px] rounded-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed',
        active
          ? 'bg-primary/15 text-primary hover:bg-primary/25'
          : 'text-muted-foreground hover:bg-accent/60',
      )}
    >
      {children}
    </button>
  )
}
