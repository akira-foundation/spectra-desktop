export function HeadersList({ headers }: { headers?: Record<string, string[]> }) {
  const entries = Object.entries(headers ?? {})
  if (entries.length === 0) {
    return <p className="text-[11.5px] text-muted-foreground italic text-center">No headers</p>
  }
  return (
    <ul className="space-y-1">
      {entries.map(([k, vs]) => (
        <li key={k} className="flex gap-2 text-[11.5px] font-mono">
          <span className="text-muted-foreground/80 shrink-0 min-w-[140px]">{k}:</span>
          <span className="text-foreground/90 break-all">{vs.join(', ')}</span>
        </li>
      ))}
    </ul>
  )
}
