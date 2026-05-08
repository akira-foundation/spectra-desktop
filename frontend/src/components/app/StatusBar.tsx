import { useEffect } from 'react'
import { Folder } from 'lucide-react'
import { useProjectStore } from '@/store/projectStore'
import { useStatsStore } from '@/store/statsStore'

export function StatusBar() {
  const projects = useProjectStore((s) => s.projects)
  const activeId = useProjectStore((s) => s.activeProjectId)
  const active = projects.find((p) => p.id === activeId)
  const lastSyncTime = active?.lastSyncTime ?? null
  const statsReport = useStatsStore((s) =>
    activeId ? s.reportByProject[activeId] ?? null : null,
  )
  const loadReport = useStatsStore((s) => s.loadReport)
  useEffect(() => {
    if (activeId) void loadReport(activeId)
  }, [activeId, loadReport])
  const statCards = (statsReport?.cards ?? []).filter((c) => c.value > 0).slice(0, 5)

  const formatTime = (date: Date | null | undefined) => {
    if (!date) return 'never'
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 1) return 'just now'
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }

  return (
    <footer className="h-6 shrink-0 flex items-center justify-between px-3 text-[10.5px] bg-[#e5e5e5] dark:bg-transparent text-foreground/70 dark:text-white/70 select-none gap-3">
      <div className="flex items-center gap-3 shrink-0">
        <span>Last scan · {formatTime(lastSyncTime)}</span>
        {statCards.length > 0 && (
          <>
            <span className="opacity-50">·</span>
            <span className="flex items-center gap-2">
              {statCards.map((c) => (
                <span key={c.key} className="tabular-nums">
                  <span className="font-medium">{c.value}</span>
                  <span className="opacity-60 ml-0.5">{c.label.toLowerCase()}</span>
                </span>
              ))}
            </span>
          </>
        )}
      </div>
      <div className="flex items-center gap-3 min-w-0 ml-auto">
        {active && (
          <span
            className="flex items-center gap-1.5 min-w-0 max-w-[480px]"
            title={active.path}
          >
            <Folder className="w-2.5 h-2.5 text-muted-foreground dark:text-white/50 shrink-0" />
            <span className="font-mono truncate text-foreground/75 dark:text-white/75">{active.path}</span>
          </span>
        )}
        <span className="opacity-60 font-mono shrink-0">v0.1.0</span>
      </div>
    </footer>
  )
}
