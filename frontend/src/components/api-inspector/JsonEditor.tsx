import { useEffect, useMemo, useRef, useState } from 'react'
import CodeMirror, { EditorView, type ReactCodeMirrorRef } from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'
import { autocompletion, type CompletionContext } from '@codemirror/autocomplete'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import { useTheme } from '@/hooks/useTheme'
import { variableDecorations, variableTheme, VAR_PILL_EVENT } from '@/lib/var-decoration'
import { VarPillPopover } from './VarPillPopover'

interface JsonEditorProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  readOnly?: boolean
  variables?: Record<string, string>
  scope?: string
}

export function JsonEditor({ value, onChange, placeholder, className, readOnly, variables, scope }: JsonEditorProps) {
  const theme = useTheme((s) => s.theme)
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const onChangeRef = useRef(onChange)
  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  const cmRef = useRef<ReactCodeMirrorRef>(null)
  const variablesRef = useRef<Record<string, string>>(variables ?? {})
  const scopeRef = useRef<string>(scope ?? 'env')
  useEffect(() => {
    variablesRef.current = variables ?? {}
    const view = cmRef.current?.view
    if (view) {
      view.dispatch({ effects: [] })
    }
  }, [variables])
  useEffect(() => {
    scopeRef.current = scope ?? 'env'
    const view = cmRef.current?.view
    if (view) view.dispatch({ effects: [] })
  }, [scope])

  const extensions = useMemo(
    () => [
      json(),
      EditorView.lineWrapping,
      EditorView.updateListener.of((u) => {
        if (u.docChanged) {
          onChangeRef.current(u.state.doc.toString())
        }
      }),
      autocompletion({
        override: [
          (context: CompletionContext) => {
            const before = context.matchBefore(/\{\{[A-Za-z0-9_.\-]*$/)
            if (!before) return null
            const names = Object.keys(variablesRef.current)
            if (!names.length) return null
            const partial = before.text.slice(2)
            return {
              from: before.from + 2,
              to: context.pos,
              options: names
                .filter((v) => v.toLowerCase().includes(partial.toLowerCase()))
                .map((v) => ({
                  label: v,
                  type: 'variable',
                  apply: (view, _completion, from, to) => {
                    const docAfter = view.state.doc.sliceString(to, to + 2)
                    const insert = docAfter === '}}' ? v : `${v}}}`
                    view.dispatch({
                      changes: { from, to, insert },
                      selection: { anchor: from + insert.length },
                    })
                  },
                })),
              filter: false,
            }
          },
        ],
      }),
      variableDecorations(() => variablesRef.current),
      EditorView.theme({
        '&': { backgroundColor: 'transparent', height: '100%' },
        '.cm-gutters': { backgroundColor: 'transparent', borderRight: 'none' },
        '.cm-content': { padding: '8px 4px', fontFamily: 'var(--font-mono)' },
        '.cm-focused': { outline: 'none' },
        '.cm-line': { padding: '0 4px' },
        ...variableTheme,
      }),
    ],
    [],
  )

  useEffect(() => {
    const view = cmRef.current?.view
    if (!view) return
    const current = view.state.doc.toString()
    if (current !== value) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: value },
      })
    }
  }, [value])

  const containerRef = useRef<HTMLDivElement>(null)
  const [popover, setPopover] = useState<{
    name: string
    value?: string
    rect: { left: number; top: number; bottom: number }
    from: number
    to: number
  } | null>(null)

  useEffect(() => {
    const node = containerRef.current
    if (!node) return
    const handler = (e: Event) => {
      const ev = e as CustomEvent<{
        name: string
        value?: string
        from: number
        to: number
        rect: DOMRect
      }>
      setPopover({
        name: ev.detail.name,
        value: ev.detail.value,
        from: ev.detail.from,
        to: ev.detail.to,
        rect: {
          left: ev.detail.rect.left,
          top: ev.detail.rect.top,
          bottom: ev.detail.rect.bottom,
        },
      })
    }
    node.addEventListener(VAR_PILL_EVENT, handler)
    return () => node.removeEventListener(VAR_PILL_EVENT, handler)
  }, [])

  const handleSwap = (newName: string) => {
    if (!popover) return
    const view = cmRef.current?.view
    if (!view) return
    const insert = `{{${newName}}}`
    view.dispatch({
      changes: { from: popover.from, to: popover.to, insert },
    })
    setPopover(null)
  }

  const validation = useMemo(() => validateJson(value), [value])
  const [dragOver, setDragOver] = useState(false)

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    if (readOnly) return
    const file = e.dataTransfer.files?.[0]
    if (!file) return
    const text = await file.text()
    onChangeRef.current(text)
  }

  return (
    <div
      ref={containerRef}
      onDragOver={(e) => {
        if (readOnly) return
        e.preventDefault()
        setDragOver(true)
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={handleDrop}
      className={`relative h-full w-full overflow-auto rounded-md border bg-muted/20 ${
        dragOver ? 'border-primary/60 bg-primary/5' :
        validation.ok ? 'border-border/40' : 'border-rose-500/40'
      } ${className ?? ''}`}
    >
      {!validation.ok && validation.message && (
        <div className="absolute top-1.5 right-2 z-10 inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-rose-500/15 text-rose-500 text-[10px] font-mono pointer-events-none">
          <span>JSON</span>
          <span className="text-rose-500/80">·</span>
          <span className="truncate max-w-[280px]" title={validation.message}>
            {validation.message}
          </span>
        </div>
      )}
      <CodeMirror
        ref={cmRef}
        value={value}
        placeholder={placeholder}
        theme={isDark ? vscodeDark : vscodeLight}
        extensions={extensions}
        basicSetup={{
          lineNumbers: false,
          foldGutter: false,
          highlightActiveLine: false,
          highlightActiveLineGutter: false,
          autocompletion: true,
        }}
        readOnly={readOnly}
        style={{ fontSize: '11.5px', height: '100%' }}
      />
      {popover && (
        <VarPillPopover
          name={popover.name}
          value={popover.value}
          variables={variables ?? {}}
          rect={popover.rect}
          onClose={() => setPopover(null)}
          onSwap={handleSwap}
        />
      )}
    </div>
  )
}

function validateJson(raw: string): { ok: boolean; message?: string } {
  const trimmed = raw.trim()
  if (!trimmed || trimmed === '{}' || trimmed === '[]') return { ok: true }
  // strip {{var}} placeholders so they don't break parsing
  const stripped = trimmed.replace(/\{\{[A-Za-z0-9_.\-]+\}\}/g, '"__var__"')
  try {
    JSON.parse(stripped)
    return { ok: true }
  } catch (err) {
    const msg = (err as Error).message.replace(/^JSON\.parse:\s*/i, '').replace(/^Unexpected token.*?in JSON/, 'syntax error')
    return { ok: false, message: msg }
  }
}
