import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'

interface Props {
  value: string
  onChange: (v: string) => void
  onComplete?: (v: string) => void
  disabled?: boolean
  length?: number
  autoFocus?: boolean
}

export function OtpInput({
  value,
  onChange,
  onComplete,
  disabled,
  length = 6,
  autoFocus,
}: Props) {
  const inputsRef = useRef<Array<HTMLInputElement | null>>([])

  useEffect(() => {
    if (autoFocus) inputsRef.current[0]?.focus()
  }, [autoFocus])

  useEffect(() => {
    if (value.length === length && onComplete) onComplete(value)
  }, [value, length, onComplete])

  const digits = Array.from({ length }).map((_, i) => value[i] ?? '')

  const handle = (idx: number, raw: string) => {
    const cleaned = raw.replace(/[^0-9]/g, '')
    if (!cleaned) {
      const next = value.slice(0, idx) + value.slice(idx + 1)
      onChange(next)
      return
    }
    if (cleaned.length > 1) {
      onChange((value + cleaned).slice(0, length))
      const focusIdx = Math.min(value.length + cleaned.length, length - 1)
      inputsRef.current[focusIdx]?.focus()
      return
    }
    const next = (value.slice(0, idx) + cleaned + value.slice(idx + 1)).slice(0, length)
    onChange(next)
    if (idx < length - 1) inputsRef.current[idx + 1]?.focus()
  }

  const handleKey = (idx: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !digits[idx] && idx > 0) {
      inputsRef.current[idx - 1]?.focus()
    } else if (e.key === 'ArrowLeft' && idx > 0) {
      inputsRef.current[idx - 1]?.focus()
    } else if (e.key === 'ArrowRight' && idx < length - 1) {
      inputsRef.current[idx + 1]?.focus()
    }
  }

  return (
    <div className="flex items-center gap-2">
      {digits.map((d, i) => (
        <input
          key={i}
          ref={(el) => {
            inputsRef.current[i] = el
          }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={d}
          disabled={disabled}
          onChange={(e) => handle(i, e.target.value)}
          onKeyDown={(e) => handleKey(i, e)}
          onPaste={(e) => {
            e.preventDefault()
            handle(0, e.clipboardData.getData('text'))
          }}
          className={cn(
            'h-11 w-10 text-center text-[15px] font-mono rounded-md border border-border/60 bg-input/60 dark:bg-input/40 outline-none transition-colors',
            'focus:border-primary focus:ring-1 focus:ring-primary/40',
            disabled && 'opacity-50 cursor-not-allowed',
          )}
        />
      ))}
    </div>
  )
}
