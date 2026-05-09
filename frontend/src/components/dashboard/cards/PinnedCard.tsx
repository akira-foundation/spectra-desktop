import { Star } from 'lucide-react'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { Card } from './Card'

interface Props {
  endpoints: Array<{ id: string; method: string; path: string }>
  onOpen: (id: string) => void
}

export function PinnedCard({ endpoints, onOpen }: Props) {
  const { getMethodColor } = useHttpMethod()
  return (
    <Card title="Pinned" icon={Star}>
      {endpoints.length === 0 ? (
        <p className="text-[11.5px] italic text-muted-foreground">
          Pin endpoints from the inspector for quick access.
        </p>
      ) : (
        <ul className="space-y-px">
          {endpoints.map((ep) => (
            <li key={ep.id}>
              <button
                type="button"
                onClick={() => onOpen(ep.id)}
                className="w-full flex items-center gap-2 px-1 py-1 rounded hover:bg-accent/40 transition-colors"
              >
                <span
                  className={cn(
                    'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                    getMethodColor(ep.method),
                  )}
                >
                  {ep.method}
                </span>
                <span className="text-[11.5px] font-mono truncate flex-1 text-left text-foreground/85">
                  {ep.path}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
