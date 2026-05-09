import { useState } from 'react'
import { Info } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'

export interface TimelineData {
  dnsMs: number
  connectMs: number
  tlsMs: number
  ttfbMs: number
  downloadMs: number
}

interface Phase {
  key: keyof TimelineData
  label: string
  ms: number
  color: string
  dot: string
  what: string
  why: string
}

function buildPhases(t: TimelineData): Phase[] {
  return [
    {
      key: 'dnsMs',
      label: 'DNS',
      ms: t.dnsMs,
      color: 'bg-cyan-500/70',
      dot: 'bg-cyan-500',
      what: 'Resolve the domain name into an IP address.',
      why: 'Slow? System or ISP resolver is sluggish, or no cache. Try a different resolver (1.1.1.1, 8.8.8.8).',
    },
    {
      key: 'connectMs',
      label: 'TCP',
      ms: t.connectMs,
      color: 'bg-blue-500/70',
      dot: 'bg-blue-500',
      what: 'TCP handshake — open a connection to the server.',
      why: 'Slow? Server is far (network latency). Connection reuse skips this phase entirely.',
    },
    {
      key: 'tlsMs',
      label: 'TLS',
      ms: t.tlsMs,
      color: 'bg-violet-500/70',
      dot: 'bg-violet-500',
      what: 'HTTPS handshake — exchange certificate and derive session keys.',
      why: 'HTTPS only. Paid once per connection (reuse skips it). Slow? Heavy certificate or poor network.',
    },
    {
      key: 'ttfbMs',
      label: 'TTFB',
      ms: t.ttfbMs,
      color: 'bg-amber-500/70',
      dot: 'bg-amber-500',
      what: 'Time To First Byte — from sending the request to receiving the first response byte.',
      why: 'Includes server-side processing (framework boot, middleware, controller, DB queries, render). This is usually where most slowness lives. Optimize with caching, eager loading, indexes.',
    },
    {
      key: 'downloadMs',
      label: 'DL',
      ms: t.downloadMs,
      color: 'bg-emerald-500/70',
      dot: 'bg-emerald-500',
      what: 'Download — receive the response body after the first byte.',
      why: 'Slow? Large payload or weak client network. Optimize with pagination, gzip/brotli, smaller payloads.',
    },
  ]
}

const SLOWEST_HINT: Record<keyof TimelineData, string> = {
  ttfbMs: 'Server is taking time to respond — investigate queries, middleware, render path.',
  downloadMs: 'Large response or slow network — consider pagination or compression.',
  tlsMs: 'TLS is heavy — certificate or first connection (reusing connections helps).',
  connectMs: 'Network latency to the server — geographic distance.',
  dnsMs: 'DNS resolver is slow — switching resolver may help.',
}

export function TimelineStrip({ timeline }: { timeline: TimelineData }) {
  const [open, setOpen] = useState(false)
  const phases = buildPhases(timeline)
  const total = phases.reduce((s, p) => s + p.ms, 0)
  if (total === 0) return null
  const slowest = phases.filter((p) => p.ms > 0).sort((a, b) => b.ms - a.ms)[0]

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="w-full px-3 py-1.5 border-b border-border/40 flex items-center gap-2 hover:bg-accent/20 transition-colors text-left"
        title="Click for explanation"
      >
        <span className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/60 shrink-0 inline-flex items-center gap-1">
          Timeline
          <Info className="w-2.5 h-2.5" />
        </span>
        <div className="flex h-2 flex-1 rounded-full overflow-hidden bg-muted/30">
          {phases.map((p) => {
            const pct = (p.ms / total) * 100
            if (pct === 0) return null
            return (
              <div
                key={p.key}
                className={p.color}
                style={{ width: `${pct}%` }}
                title={`${p.label}: ${p.ms}ms`}
              />
            )
          })}
        </div>
        <div className="flex items-center gap-2 text-[9px] font-mono text-muted-foreground tabular-nums shrink-0">
          {phases
            .filter((p) => p.ms > 0)
            .map((p) => (
              <span key={p.key} className="flex items-center gap-1">
                <span className={`inline-block w-1.5 h-1.5 rounded-full ${p.dot}`} />
                {p.label} {p.ms}ms
              </span>
            ))}
        </div>
      </button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-xl max-h-[85vh] flex flex-col gap-0 p-0 overflow-hidden">
          <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
            <DialogTitle className="text-base">Request timeline</DialogTitle>
            <DialogDescription className="text-[12.5px]">
              Each phase is a step between hitting Run and the response arriving. Total{' '}
              <span className="font-mono text-foreground tabular-nums">{total}ms</span>.
            </DialogDescription>
          </DialogHeader>

          <div className="px-6 py-3 border-b border-border/40 bg-muted/20 shrink-0">
            <p className="text-[11px] text-muted-foreground">
              Real order:{' '}
              <span className="font-mono text-foreground/85">
                DNS → TCP → TLS → server processing → TTFB → DL
              </span>
            </p>
          </div>

          <ul className="m-0 p-0 list-none divide-y divide-border/30 flex-1 overflow-y-auto">
            {phases.map((p) => {
              const pct = total > 0 ? (p.ms / total) * 100 : 0
              const dim = p.ms === 0
              return (
                <li key={p.key} className={`px-6 py-3 ${dim ? 'opacity-50' : ''}`}>
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`inline-block w-2 h-2 rounded-full ${p.dot}`} />
                    <span className="text-[12px] font-mono font-semibold text-foreground">{p.label}</span>
                    <span className="text-[11px] font-mono tabular-nums text-muted-foreground ml-auto">
                      {p.ms}ms · {pct.toFixed(0)}%
                    </span>
                  </div>
                  <p className="text-[11.5px] text-foreground/85 leading-relaxed">{p.what}</p>
                  <p className="text-[11px] text-muted-foreground leading-relaxed mt-1">{p.why}</p>
                </li>
              )
            })}
          </ul>

          {slowest && (
            <div className="px-6 py-3 shrink-0 border-t border-border/40 bg-amber-500/5">
              <p className="text-[11px] text-foreground/85 leading-relaxed">
                <span className="font-semibold text-amber-500/90">Where to look first:</span>{' '}
                <span className="font-mono">{slowest.label}</span> dominates (
                {((slowest.ms / total) * 100).toFixed(0)}% of total). {SLOWEST_HINT[slowest.key]}
              </p>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
