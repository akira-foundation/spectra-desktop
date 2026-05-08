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

const sampleRequestBody = {
  name: 'Bob Wilson',
  email: 'user2059@example.com',
  password: 'password205',
  password_confirmation: 'password205',
}

const sampleResponseData = {
  id: 63,
  name: 'John Doe',
  email: 'user1379@example.com',
  email_verified_at: null,
  created_at: '2025-11-30T18:27:03+00:00',
  updated_at: '2025-11-30T18:27:03+00:00',
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
  const [hasResponse, setHasResponse] = useState(false)
  const [infoOpen, setInfoOpen] = useState(false)

  useEffect(() => {
    if (activeProjectId && status === 'idle') {
      void load(activeProjectId)
    }
  }, [activeProjectId, status, load])

  useEffect(() => {
    setSelectedTag(null)
    setRouteValues([])
    setQueryParams([])
    setHasResponse(false)
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
    setHasResponse(false)
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
    if (!selected) return
    setHasResponse(true)
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

        {!selected ? (
          <EndpointEmptyState />
        ) : (
          <>
            <EndpointHeader
              method={selected.method}
              path={resolvedPath}
              statusCode={hasResponse ? 201 : 0}
              responseTime={hasResponse ? '260ms' : '—'}
              responseSize={hasResponse ? '0.16KB' : '—'}
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
              />
              {hasResponse ? (
                <ResponsePanel responseData={sampleResponseData} />
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
