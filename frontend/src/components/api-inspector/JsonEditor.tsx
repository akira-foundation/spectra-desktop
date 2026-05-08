import { useEffect, useMemo, useRef } from 'react'
import CodeMirror, { EditorView, type ReactCodeMirrorRef } from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import { useTheme } from '@/hooks/useTheme'

interface JsonEditorProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  readOnly?: boolean
}

export function JsonEditor({ value, onChange, placeholder, className, readOnly }: JsonEditorProps) {
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

  const extensions = useMemo(
    () => [
      json(),
      EditorView.lineWrapping,
      EditorView.updateListener.of((u) => {
        if (u.docChanged) {
          onChangeRef.current(u.state.doc.toString())
        }
      }),
      EditorView.theme({
        '&': { backgroundColor: 'transparent', height: '100%' },
        '.cm-gutters': { backgroundColor: 'transparent', borderRight: 'none' },
        '.cm-content': { padding: '8px 4px', fontFamily: 'var(--font-mono)' },
        '.cm-focused': { outline: 'none' },
        '.cm-line': { padding: '0 4px' },
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

  return (
    <div
      className={`h-full w-full overflow-auto rounded-md border border-border/40 bg-muted/20 ${className ?? ''}`}
    >
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
          autocompletion: false,
        }}
        readOnly={readOnly}
        style={{ fontSize: '11.5px', height: '100%' }}
      />
    </div>
  )
}
