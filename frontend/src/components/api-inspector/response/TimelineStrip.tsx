export interface TimelineData {
  dnsMs: number
  connectMs: number
  tlsMs: number
  ttfbMs: number
  downloadMs: number
}

export function TimelineStrip({ timeline }: { timeline: TimelineData }) {
  const total =
    timeline.dnsMs + timeline.connectMs + timeline.tlsMs + timeline.ttfbMs + timeline.downloadMs
  if (total === 0) return null
  const segments = [
    { label: 'DNS', ms: timeline.dnsMs, color: 'bg-cyan-500/70' },
    { label: 'TCP', ms: timeline.connectMs, color: 'bg-blue-500/70' },
    { label: 'TLS', ms: timeline.tlsMs, color: 'bg-violet-500/70' },
    { label: 'TTFB', ms: timeline.ttfbMs, color: 'bg-amber-500/70' },
    { label: 'DL', ms: timeline.downloadMs, color: 'bg-emerald-500/70' },
  ]
  return (
    <div className="px-3 py-1.5 border-b border-border/40 flex items-center gap-2">
      <span className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/60 shrink-0">
        Timeline
      </span>
      <div className="flex h-2 flex-1 rounded-full overflow-hidden bg-muted/30">
        {segments.map((s) => {
          const pct = total > 0 ? (s.ms / total) * 100 : 0
          if (pct === 0) return null
          return (
            <div
              key={s.label}
              className={s.color}
              style={{ width: `${pct}%` }}
              title={`${s.label}: ${s.ms}ms`}
            />
          )
        })}
      </div>
      <div className="flex items-center gap-2 text-[9px] font-mono text-muted-foreground tabular-nums shrink-0">
        {segments
          .filter((s) => s.ms > 0)
          .map((s) => (
            <span key={s.label} className="flex items-center gap-1">
              <span className={`inline-block w-1.5 h-1.5 rounded-full ${s.color}`} />
              {s.label} {s.ms}ms
            </span>
          ))}
      </div>
    </div>
  )
}
