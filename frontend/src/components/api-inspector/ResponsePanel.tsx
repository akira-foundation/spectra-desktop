import { useEffect, useState } from 'react'
import { Download, Trash2 } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { useHistoryStore } from '@/store/historyStore'
import type { HistoryListItem } from '@/services/historyService'
import { useProjectStore } from '@/store/projectStore'
import { useUIStore } from '@/store/uiStore'
import { formatBody } from '@/lib/format'
import {
  TimelineStrip,
  ExceptionPanel,
  HeadersList,
  CopyButton,
  SaveResponseButton,
  ResponseBodyView,
  HistoryRow,
  HistoryDetailView,
  type TimelineData,
} from './response'

const EMPTY_HISTORY: HistoryListItem[] = []

interface ResponsePanelProps {
  responseData: any
  responseHeaders?: Record<string, string[]>
  onReplay?: (entryId: string) => void
  endpointId?: string
  endpointMethod?: string
  endpointPath?: string
  responseStatus?: number
  responseTimeline?: TimelineData | null
}

export function ResponsePanel({
  responseData,
  responseHeaders,
  onReplay,
  endpointId,
  endpointMethod,
  endpointPath,
  responseStatus,
  responseTimeline,
}: ResponsePanelProps) {
  const formatted = formatBody(responseData)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const allHistory = useHistoryStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_HISTORY : EMPTY_HISTORY,
  )
  const history = endpointId ? allHistory.filter((h) => h.endpointID === endpointId) : allHistory
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [tab, setTab] = useState<string>('json')
  const inspectorPending = useUIStore((s) => s.inspectorPending)
  const setInspectorPending = useUIStore((s) => s.setInspectorPending)

  useEffect(() => {
    setExpandedId(null)
  }, [endpointId, activeProjectId])

  useEffect(() => {
    if (!inspectorPending || !endpointId) return
    if (inspectorPending.endpointId !== endpointId) return
    if (inspectorPending.openHistoryLatest) {
      setTab('history')
      const latest = history[0]
      if (latest) setExpandedId(latest.id)
    }
    setInspectorPending(null)
  }, [inspectorPending, endpointId, history])

  const clearHistory = useHistoryStore((s) => s.clear)
  const loadHistory = useHistoryStore((s) => s.load)

  useEffect(() => {
    if (activeProjectId) void loadHistory(activeProjectId)
  }, [activeProjectId, loadHistory])

  return (
    <div className="flex flex-col min-w-0 min-h-0 h-full bg-transparent">
      <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
        <div className="flex items-center gap-1.5">
          <Download className="w-3.5 h-3.5 text-muted-foreground" />
          <h3 className="text-[11.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            Response
          </h3>
        </div>
        <div className="flex items-center gap-1">
          <SaveResponseButton text={formatted} method={endpointMethod} path={endpointPath} />
          <CopyButton text={formatted} title="Copy response" />
        </div>
      </div>

      {responseTimeline && <TimelineStrip timeline={responseTimeline} />}
      {responseStatus !== undefined && responseStatus >= 400 && responseData != null && (
        <ExceptionPanel projectId={activeProjectId ?? null} body={formatted} status={responseStatus} />
      )}

      <Tabs value={tab} onValueChange={setTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {[
            { v: 'json', label: 'JSON' },
            { v: 'headers', label: 'Headers' },
            { v: 'history', label: `History${history.length > 0 ? ` · ${history.length}` : ''}` },
          ].map((t) => (
            <TabsTrigger
              key={t.v}
              value={t.v}
              className="text-[11.5px] px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
            >
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="json" className="flex-1 min-h-0 p-3 overflow-hidden mt-0">
          {responseData == null ? (
            <p className="h-full flex items-center justify-center text-[11.5px] text-muted-foreground italic">
              No response yet
            </p>
          ) : (
            <ResponseBodyView raw={formatted} />
          )}
        </TabsContent>

        <TabsContent value="headers" className="flex-1 p-3 overflow-auto mt-0">
          <HeadersList headers={responseHeaders} />
        </TabsContent>

        <TabsContent value="history" className="flex-1 min-h-0 mt-0 overflow-hidden flex flex-col">
          {expandedId ? (
            <HistoryDetailView
              entryId={expandedId}
              entry={history.find((h) => h.id === expandedId)}
              onBack={() => setExpandedId(null)}
              onReplay={onReplay}
            />
          ) : (
            <>
              <div className="px-3 py-1.5 border-b border-border/40 flex items-center justify-between">
                <span className="text-[10.5px] text-muted-foreground">{history.length} runs</span>
                <Button
                  size="icon-sm"
                  variant="ghost"
                  className="h-6 w-6 text-muted-foreground hover:text-destructive"
                  onClick={() => activeProjectId && clearHistory(activeProjectId)}
                  title="Clear history"
                  disabled={history.length === 0}
                >
                  <Trash2 className="w-3 h-3" />
                </Button>
              </div>
              <div className="flex-1 overflow-auto">
                {history.length === 0 ? (
                  <p className="p-4 text-center text-[11.5px] text-muted-foreground italic">
                    No requests yet. Execute a request to see history.
                  </p>
                ) : (
                  <ul className="space-y-px p-1">
                    {history.map((entry) => (
                      <HistoryRow
                        key={entry.id}
                        entry={entry}
                        onReplay={onReplay}
                        onOpen={() => setExpandedId(entry.id)}
                      />
                    ))}
                  </ul>
                )}
              </div>
            </>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
