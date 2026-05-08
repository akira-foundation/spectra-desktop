import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

interface PathAutocompleteProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  suggestions: string[]
  className?: string
}

export function PathAutocomplete({
  value,
  onChange,
  placeholder,
  suggestions,
  className,
}: PathAutocompleteProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const wrapRef = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false)
  const [highlight, setHighlight] = useState(0)
  const [rect, setRect] = useState<{ left: number; top: number; width: number } | null>(null)

  const filtered = (() => {
    if (!value) return suggestions.slice(0, 12)
    const lower = value.toLowerCase()
    return suggestions
      .filter((s) => s.toLowerCase().includes(lower) && s !== value)
      .slice(0, 12)
  })()

  useEffect(() => {
    setHighlight(0)
  }, [value])

  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [])

  useLayoutEffect(() => {
    if (!open || !wrapRef.current) return
    const update = () => {
      if (!wrapRef.current) return
      const r = wrapRef.current.getBoundingClientRect()
      setRect({ left: r.left, top: r.bottom, width: r.width })
    }
    update()
    window.addEventListener('scroll', update, true)
    window.addEventListener('resize', update)
    return () => {
      window.removeEventListener('scroll', update, true)
      window.removeEventListener('resize', update)
    }
  }, [open, value])

  const select = (path: string) => {
    onChange(path)
    setOpen(false)
    requestAnimationFrame(() => inputRef.current?.focus())
  }

  const handleKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!open || filtered.length === 0) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlight((i) => (i + 1) % filtered.length)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlight((i) => (i - 1 + filtered.length) % filtered.length)
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      if (filtered[highlight]) {
        e.preventDefault()
        select(filtered[highlight])
      }
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <div ref={wrapRef} className="relative w-full">
      <Input
        ref={inputRef}
        value={value}
        onChange={(e) => {
          onChange(e.target.value)
          setOpen(true)
        }}
        onFocus={() => setOpen(true)}
        onKeyDown={handleKey}
        placeholder={placeholder}
        className={cn('font-mono', className)}
      />
      {open && filtered.length > 0 && rect &&
        createPortal(
          <div
            style={{
              position: 'fixed',
              left: rect.left,
              top: rect.top + 4,
              width: rect.width,
              zIndex: 60,
            }}
            className="max-h-56 overflow-auto rounded-md border border-border bg-popover shadow-md"
          >
            <ul className="py-1">
              {filtered.map((p, i) => (
                <li key={p}>
                  <button
                    type="button"
                    onMouseDown={(e) => {
                      e.preventDefault()
                      select(p)
                    }}
                    onMouseEnter={() => setHighlight(i)}
                    className={cn(
                      'w-full text-left px-2 py-1 text-[11.5px] font-mono',
                      i === highlight
                        ? 'bg-accent text-foreground'
                        : 'text-foreground/85 hover:bg-accent/40',
                    )}
                  >
                    {p}
                  </button>
                </li>
              ))}
            </ul>
          </div>,
          document.body,
        )}
    </div>
  )
}
