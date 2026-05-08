import { forwardRef, useEffect, useMemo, useRef, useState } from 'react'
import CodeMirror, { EditorView, type ReactCodeMirrorRef } from '@uiw/react-codemirror'
import { autocompletion, type CompletionContext } from '@codemirror/autocomplete'
import { keymap, tooltips } from '@codemirror/view'
import { Prec } from '@codemirror/state'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import { useTheme } from '@/hooks/useTheme'
import { variableDecorations, variableTheme, VAR_PILL_EVENT } from '@/lib/var-decoration'
import { cn } from '@/lib/utils'
import { VarPillPopover } from './VarPillPopover'

interface VarInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  disabled?: boolean
  variables?: Record<string, string>
  suggestions?: string[]
  scope?: string
  onKeyDown?: (e: React.KeyboardEvent) => void
  onBlur?: (e: React.FocusEvent) => void
}

export const VarInput = forwardRef<HTMLInputElement, VarInputProps>(function VarInput(
  { value, onChange, placeholder, className, disabled, variables, suggestions, scope, onKeyDown, onBlur },
  _ref,
) {
  void _ref
  const cmRef = useRef<ReactCodeMirrorRef>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [popover, setPopover] = useState<{
    name: string
    value?: string
    rect: { left: number; top: number; bottom: number }
    from: number
    to: number
  } | null>(null)
  const onChangeRef = useRef(onChange)
  const onKeyDownRef = useRef(onKeyDown)
  const onBlurRef = useRef(onBlur)
  const variablesRef = useRef<Record<string, string>>(variables ?? {})
  const suggestionsRef = useRef<string[]>(suggestions ?? [])
  const scopeRef = useRef<string>(scope ?? 'env')

  useEffect(() => {
    onChangeRef.current = onChange
    onKeyDownRef.current = onKeyDown
    onBlurRef.current = onBlur
  })
  useEffect(() => {
    variablesRef.current = variables ?? {}
  }, [variables])
  useEffect(() => {
    suggestionsRef.current = suggestions ?? []
  }, [suggestions])
  useEffect(() => {
    scopeRef.current = scope ?? 'env'
  }, [scope])

  const theme = useTheme((s) => s.theme)
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const extensions = useMemo(
    () => [
      EditorView.contentAttributes.of({ spellcheck: 'false', autocorrect: 'off', autocapitalize: 'off' }),
      tooltips({ position: 'fixed', parent: document.body }),
      Prec.highest(
        keymap.of([
          {
            key: 'Enter',
            run: () => {
              onKeyDownRef.current?.({
                key: 'Enter',
                preventDefault: () => {},
              } as unknown as React.KeyboardEvent)
              return true
            },
          },
        ]),
      ),
      EditorView.updateListener.of((u) => {
        if (u.docChanged) {
          const text = u.state.doc.toString().replace(/\n/g, '')
          onChangeRef.current(text)
        }
      }),
      autocompletion({
        activateOnTyping: true,
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
          (context: CompletionContext) => {
            const list = suggestionsRef.current
            if (!list.length) return null
            const doc = context.state.doc.toString()
            if (doc.includes('{{')) {
              const lastOpen = doc.lastIndexOf('{{', context.pos)
              const lastClose = doc.lastIndexOf('}}', context.pos)
              if (lastOpen > lastClose) return null
            }
            const word = context.matchBefore(/[A-Za-z0-9_\-]*/)
            const from = word ? word.from : context.pos
            const partial = (word?.text ?? '').toLowerCase()
            const filtered = partial
              ? list.filter((s) => s.toLowerCase().includes(partial) && s.toLowerCase() !== partial)
              : list
            if (!filtered.length) return null
            return {
              from,
              to: context.pos,
              options: filtered.map((s) => ({ label: s, type: 'keyword' })),
              filter: false,
            }
          },
        ],
      }),
      variableDecorations(() => variablesRef.current),
      EditorView.theme({
        '&': { backgroundColor: 'transparent' },
        '.cm-scroller': {
          overflow: 'auto',
          overflowY: 'hidden',
          scrollbarWidth: 'none',
        },
        '.cm-scroller::-webkit-scrollbar': { height: '0', width: '0', display: 'none' },
        '.cm-content': {
          padding: '5px 8px',
          fontFamily: 'var(--font-mono)',
          fontSize: '12px',
          minHeight: '18px',
          caretColor: 'var(--foreground)',
          whiteSpace: 'pre',
        },
        '.cm-line': { padding: '0', whiteSpace: 'pre' },
        '.cm-focused': { outline: 'none' },
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
      view.dispatch({ changes: { from: 0, to: current.length, insert: value } })
    }
  }, [value])

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

  return (
    <div
      ref={containerRef}
      className={cn(
        'w-full min-w-0 overflow-hidden rounded-md border border-border/50 bg-input/40 dark:bg-input/30 text-foreground transition-colors outline-none',
        'hover:border-border focus-within:border-border',
        disabled && 'opacity-60 pointer-events-none',
        className,
      )}
      onBlur={(e) => onBlurRef.current?.(e)}
    >
      <CodeMirror
        ref={cmRef}
        value={value}
        placeholder={placeholder}
        theme={isDark ? vscodeDark : vscodeLight}
        extensions={extensions}
        editable={!disabled}
        basicSetup={{
          lineNumbers: false,
          foldGutter: false,
          highlightActiveLine: false,
          highlightActiveLineGutter: false,
          autocompletion: true,
          history: false,
          drawSelection: true,
          dropCursor: false,
          allowMultipleSelections: false,
          indentOnInput: false,
          syntaxHighlighting: false,
          bracketMatching: false,
          closeBrackets: false,
          rectangularSelection: false,
          crosshairCursor: false,
          highlightSelectionMatches: false,
        }}
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
})
