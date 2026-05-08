import { Decoration, ViewPlugin, type DecorationSet, type EditorView, type ViewUpdate, WidgetType } from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'

export const VAR_PILL_EVENT = 'spectra:var-pill-click'

class VarPillWidget extends WidgetType {
  constructor(
    private name: string,
    private value: string | undefined,
    private from: number,
    private to: number,
  ) {
    super()
  }

  eq(other: VarPillWidget): boolean {
    return (
      other.name === this.name &&
      other.value === this.value &&
      other.from === this.from &&
      other.to === this.to
    )
  }

  toDOM(): HTMLElement {
    const wrap = document.createElement('span')
    wrap.className = 'cm-var-pill'
    wrap.dataset.varName = this.name
    wrap.dataset.hasValue = this.value !== undefined ? 'true' : 'false'
    wrap.title = this.value !== undefined ? `${this.name} = ${this.value}` : `${this.name} (unset)`

    const namePart = document.createElement('span')
    namePart.className = 'cm-var-pill-scope'
    namePart.textContent = this.name

    const valuePart = document.createElement('span')
    valuePart.className = 'cm-var-pill-name'
    if (this.value !== undefined && this.value !== '') {
      const truncated = this.value.length > 24 ? this.value.slice(0, 24) + '…' : this.value
      valuePart.textContent = truncated
    } else {
      valuePart.textContent = '∅'
    }

    wrap.appendChild(namePart)
    wrap.appendChild(valuePart)

    wrap.addEventListener('mousedown', (e) => {
      e.preventDefault()
      e.stopPropagation()
      const rect = wrap.getBoundingClientRect()
      wrap.dispatchEvent(
        new CustomEvent(VAR_PILL_EVENT, {
          bubbles: true,
          detail: {
            name: this.name,
            value: this.value,
            from: this.from,
            to: this.to,
            rect,
          },
        }),
      )
    })

    return wrap
  }

  ignoreEvent(): boolean {
    return false
  }
}

const VAR_REGEX = /\{\{([A-Za-z0-9_.\-]+)\}\}/g

export function variableDecorations(
  getVariables: () => Record<string, string>,
) {
  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet
      constructor(view: EditorView) {
        this.decorations = this.build(view)
      }
      update(update: ViewUpdate) {
        if (
          update.docChanged ||
          update.viewportChanged ||
          update.selectionSet ||
          update.focusChanged
        ) {
          this.decorations = this.build(update.view)
        }
      }
      build(view: EditorView): DecorationSet {
        const builder = new RangeSetBuilder<Decoration>()
        const vars = getVariables()
        const cursorPos = view.state.selection.main.head
        const hasFocus = view.hasFocus
        for (const range of view.visibleRanges) {
          const text = view.state.doc.sliceString(range.from, range.to)
          let m: RegExpExecArray | null
          VAR_REGEX.lastIndex = 0
          while ((m = VAR_REGEX.exec(text)) !== null) {
            const start = range.from + m.index
            const end = start + m[0].length
            if (hasFocus && cursorPos >= start && cursorPos <= end) continue
            const name = m[1]
            const value = vars[name]
            builder.add(
              start,
              end,
              Decoration.replace({
                widget: new VarPillWidget(name, value, start, end),
              }),
            )
          }
        }
        return builder.finish()
      }
    },
    { decorations: (v) => v.decorations },
  )
}

export const variableTheme = {
  '.cm-var-pill': {
    display: 'inline-flex',
    alignItems: 'stretch',
    height: '18px',
    margin: '0 2px',
    borderRadius: '4px',
    overflow: 'hidden',
    fontFamily: 'var(--font-mono)',
    fontSize: '10.5px',
    fontWeight: '600',
    lineHeight: '1',
    cursor: 'pointer',
    verticalAlign: 'middle',
    whiteSpace: 'nowrap',
    border: '1px solid var(--var-pill-border)',
  },
  '.cm-var-pill-scope': {
    display: 'inline-flex',
    alignItems: 'center',
    padding: '0 5px',
    background: 'var(--var-pill-name-bg)',
    color: 'var(--var-pill-name-fg)',
    borderRight: '1px solid var(--var-pill-name-border)',
  },
  '.cm-var-pill-name': {
    display: 'inline-flex',
    alignItems: 'center',
    padding: '0 5px',
    background: 'var(--var-pill-value-bg)',
    color: 'var(--var-pill-value-fg)',
  },
  '.cm-var-pill[data-has-value="false"]': {
    borderColor: 'var(--var-pill-unset-border)',
  },
  '.cm-var-pill[data-has-value="false"] .cm-var-pill-scope': {
    background: 'var(--var-pill-unset-name-bg)',
    color: 'var(--var-pill-unset-name-fg)',
    borderRightColor: 'var(--var-pill-unset-name-border)',
  },
  '.cm-var-pill[data-has-value="false"] .cm-var-pill-name': {
    background: 'var(--var-pill-unset-value-bg)',
    color: 'var(--var-pill-unset-value-fg)',
  },
}
