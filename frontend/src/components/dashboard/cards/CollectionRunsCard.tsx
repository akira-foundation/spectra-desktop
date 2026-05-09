import { Layers } from 'lucide-react'

interface RunSummary {
  collection: { id: string; name: string }
  run: {
    passCount?: number
    failCount?: number
    skipCount?: number
    durationMs?: number
    startedAt?: number
  } | null | undefined
}

export function CollectionRunsCard({ runs, onOpen }: { runs: RunSummary[]; onOpen: () => void }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card/30 p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-1.5">
          <Layers className="w-3 h-3 text-muted-foreground" />
          <h3 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Recent collection runs
          </h3>
        </div>
        <button
          type="button"
          onClick={onOpen}
          className="text-[10.5px] text-muted-foreground hover:text-foreground"
        >
          Open
        </button>
      </div>
      {runs.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground/70 text-center py-6">
          No collection runs yet.
        </p>
      ) : (
        <ul className="m-0 p-0 list-none space-y-1.5">
          {runs.map(({ collection, run }) => {
            const total = (run?.passCount ?? 0) + (run?.failCount ?? 0) + (run?.skipCount ?? 0)
            const passPct = total > 0 ? ((run?.passCount ?? 0) / total) * 100 : 0
            return (
              <li key={collection.id} className="space-y-1">
                <div className="flex items-center justify-between gap-2 text-[11px]">
                  <span className="font-medium truncate flex-1">{collection.name}</span>
                  <span className="text-[10px] font-mono tabular-nums text-muted-foreground shrink-0">
                    {run?.passCount ?? 0}/{total}
                  </span>
                  <span className="text-[10px] text-muted-foreground/70 shrink-0">
                    {run?.durationMs ?? 0}ms
                  </span>
                </div>
                <div className="flex h-1.5 rounded-full overflow-hidden bg-muted/40">
                  <div className="bg-emerald-500/70" style={{ width: `${passPct}%` }} />
                  <div className="bg-rose-500/70" style={{ width: `${100 - passPct}%` }} />
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
