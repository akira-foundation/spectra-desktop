import { Info } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

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
      what: 'Resolver o domínio em endereço IP.',
      why: 'Lento? DNS do sistema/ISP devagar ou sem cache. Trocar resolver (1.1.1.1, 8.8.8.8).',
    },
    {
      key: 'connectMs',
      label: 'TCP',
      ms: t.connectMs,
      color: 'bg-blue-500/70',
      dot: 'bg-blue-500',
      what: 'Aperto de mão TCP — abrir conexão até o servidor.',
      why: 'Lento? Servidor longe (latência de rede). Reuso de conexão pula essa fase.',
    },
    {
      key: 'tlsMs',
      label: 'TLS',
      ms: t.tlsMs,
      color: 'bg-violet-500/70',
      dot: 'bg-violet-500',
      what: 'Aperto de mão HTTPS — trocar certificado e gerar chaves de sessão.',
      why: 'Só em HTTPS. Só na primeira conexão (depois reusa). Lento? Servidor com certificado pesado ou rede ruim.',
    },
    {
      key: 'ttfbMs',
      label: 'TTFB',
      ms: t.ttfbMs,
      color: 'bg-amber-500/70',
      dot: 'bg-amber-500',
      what: 'Time To First Byte — tempo entre mandar a request e o servidor começar a responder.',
      why: 'Inclui o processamento do servidor (Laravel boot, middleware, controller, query no banco, render). É aqui que mora a maior parte do tempo lento. Otimizar = cache, eager load, índices.',
    },
    {
      key: 'downloadMs',
      label: 'DL',
      ms: t.downloadMs,
      color: 'bg-emerald-500/70',
      dot: 'bg-emerald-500',
      what: 'Download — baixar o corpo da resposta após o primeiro byte.',
      why: 'Lento? Resposta grande ou rede do cliente fraca. Otimizar = paginação, gzip/brotli, payload menor.',
    },
  ]
}

export function TimelineStrip({ timeline }: { timeline: TimelineData }) {
  const phases = buildPhases(timeline)
  const total = phases.reduce((s, p) => s + p.ms, 0)
  if (total === 0) return null
  const slowest = phases.filter((p) => p.ms > 0).sort((a, b) => b.ms - a.ms)[0]

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
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
      </PopoverTrigger>
      <PopoverContent align="start" className="w-[420px] p-0">
        <div className="px-3 py-2 border-b border-border/40">
          <p className="text-[11px] font-semibold">Linha do tempo da request</p>
          <p className="text-[10.5px] text-muted-foreground leading-relaxed mt-0.5">
            Cada fase é uma etapa entre você apertar Run e a resposta chegar. Total{' '}
            <span className="font-mono text-foreground tabular-nums">{total}ms</span>.
          </p>
        </div>
        <div className="px-3 py-2 border-b border-border/40 bg-muted/20">
          <p className="text-[10.5px] text-muted-foreground">
            Ordem real:{' '}
            <span className="font-mono text-foreground/85">
              DNS → TCP → TLS → server pensa → TTFB → DL
            </span>
          </p>
        </div>
        <ul className="m-0 p-0 list-none divide-y divide-border/30 max-h-80 overflow-y-auto">
          {phases.map((p) => {
            const pct = total > 0 ? (p.ms / total) * 100 : 0
            const dim = p.ms === 0
            return (
              <li key={p.key} className={`px-3 py-2 ${dim ? 'opacity-50' : ''}`}>
                <div className="flex items-center gap-2 mb-1">
                  <span className={`inline-block w-1.5 h-1.5 rounded-full ${p.dot}`} />
                  <span className="text-[10.5px] font-mono font-semibold text-foreground">{p.label}</span>
                  <span className="text-[10px] font-mono tabular-nums text-muted-foreground ml-auto">
                    {p.ms}ms · {pct.toFixed(0)}%
                  </span>
                </div>
                <p className="text-[10.5px] text-foreground/80 leading-relaxed">{p.what}</p>
                <p className="text-[10px] text-muted-foreground leading-relaxed mt-0.5">{p.why}</p>
              </li>
            )
          })}
        </ul>
        {slowest && (
          <div className="px-3 py-2 border-t border-border/40 bg-amber-500/5">
            <p className="text-[10px] text-muted-foreground leading-relaxed">
              <span className="font-semibold text-amber-500/90">Onde olhar primeiro:</span>{' '}
              <span className="font-mono">{slowest.label}</span> domina (
              {((slowest.ms / total) * 100).toFixed(0)}% do tempo).{' '}
              {slowest.key === 'ttfbMs'
                ? 'Servidor está demorando — investigar query, middleware, render.'
                : slowest.key === 'downloadMs'
                  ? 'Resposta grande ou rede lenta — considerar paginação ou compressão.'
                  : slowest.key === 'tlsMs'
                    ? 'TLS pesado — certificado/cipher ou primeira conexão (reusar conexões ajuda).'
                    : slowest.key === 'connectMs'
                      ? 'Latência de rede até o servidor — distância geográfica.'
                      : 'Resolver DNS lento — trocar resolver pode ajudar.'}
            </p>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
