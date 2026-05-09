import { useEffect, useState } from 'react'

interface FormattedException {
  message: string
  class?: string
  file?: string
  line?: number
  trace?: { function: string; file?: string; line?: number }[]
}

interface Props {
  projectId: string | null
  body: string
  status: number
}

export function ExceptionPanel({ projectId, body, status }: Props) {
  const [exc, setExc] = useState<FormattedException | null>(null)
  useEffect(() => {
    let cancelled = false
    if (!projectId || !body || status < 400) return
    void import('../../../../wailsjs/go/app/App').then(async ({ FormatException }) => {
      try {
        const result = await FormatException(projectId, body, status)
        if (cancelled) return
        if (Array.isArray(result) && result[1]) setExc(result[0] as FormattedException)
        else if (result && typeof result === 'object' && (result as FormattedException).message) {
          setExc(result as FormattedException)
        }
      } catch {}
    })
    return () => {
      cancelled = true
    }
  }, [projectId, body, status])
  if (!exc) return null
  return (
    <div className="px-3 py-2 border-b border-border/40 bg-rose-500/5">
      <div className="flex items-baseline gap-2 mb-1">
        <span className="text-[10px] font-mono uppercase tracking-wider text-rose-500/90">
          Exception
        </span>
        {exc.class && <code className="text-[10.5px] font-mono text-rose-500/80">{exc.class}</code>}
      </div>
      <p className="text-[11.5px] text-foreground/90 mb-1.5">{exc.message}</p>
      {exc.file && (
        <p className="text-[10px] font-mono text-muted-foreground">
          {exc.file}
          {exc.line ? `:${exc.line}` : ''}
        </p>
      )}
      {exc.trace && exc.trace.length > 0 && (
        <details className="mt-1.5">
          <summary className="text-[10px] text-muted-foreground/70 cursor-pointer hover:text-foreground">
            Stack trace ({exc.trace.length})
          </summary>
          <ul className="m-0 mt-1 p-0 list-none space-y-0.5 max-h-40 overflow-auto">
            {exc.trace.map((t, i) => (
              <li key={i} className="text-[10px] font-mono text-muted-foreground">
                <span className="text-foreground/70">{t.function}</span>
                {t.file && (
                  <span className="text-muted-foreground/60">
                    {' '}
                    @ {t.file}
                    {t.line ? `:${t.line}` : ''}
                  </span>
                )}
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  )
}
