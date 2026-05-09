import { cn } from '@/lib/utils'
import { formatLeaf, valueTone, type TableRows } from '@/lib/format'

export function TableView({ rows }: { rows: TableRows }) {
  return (
    <div className="overflow-auto rounded-md border border-border/40">
      <table className="w-full text-[11px] font-mono">
        <thead className="sticky top-0 bg-muted/50">
          <tr>
            {rows.columns.map((c) => (
              <th
                key={c}
                className="px-2 py-1.5 text-left text-[10px] uppercase tracking-wider font-semibold text-muted-foreground/80 border-b border-border/40"
              >
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.data.map((r, i) => (
            <tr key={i} className="border-b border-border/20 hover:bg-accent/20">
              {r.map((v, j) => (
                <td key={j} className={cn('px-2 py-1 align-top truncate max-w-[300px]', valueTone(v))}>
                  {formatLeaf(v)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
