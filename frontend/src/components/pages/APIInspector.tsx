import { useEffect, useMemo, useRef, useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useAuthStore } from '@/store/authStore'
import { useHistoryStore } from '@/store/historyStore'
import { useEnvironmentStore } from '@/store/environmentStore'
import { historyService } from '@/services/historyService'
import toast from 'react-hot-toast'
import {
  EndpointList,
  AuthenticationDrawer,
  RequestPanel,
  ResponsePanel,
  EndpointHeader,
} from '@/components/api-inspector'
import { EndpointInfoSheet } from '@/components/api-inspector/EndpointInfoSheet'
import { EndpointEmptyState } from '@/components/api-inspector/EndpointEmptyState'
import { ResponseEmptyState } from '@/components/api-inspector/ResponseEmptyState'
import {
  ScanErrorState,
  NoRoutesState,
} from '@/components/api-inspector/ScanStates'
import { EndpointListSkeleton } from '@/components/api-inspector/EndpointListSkeleton'
import { Skeleton } from '@/components/ui/skeleton'
import { groupEndpoints, type GroupedEndpoint, PINNED_CATEGORY } from '@/lib/group-endpoints'
import type { ScannedEndpoint } from '@/services/scannerService'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []
import {
  buildQueryString,
  extractRouteParams,
  resolveRoutePath,
  type QueryParam,
} from '@/lib/route-params'
import type { HeaderRow } from '@/components/api-inspector/HeadersEditor'
import { BaseURLBar } from '@/components/api-inspector/BaseURLBar'
import { useRequestRunner } from '@/hooks/useRequestRunner'
import {
  buildExampleBody,
  parseRequestSchema,
  type RequestSchema,
} from '@/lib/request-schema'

