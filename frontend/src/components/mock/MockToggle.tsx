import { Server } from 'lucide-react'
import { useMockStore } from '@/store/mockStore'
import { cn } from '@/lib/utils'

interface Props {
  projectId: string
}

export function MockToggle({ projectId }: Props) {
  const enabled = useMockStore((s) => !!s.useMockByProject[projectId])
  const running = useMockStore((s) => s.status.running)
  const url = useMockStore((s) => s.status.url)
  const setUseMock = useMockStore((s) => s.setUseMockForProject)

  if (!running) return null

  const intercepting = enabled

  return (
    <button
      type="button"
      onClick={() => setUseMock(projectId, !enabled)}
      title={
        intercepting
          ? `Routing requests to ${url}`
          : 'Click to route this project to the mock server'
      }
      className={cn(
        'shrink-0 h-7 inline-flex items-center gap-1.5 px-2 rounded-md text-[11px] border transition-colors',
        intercepting
          ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
          : 'border-border/50 hover:bg-accent/40 text-muted-foreground',
      )}
    >
      <Server className="w-3 h-3" />
      {intercepting && (
        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
      )}
    </button>
  )
}
