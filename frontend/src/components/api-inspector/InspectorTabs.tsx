import { X } from 'lucide-react'
import { useUIStore, type InspectorTab } from '@/store/uiStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useProjectStore } from '@/store/projectStore'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import type { ScannedEndpoint } from '@/services/scannerService'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []
const EMPTY_TABS: InspectorTab[] = []

export function InspectorTabs() {
  const projectId = useProjectStore((s) => s.activeProjectId)
  const tabs = useUIStore((s) =>
    projectId ? s.inspectorTabsByProject[projectId] ?? EMPTY_TABS : EMPTY_TABS,
  )
  const activeTabId = useUIStore((s) =>
    projectId ? s.activeInspectorTabByProject[projectId] ?? null : null,
  )
  const endpoints = useEndpointsStore((s) =>
    projectId ? s.byProject[projectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
  )
  const setActive = useUIStore((s) => s.setActiveInspectorTab)
  const closeTab = useUIStore((s) => s.closeInspectorTab)
  const { getMethodColor } = useHttpMethod()

  if (!projectId || tabs.length === 0) return null

  const epMap = new Map<string, ScannedEndpoint>()
  for (const e of endpoints) epMap.set(e.id, e)

  return (
    <div className="flex items-end gap-px border-b border-border/40 bg-card/40 px-1 pt-1 overflow-x-auto shrink-0 scrollbar-hairline">
      {tabs.map((tab) => {
        const ep = epMap.get(tab.endpointId)
        const active = tab.id === activeTabId
        const method = ep?.method ?? '—'
        const path = ep?.path ?? '(missing)'
        return (
          <div
            key={tab.id}
            className={cn(
              'group relative flex items-center gap-1.5 h-7 pl-2 pr-1 rounded-t-md text-[11px] cursor-pointer max-w-[260px] min-w-0 shrink-0 transition-colors',
              active
                ? 'bg-card border-x border-t border-border/60 text-foreground'
                : 'border-x border-t border-transparent hover:bg-accent/40 text-muted-foreground',
            )}
            onClick={() => setActive(projectId, tab.id)}
            onAuxClick={(e) => {
              if (e.button === 1) {
                e.preventDefault()
                closeTab(projectId, tab.id)
              }
            }}
            title={`${method} ${path}`}
          >
            <span
              className={cn(
                'inline-flex shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
                getMethodColor(method),
              )}
            >
              {method}
            </span>
            <span className="font-mono truncate flex-1 min-w-0">{path}</span>
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                closeTab(projectId, tab.id)
              }}
              className={cn(
                'inline-flex h-4 w-4 items-center justify-center rounded shrink-0 transition-opacity',
                active
                  ? 'opacity-60 hover:opacity-100 hover:bg-accent/60'
                  : 'opacity-0 group-hover:opacity-100 hover:bg-accent/60',
              )}
              aria-label="Close tab"
            >
              <X className="w-3 h-3" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
