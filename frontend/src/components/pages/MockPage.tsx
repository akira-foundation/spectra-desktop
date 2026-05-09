import { useEffect, useMemo, useState } from 'react'
import { Server } from 'lucide-react'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useMockStore } from '@/store/mockStore'
import { MockHeader } from '@/components/mock/MockHeader'
import { MockLogList } from '@/components/mock/MockLogList'
import { MockEndpointRow } from '@/components/mock/MockEndpointRow'
import { Input } from '@/components/ui/input'
import type { ScannedEndpoint } from '@/services/scannerService'
import type { MockOverride } from '@/services/mockService'
import toast from 'react-hot-toast'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []
const EMPTY_OVERRIDES: MockOverride[] = []

export function MockPage() {
  const projectId = useProjectStore((s) => s.activeProjectId ?? '')
  const project = useProjectStore((s) =>
    projectId ? s.projects.find((p) => p.id === projectId) ?? null : null,
  )
  const endpoints = useEndpointsStore((s) =>
    projectId ? s.byProject[projectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
  )
  const status = useMockStore((s) => s.status)
  const logs = useMockStore((s) => s.logs)
  const overrides = useMockStore((s) =>
    projectId ? s.overridesByProject[projectId] ?? EMPTY_OVERRIDES : EMPTY_OVERRIDES,
  )
  const init = useMockStore((s) => s.init)
  const start = useMockStore((s) => s.start)
  const stop = useMockStore((s) => s.stop)
  const listOverrides = useMockStore((s) => s.listOverrides)
  const saveOverride = useMockStore((s) => s.saveOverride)
  const clearLogs = useMockStore((s) => s.clearLogs)

  const [filter, setFilter] = useState('')

  useEffect(() => {
    void init()
  }, [init])

  useEffect(() => {
    if (projectId) void listOverrides(projectId)
  }, [projectId, listOverrides])

  const overrideByEndpoint = useMemo(() => {
    const map = new Map<string, MockOverride>()
    for (const o of overrides) map.set(o.endpointId, o)
    return map
  }, [overrides])

  const filtered = useMemo(() => {
    if (!filter.trim()) return endpoints
    const q = filter.toLowerCase()
    return endpoints.filter(
      (e) => e.path.toLowerCase().includes(q) || e.method.toLowerCase().includes(q),
    )
  }, [endpoints, filter])

  if (!project) {
    return (
      <div className="flex h-full items-center justify-center text-muted-foreground text-[12px]">
        Select a project first
      </div>
    )
  }

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-4xl mx-auto p-6 space-y-4">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Mock server</h1>
          <p className="text-muted-foreground text-[12.5px] mt-1">
            Serve <span className="font-medium text-foreground/80">{project.name}</span>'s
            endpoints locally. Responses come from past history first, then generated from
            schema, with optional per-endpoint overrides.
          </p>
        </div>

        <div className="rounded-md border border-border/40 px-3.5 py-3">
          <MockHeader
            projectId={projectId}
            status={status}
            onStart={async (port) => {
              try {
                await start(projectId, port)
                toast.success('Mock server started')
              } catch (err) {
                toast.error(err instanceof Error ? err.message : String(err))
              }
            }}
            onStop={async () => {
              await stop()
              toast.success('Mock server stopped')
            }}
            onClearLogs={clearLogs}
          />
        </div>

        {status.running && (
          <div className="rounded-md border border-border/40 overflow-hidden">
            <div className="px-3.5 py-2 border-b border-border/40 flex items-center justify-between">
              <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground font-semibold">
                Live requests
              </span>
              <span className="text-[10.5px] text-muted-foreground tabular-nums">
                {logs.length} entries
              </span>
            </div>
            <MockLogList logs={logs} />
          </div>
        )}

        <div className="rounded-md border border-border/40 overflow-hidden">
          <div className="px-3.5 py-2 border-b border-border/40 flex items-center justify-between gap-3">
            <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground font-semibold">
              Endpoints
            </span>
            <Input
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Filter by method or path"
              className="h-7 max-w-xs text-[12px]"
            />
          </div>
          {filtered.length === 0 ? (
            <p className="px-3.5 py-6 text-center text-[11.5px] italic text-muted-foreground/70">
              <Server className="h-7 w-7 mx-auto mb-2 text-muted-foreground/50" strokeWidth={1.5} />
              No endpoints scanned for this project yet.
            </p>
          ) : (
            <ul>
              {filtered.map((ep) => (
                <MockEndpointRow
                  key={ep.id}
                  endpoint={ep}
                  override={overrideByEndpoint.get(ep.id)}
                  onSave={async (input) => {
                    await saveOverride({
                      projectID: projectId,
                      endpointId: input.endpointId,
                      enabled: input.enabled,
                      status: input.status,
                      latencyMs: input.latencyMs,
                      source: input.source,
                      body: input.body,
                    })
                    toast.success('Mock saved')
                  }}
                />
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}