export function APIInspector() {
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const load = useEndpointsStore((s) => s.load)
  const scan = useEndpointsStore((s) => s.scan)
  const allEndpoints = useEndpointsStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? EMPTY_ENDPOINTS : EMPTY_ENDPOINTS,
  )
  const status = useEndpointsStore((s) =>
    activeProjectId ? s.status[activeProjectId] ?? 'idle' : 'idle',
  )
  const error = useEndpointsStore((s) =>
    activeProjectId ? s.errors[activeProjectId] ?? null : null,
  )

  const activeAuthMethod = useUIStore((s) => s.activeAuthMethod)
  const setActiveAuthMethod = useUIStore((s) => s.setActiveAuthMethod)
  const persistedTag = useUIStore((s) =>
    activeProjectId ? s.selectedEndpointByProject[activeProjectId] ?? null : null,
  )
  const setSelectedEndpoint = useUIStore((s) => s.setSelectedEndpoint)
  const persistedBodies = useUIStore((s) => s.requestBodyByEndpoint)
  const persistedHeaders = useUIStore((s) => s.requestHeadersByEndpoint)
  const persistBody = useUIStore((s) => s.setRequestBody)
  const persistHeaders = useUIStore((s) => s.setRequestHeaders)
  const loadAuth = useAuthStore((s) => s.load)
  const refreshAuth = useAuthStore((s) => s.refresh)

  useEffect(() => {
    if (activeProjectId) void loadAuth(activeProjectId)
  }, [activeProjectId, loadAuth])

  const pinnedKeys = useUIStore((s) =>
    activeProjectId ? s.pinnedEndpointsByProject[activeProjectId] ?? null : null,
  )
  const groupOrder = useUIStore((s) =>
    activeProjectId ? s.groupOrderByProject[activeProjectId] ?? null : null,
  )
  const togglePinnedEndpoint = useUIStore((s) => s.togglePinnedEndpoint)
  const setGroupOrder = useUIStore((s) => s.setGroupOrder)

  const project = useProjectStore((s) => s.projects.find((p) => p.id === activeProjectId))
  const envs = useEnvironmentStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const activeEnv = envs?.find((e) => e.id === project?.activeEnvironmentId) ?? null
  const variableNames = useMemo<Record<string, string>>(
    () => activeEnv?.vars ?? {},
    [activeEnv],
  )

  const groups = useMemo(
    () =>
      groupEndpoints(allEndpoints, {
        pinnedKeys: pinnedKeys ?? [],
        groupOrder: groupOrder ?? [],
      }),
    [allEndpoints, pinnedKeys, groupOrder],
  )

  const [selectedTag, setSelectedTag] = useState<string | null>(null)
  const [routeValues, setRouteValues] = useState<string[]>([])
  const [queryParams, setQueryParams] = useState<QueryParam[]>([])
  const [infoOpen, setInfoOpen] = useState(false)
  const [requestBody, setRequestBody] = useState<string>('')
  const [bodyTouched, setBodyTouched] = useState(false)
  const [headers, setHeaders] = useState<HeaderRow[]>([])
  const runner = useRequestRunner()

  useEffect(() => {
    if (activeProjectId && status === 'idle') {
      void load(activeProjectId)
    }
  }, [activeProjectId, status, load])

  useEffect(() => {
    setSelectedTag(persistedTag)
    setRouteValues([])
    setQueryParams([])
    runner.reset()
  }, [activeProjectId])

  const decoratedGroups = useMemo(
    () =>
      groups.map((g) => ({
        ...g,
        items: g.items.map((item) => ({ ...item, active: item.tag === selectedTag })),
      })),
    [groups, selectedTag],
  )

  const selected: GroupedEndpoint | null = useMemo(() => {
    if (!selectedTag) return null
    for (const g of groups) {
      const found = g.items.find((i) => i.tag === selectedTag)
      if (found) return found
    }
    return null
  }, [groups, selectedTag])

  const routeParams = useMemo(
    () => (selected ? extractRouteParams(selected.path) : []),
    [selected],
  )

  const rawSchema = useMemo(() => {
    if (!selectedTag) return null
    return allEndpoints.find((e) => e.id === selectedTag)?.requestSchema ?? null
  }, [selectedTag, allEndpoints])

  const requestSchema: RequestSchema | null = useMemo(
    () => parseRequestSchema(rawSchema),
    [rawSchema],
  )

  const selectedRaw = useMemo(
    () => (selectedTag ? allEndpoints.find((e) => e.id === selectedTag) ?? null : null),
    [selectedTag, allEndpoints],
  )

  useEffect(() => {
    setRouteValues(routeParams.map(() => ''))
    setQueryParams([])
    runner.reset()
    if (!selectedTag) {
      setRequestBody('')
      setHeaders([])
      setBodyTouched(false)
      return
    }
    const stored = persistedBodies[selectedTag]
    if (stored !== undefined) {
      setRequestBody(stored)
      setBodyTouched(true)
    } else {
      const parsed = parseRequestSchema(rawSchema)
      if (parsed && parsed.fields.length > 0) {
        setRequestBody(JSON.stringify(buildExampleBody(parsed.fields), null, 2))
      } else {
        setRequestBody('')
      }
      setBodyTouched(false)
    }
    setHeaders(persistedHeaders[selectedTag] ?? [])
  }, [selectedTag, rawSchema])

  const persistedBodyForSelected = selectedTag ? persistedBodies[selectedTag] : undefined
  const persistedHeadersForSelected = selectedTag ? persistedHeaders[selectedTag] : undefined
  useEffect(() => {
    if (persistedBodyForSelected !== undefined) {
      setRequestBody(persistedBodyForSelected)
    }
  }, [persistedBodyForSelected])
  useEffect(() => {
    if (persistedHeadersForSelected !== undefined) {
      setHeaders(persistedHeadersForSelected)
    }
  }, [persistedHeadersForSelected])

  const resolvedPath = useMemo(() => {
    if (!selected) return ''
    return resolveRoutePath(selected.path, routeValues) + buildQueryString(queryParams)
  }, [selected, routeValues, queryParams])

  const handleSelect = (tag: string) => {
    setSelectedTag(tag)
    if (activeProjectId) setSelectedEndpoint(activeProjectId, tag)
  }
  const handleRetry = () => {
    if (activeProjectId) void scan(activeProjectId)
  }
  const handleRouteValueChange = (index: number, value: string) => {
    setRouteValues((prev) => prev.map((v, i) => (i === index ? value : v)))
  }
  const handleQueryAdd = () => {
    setQueryParams((prev) => [...prev, { key: '', value: '' }])
  }
  const handleQueryChange = (index: number, patch: Partial<QueryParam>) => {
    setQueryParams((prev) => prev.map((p, i) => (i === index ? { ...p, ...patch } : p)))
  }
  const handleQueryRemove = (index: number) => {
    setQueryParams((prev) => prev.filter((_, i) => i !== index))
  }
  const handleHeaderAdd = () => {
    setHeaders((prev) => {
      const next = [...prev, { key: '', value: '', enabled: true }]
      if (selectedTag) persistHeaders(selectedTag, next)
      return next
    })
  }
  const handleHeaderChange = (index: number, patch: Partial<HeaderRow>) => {
    setHeaders((prev) => {
      const next = prev.map((h, i) => (i === index ? { ...h, ...patch } : h))
      if (selectedTag) persistHeaders(selectedTag, next)
      return next
    })
  }
  const handleHeaderRemove = (index: number) => {
    setHeaders((prev) => {
      const next = prev.filter((_, i) => i !== index)
      if (selectedTag) persistHeaders(selectedTag, next)
      return next
    })
  }
  const handleBodyChange = (value: string) => {
    setRequestBody(value)
    setBodyTouched(true)
    if (selectedTag) persistBody(selectedTag, value)
  }

  const handleResetBody = () => {
    if (requestSchema && requestSchema.fields.length > 0) {
      setRequestBody(JSON.stringify(buildExampleBody(requestSchema.fields), null, 2))
    } else {
      setRequestBody('')
    }
    setBodyTouched(false)
  }

  const handleExecute = () => {
    if (!selected || !activeProjectId) return
    const headerMap: Record<string, string> = {}
    for (const h of headers) {
      if (!h.enabled) continue
      const k = h.key.trim()
      if (!k) continue
      headerMap[k] = h.value
    }
    const prevToken = activeProjectId
      ? useAuthStore.getState().byProject[activeProjectId]?.tokenPreview ?? null
      : null
    void runner
      .execute({
        projectID: activeProjectId,
        endpointID: selected.id,
        method: selected.method,
        path: resolvedPath,
        headers: headerMap,
        body: requestBody,
      })
      .then(async () => {
        if (!activeProjectId) return
        await refreshAuth(activeProjectId)
        await useHistoryStore.getState().refresh(activeProjectId)
        const next = useAuthStore.getState().byProject[activeProjectId]
        if (next?.tokenPreview && next.tokenPreview !== prevToken) {
          const who = next.user?.name || next.user?.username || next.user?.email || 'user'
          toast.success(`Token captured · ${who}`)
        }
      })
  }

  const handleReplay = async (entryId: string) => {
    if (!selected || !activeProjectId) return
    try {
      const detail = await historyService.get(entryId)
      if (!detail) return

      let headerMap: Record<string, string> = {}
      try {
        headerMap = JSON.parse(detail.requestHeaders || '{}') as Record<string, string>
      } catch {}

      setRequestBody(detail.requestBody ?? '')
      setBodyTouched(true)
      setHeaders(
        Object.entries(headerMap).map(([key, value]) => ({ key, value, enabled: true })),
      )

      await runner.execute({
        projectID: activeProjectId,
        endpointID: selected.id,
        method: detail.method || selected.method,
        path: resolvedPath,
        headers: headerMap,
        body: detail.requestBody ?? '',
      })
      await useHistoryStore.getState().refresh(activeProjectId)
      toast.success('Replayed')
    } catch (err) {
      console.error('replay failed:', err)
      toast.error('Replay failed')
    }
  }

  const executeRef = useRef(handleExecute)
  useEffect(() => {
    executeRef.current = handleExecute
  })

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const isEnter = e.key === 'Enter' || e.code === 'Enter' || e.keyCode === 13
      if ((e.metaKey || e.ctrlKey) && isEnter) {
        e.preventDefault()
        e.stopPropagation()
        executeRef.current()
      }
    }
    document.addEventListener('keydown', onKey, true)
    return () => document.removeEventListener('keydown', onKey, true)
  }, [])

  if (status === 'loading' || status === 'scanning') {
    return (
      <div className="h-full flex gap-2 p-2 overflow-hidden">
        <div className="w-64 shrink-0 rounded-md border border-border/40 bg-foreground/[0.025] dark:bg-white/[0.02]">
          <EndpointListSkeleton />
        </div>
        <div className="flex-1 rounded-md border border-border/40 bg-card/30" />
      </div>
    )
  }
  if (status === 'error' && error) {
    return <CenterPane><ScanErrorState error={error} onRetry={handleRetry} /></CenterPane>
  }
  if (status === 'empty' || (status === 'ready' && groups.length === 0)) {
    return <CenterPane><NoRoutesState onRetry={handleRetry} /></CenterPane>
  }

  return (
    <div className="h-full flex gap-2 p-2 overflow-hidden">
      <EndpointList
        endpoints={decoratedGroups}
        onSelectEndpoint={handleSelect}
        pinnedKeys={pinnedKeys ?? []}
        onTogglePin={(key) => activeProjectId && togglePinnedEndpoint(activeProjectId, key)}
        onReorder={(order) => activeProjectId && setGroupOrder(activeProjectId, order)}
      />

      <div className="flex-1 flex flex-col overflow-hidden rounded-md border border-border/40 bg-card/30">
        <AuthenticationDrawer
          activeMethod={activeAuthMethod}
          onMethodChange={setActiveAuthMethod}
        />

        <BaseURLBar />

        {!selected ? (
          <EndpointEmptyState />
        ) : (
          <>
            <EndpointHeader
              method={selected.method}
              path={resolvedPath}
              statusCode={runner.response?.status ?? 0}
              responseTime={runner.response ? `${runner.response.durationMs}ms` : '—'}
              responseSize={runner.response ? formatSize(runner.response.sizeBytes) : '—'}
              onInfoClick={
                hasMetadata(selected) || (requestSchema && requestSchema.fields.length > 0)
                  ? () => setInfoOpen(true)
                  : undefined
              }
            />

            <div className="flex-1 grid grid-cols-2 overflow-hidden">
              <RequestPanel
                requestBody={requestBody}
                onRequestBodyChange={handleBodyChange}
                onResetBody={handleResetBody}
                bodyTouched={bodyTouched}
                schema={requestSchema}
                variables={variableNames}
                routeParams={routeParams}
                routeValues={routeValues}
                onRouteValueChange={handleRouteValueChange}
                queryParams={queryParams}
                onQueryAdd={handleQueryAdd}
                onQueryChange={handleQueryChange}
                onQueryRemove={handleQueryRemove}
                headers={headers}
                onHeaderAdd={handleHeaderAdd}
                onHeaderChange={handleHeaderChange}
                onHeaderRemove={handleHeaderRemove}
                onExecute={handleExecute}
                executing={runner.loading}
              />
              {runner.loading ? (
                <ResponseLoadingSkeleton />
              ) : runner.error ? (
                <div className="bg-transparent flex items-center justify-center">
                  <RunnerErrorBlock code={runner.error.code} message={runner.error.message} />
                </div>
              ) : runner.response ? (
                <ResponsePanel
                  responseData={parseBody(runner.response.body ?? '')}
                  responseHeaders={runner.response.headers as Record<string, string[]>}
                  onReplay={handleReplay}
                  endpointId={selected?.id}
                />
              ) : (
                <ResponsePanel
                  responseData={null}
                  responseHeaders={undefined}
                  onReplay={handleReplay}
                  endpointId={selected?.id}
                />
              )}
            </div>
          </>
        )}
      </div>

      {selected && (
        <EndpointInfoSheet
          open={infoOpen}
          onOpenChange={setInfoOpen}
          method={selected.method}
          path={selected.path}
          controller={selected.controller}
          middleware={selected.middleware}
          authRequired={selected.authRequired}
          schema={requestSchema}
        />
      )}
    </div>
  )
}

