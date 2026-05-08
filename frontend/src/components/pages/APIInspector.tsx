import { useEffect, useMemo, useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
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
  ScanLoadingState,
  ScanErrorState,
  NoRoutesState,
} from '@/components/api-inspector/ScanStates'
import { groupEndpoints, type GroupedEndpoint } from '@/lib/group-endpoints'
import type { ScannedEndpoint } from '@/services/scannerService'

const EMPTY_ENDPOINTS: ScannedEndpoint[] = []
import {
  buildQueryString,
  extractRouteParams,
  resolveRoutePath,
  type QueryParam,
} from '@/lib/route-params'
import { BaseURLBar } from '@/components/api-inspector/BaseURLBar'
import { useRequestRunner } from '@/hooks/useRequestRunner'

const sampleRequestBody = {
  name: 'Bob Wilson',
  email: 'user2059@example.com',
  password: 'password205',
  password_confirmation: 'password205',
}

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

  const groups = useMemo(() => groupEndpoints(allEndpoints), [allEndpoints])

  const [selectedTag, setSelectedTag] = useState<string | null>(null)
  const [routeValues, setRouteValues] = useState<string[]>([])
  const [queryParams, setQueryParams] = useState<QueryParam[]>([])
  const [infoOpen, setInfoOpen] = useState(false)
  const runner = useRequestRunner()

  useEffect(() => {
    if (activeProjectId && status === 'idle') {
      void load(activeProjectId)
    }
  }, [activeProjectId, status, load])

  useEffect(() => {
    setSelectedTag(null)
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

  useEffect(() => {
    setRouteValues(routeParams.map(() => ''))
    setQueryParams([])
    runner.reset()
  }, [selectedTag, routeParams.length])

  const resolvedPath = useMemo(() => {
    if (!selected) return ''
    return resolveRoutePath(selected.path, routeValues) + buildQueryString(queryParams)
  }, [selected, routeValues, queryParams])

  const handleSelect = (tag: string) => setSelectedTag(tag)
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
  const handleExecute = () => {
    if (!selected || !activeProjectId) return
    void runner.execute({
      projectID: activeProjectId,
      method: selected.method,
      path: resolvedPath,
      headers: {},
      body: '',
    })
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const isEnter = e.key === 'Enter' || e.code === 'Enter' || e.keyCode === 13
      if ((e.metaKey || e.ctrlKey) && isEnter) {
        e.preventDefault()
        e.stopPropagation()
        handleExecute()
      }
    }
    document.addEventListener('keydown', onKey, true)
    return () => document.removeEventListener('keydown', onKey, true)
  }, [selected])

  if (status === 'loading' || status === 'scanning') {
    return <CenterPane><ScanLoadingState /></CenterPane>
  }
  if (status === 'error' && error) {
    return <CenterPane><ScanErrorState error={error} onRetry={handleRetry} /></CenterPane>
  }
  if (status === 'empty' || (status === 'ready' && groups.length === 0)) {
    return <CenterPane><NoRoutesState onRetry={handleRetry} /></CenterPane>
  }

  return (
    <div className="h-full flex overflow-hidden">
      <EndpointList endpoints={decoratedGroups} onSelectEndpoint={handleSelect} />

      <div className="flex-1 flex flex-col overflow-hidden">
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
              onInfoClick={hasMetadata(selected) ? () => setInfoOpen(true) : undefined}
            />

            <div className="flex-1 grid grid-cols-2 overflow-hidden">
              <RequestPanel
                requestBody={sampleRequestBody}
                routeParams={routeParams}
                routeValues={routeValues}
                onRouteValueChange={handleRouteValueChange}
                queryParams={queryParams}
                onQueryAdd={handleQueryAdd}
                onQueryChange={handleQueryChange}
                onQueryRemove={handleQueryRemove}
                onExecute={handleExecute}
                executing={runner.loading}
              />
              {runner.loading ? (
                <div className="bg-transparent flex items-center justify-center">
                  <ScanLoadingState />
                </div>
              ) : runner.error ? (
                <div className="bg-transparent flex items-center justify-center">
                  <RunnerErrorBlock code={runner.error.code} message={runner.error.message} />
                </div>
              ) : runner.response ? (
                <ResponsePanel responseData={parseBody(runner.response.body ?? '')} />
              ) : (
                <div className="bg-transparent flex items-center justify-center">
                  <ResponseEmptyState />
                </div>
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
        />
      )}
    </div>
  )
}

function CenterPane({ children }: { children: React.ReactNode }) {
  return <div className="h-full flex items-center justify-center">{children}</div>
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