function CenterPane({ children }: { children: React.ReactNode }) {
  return <div className="h-full flex items-center justify-center">{children}</div>
}

function ResponseLoadingSkeleton() {
  return (
    <div className="flex flex-col min-w-0 min-h-0 h-full bg-transparent p-3 gap-2">
      <Skeleton className="h-5 w-24" />
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="h-4 w-2/3" />
      <Skeleton className="flex-1 min-h-32" />
    </div>
  )
}

function hasMetadata(endpoint: GroupedEndpoint): boolean {
  return Boolean(
    endpoint.controller ||
      (endpoint.middleware && endpoint.middleware.length > 0) ||
      endpoint.authRequired !== undefined,
  )
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)}KB`
  return `${(bytes / 1024 / 1024).toFixed(2)}MB`
}

function parseBody(body: string): unknown {
  if (!body) return ''
  try {
    return JSON.parse(body)
  } catch {
    return body
  }
}

interface RunnerErrorBlockProps {
  code?: string
  message: string
}

function RunnerErrorBlock({ code, message }: RunnerErrorBlockProps) {
  const copy = errorCopy(code)
  return (
    <div className="max-w-sm w-full text-center space-y-2 px-6">
      <div className="inline-flex w-9 h-9 items-center justify-center rounded-lg bg-destructive/15 text-destructive">
        <span className="text-base">!</span>
      </div>
      <div className="space-y-1">
        <p className="text-[13px] font-semibold tracking-tight">{copy.title}</p>
        <p className="text-[12px] text-muted-foreground leading-relaxed">{copy.description}</p>
      </div>
      <code className="block text-[10.5px] font-mono text-muted-foreground/80 break-all">
        {message}
      </code>
    </div>
  )
}

function errorCopy(code?: string): { title: string; description: string } {
  switch (code) {
    case 'connection_refused':
      return {
        title: 'Connection refused',
        description: 'The server is not reachable. Is your local server running?',
      }
    case 'timeout':
      return {
        title: 'Request timed out',
        description: 'The server took too long to respond.',
      }
    case 'dns':
      return {
        title: 'DNS lookup failed',
        description: 'Could not resolve the host. Check the base URL.',
      }
    case 'invalid_url':
      return {
        title: 'Invalid URL',
        description: 'The base URL or path is malformed.',
      }
    case 'tls':
      return {
        title: 'TLS handshake failed',
        description: 'Unable to establish a secure connection.',
      }
    case 'missing_base_url':
      return {
        title: 'Base URL not set',
        description: 'Set the workspace base URL above and try again.',
      }
    default:
      return {
        title: 'Request failed',
        description: 'Spectra could not complete the request.',
      }
  }
}
