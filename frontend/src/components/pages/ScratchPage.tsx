import { useEffect, useMemo, useRef, useState } from 'react'
import { Plus, Play, Trash2, Terminal, Loader2, FileText } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { useProjectStore } from '@/store/projectStore'
import { runnerService } from '@/services/runnerService'
import { scratchService, type ScratchRequest } from '@/services/scratchService'
import { Island } from '@/components/app/Island'
import { BodyEditor } from '@/components/api-inspector/BodyEditor'
import { CurlImportDialog } from '@/components/api-inspector/CurlImportDialog'
import { HARImportDialog } from '@/components/api-inspector/HARImportDialog'
import { HeadersEditor, type HeaderRow } from '@/components/api-inspector/HeadersEditor'
import {
  TimelineStrip,
  ExceptionPanel,
  ResponseBodyView,
  HeadersList,
  CopyButton,
  SaveResponseButton,
} from '@/components/api-inspector/response'
import { formatBody, prettyJSON } from '@/lib/format'
import { cn } from '@/lib/utils'

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
const LEGACY_KEY = 'spectra:scratch-requests'

function newScratch(projectID: string): ScratchRequest {
  return {
    id: '',
    projectID,
    name: 'Untitled',
    method: 'GET',
    url: '',
    headers: [],
    body: '',
    response: null,
    sortOrder: 0,
  }
}

function ImportMenu({ onCurl, onHAR }: { onCurl: (p: any) => void; onHAR: (e: any[]) => void }) {
  const [curlOpen, setCurlOpen] = useState(false)
  const [harOpen, setHarOpen] = useState(false)
  return (
    <>
      <div className="grid grid-cols-2 gap-2">
        <button
          type="button"
          onClick={() => setCurlOpen(true)}
          className="group h-8 inline-flex items-center justify-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[11.5px] font-medium transition-colors gap-1.5"
        >
          <Terminal className="w-3.5 h-3.5 text-emerald-500" />
          curl
        </button>
        <button
          type="button"
          onClick={() => setHarOpen(true)}
          className="group h-8 inline-flex items-center justify-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[11.5px] font-medium transition-colors gap-1.5"
        >
          <FileText className="w-3.5 h-3.5 text-emerald-500" />
          HAR
        </button>
      </div>
      <CurlImportDialog open={curlOpen} onOpenChange={setCurlOpen} onImport={onCurl} />
      <HARImportDialog open={harOpen} onOpenChange={setHarOpen} onImport={onHAR} />
    </>
  )
}

export function ScratchPage() {
  const projectId = useProjectStore((s) => s.activeProjectId)
  const [items, setItems] = useState<ScratchRequest[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const [running, setRunning] = useState(false)
  const saveTimers = useRef<Record<string, ReturnType<typeof setTimeout>>>({})

  useEffect(() => {
    if (!projectId) return
    void (async () => {
      const rows = await scratchService.list(projectId)
      // one-time migration from legacy localStorage if backend empty
      if (rows.length === 0) {
        try {
          const legacy = localStorage.getItem(LEGACY_KEY)
          if (legacy) {
            const parsed = JSON.parse(legacy)
            if (Array.isArray(parsed) && parsed.length > 0) {
              const migrated: ScratchRequest[] = []
              for (const old of parsed) {
                const saved = await scratchService.save({
                  id: '',
                  projectID: projectId,
                  name: old.name ?? 'Untitled',
                  method: old.method ?? 'GET',
                  url: old.url ?? '',
                  headers: old.headers ?? [],
                  body: old.body ?? '',
                  response: old.response ?? null,
                  sortOrder: migrated.length,
                })
                migrated.push(saved)
              }
              localStorage.removeItem(LEGACY_KEY)
              setItems(migrated)
              if (migrated[0]) setActiveId(migrated[0].id)
              return
            }
          }
        } catch {}
      }
      setItems(rows)
    })()
  }, [projectId])

  useEffect(() => {
    if (!activeId && items.length > 0) setActiveId(items[0].id)
    if (activeId && !items.find((i) => i.id === activeId) && items.length > 0) {
      setActiveId(items[0].id)
    }
  }, [items, activeId])

  const active = useMemo(() => items.find((i) => i.id === activeId) ?? null, [items, activeId])

  const scheduleSave = (req: ScratchRequest) => {
    const existing = saveTimers.current[req.id]
    if (existing) clearTimeout(existing)
    saveTimers.current[req.id] = setTimeout(() => {
      void scratchService.save(req)
    }, 400)
  }

  const create = async () => {
    if (!projectId) return
    const draft = { ...newScratch(projectId), sortOrder: items.length }
    const saved = await scratchService.save(draft)
    setItems((prev) => [saved, ...prev])
    setActiveId(saved.id)
  }

  const remove = async (id: string) => {
    setItems((prev) => prev.filter((i) => i.id !== id))
    try {
      await scratchService.remove(id)
    } catch {}
  }

  const update = (id: string, patch: Partial<ScratchRequest>) => {
    setItems((prev) => {
      const next = prev.map((i) => (i.id === id ? { ...i, ...patch } : i))
      const updated = next.find((i) => i.id === id)
      if (updated) scheduleSave(updated)
      return next
    })
  }

  const run = async () => {
    if (!active || !active.url.trim()) return
    setRunning(true)
    try {
      const headers: Record<string, string> = {}
      for (const h of active.headers) {
        if (h.enabled && h.key.trim()) headers[h.key.trim()] = h.value
      }
      const url = new URL(active.url)
      const baseURL = `${url.protocol}//${url.host}`
      const path = url.pathname + url.search
      const result = await runnerService.execute({
        projectID: projectId ?? '',
        method: active.method,
        path,
        baseUrl: baseURL,
        headers,
        body: active.body || undefined,
        skipAuth: true,
      })
      update(active.id, {
        response: {
          status: result.status,
          body: result.body ?? '',
          headers: result.headers ?? {},
          durationMs: result.durationMs ?? 0,
          timeline: result.timeline ?? null,
        },
      })
    } catch (err) {
      update(active.id, {
        response: {
          status: 0,
          body: (err as Error).message ?? 'Request failed',
          headers: {},
          durationMs: 0,
        },
      })
    } finally {
      setRunning(false)
    }
  }

  const importCurl = async (parsed: any) => {
    if (!projectId) return
    const draft: ScratchRequest = {
      id: '',
      projectID: projectId,
      name: parsed.path || parsed.url || 'Imported',
      method: parsed.method || 'GET',
      url: parsed.url || `${parsed.baseURL ?? ''}${parsed.path ?? ''}`,
      headers: Object.entries(parsed.headers ?? {}).map(([k, v]) => ({
        key: k,
        value: String(v),
        enabled: true,
      })),
      body: prettyJSON(parsed.body || ''),
      response: null,
      sortOrder: items.length,
    }
    const saved = await scratchService.save(draft)
    setItems((prev) => [saved, ...prev])
    setActiveId(saved.id)
  }

  const importHAR = async (entries: any[]) => {
    if (!projectId) return
    const created: ScratchRequest[] = []
    for (let i = 0; i < entries.length; i++) {
      const e = entries[i]
      const saved = await scratchService.save({
        id: '',
        projectID: projectId,
        name: `${e.method} ${e.path || e.url}`,
        method: e.method,
        url: e.url,
        headers: Object.entries(e.headers ?? {}).map(([k, v]) => ({
          key: k,
          value: String(v),
          enabled: true,
        })),
        body: prettyJSON(e.body || ''),
        response: null,
        sortOrder: items.length + i,
      })
      created.push(saved)
    }
    setItems((prev) => [...created, ...prev])
    if (created[0]) setActiveId(created[0].id)
  }

  return (
    <div className="h-full flex gap-2 p-2 overflow-hidden">
      <Island as="aside" className="w-72 shrink-0">
        <div className="h-10 px-3 flex items-center justify-between border-b border-border/40 shrink-0">
          <div className="flex items-center gap-1.5">
            <Terminal className="w-3 h-3 text-muted-foreground" />
            <span className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
              Scratch
            </span>
            <span className="text-[10px] font-mono text-muted-foreground/60 tabular-nums">
              {items.length}
            </span>
          </div>
          <button
            type="button"
            onClick={create}
            className="inline-flex items-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground"
          >
            <Plus className="w-3 h-3" />
            New
          </button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {items.length === 0 ? (
            <p className="px-3 py-6 text-[11px] italic text-muted-foreground/70 text-center">
              No scratch requests yet.
            </p>
          ) : (
            <ul className="m-0 py-1 p-0 list-none space-y-px">
              {items.map((it) => (
                <ScratchItemRow
                  key={it.id}
                  item={it}
                  active={activeId === it.id}
                  onSelect={() => setActiveId(it.id)}
                  onRemove={() => remove(it.id)}
                />
              ))}
            </ul>
          )}
        </div>
        <div className="px-3 py-2 border-t border-border/40 shrink-0">
          <ImportMenu onCurl={importCurl} onHAR={importHAR} />
        </div>
      </Island>

      <Island as="main" className="flex-1">
        {!active ? (
          <div className="flex-1 flex items-center justify-center text-[12px] text-muted-foreground">
            Create or import a scratch request
          </div>
        ) : (
          <ScratchEditor
            key={active.id}
            value={active}
            onChange={(patch) => update(active.id, patch)}
            onRun={run}
            running={running}
            projectId={projectId ?? null}
          />
        )}
      </Island>
    </div>
  )
}

function ScratchItemRow({
  item,
  active,
  onSelect,
  onRemove,
}: {
  item: ScratchRequest
  active: boolean
  onSelect: () => void
  onRemove: () => void
}) {
  const { getMethodColor } = useHttpMethod()
  return (
    <li className="px-1.5">
      <div
        className={cn(
          'group flex items-center gap-2 px-2.5 py-2 rounded-md hover:bg-accent/40 cursor-pointer',
          active && 'bg-accent/60',
        )}
        onClick={onSelect}
      >
        <span
          className={cn(
            'inline-flex w-10 shrink-0 justify-center text-[8.5px] font-bold tracking-wider rounded px-1 py-px',
            getMethodColor(item.method),
          )}
        >
          {item.method}
        </span>
        <div className="flex-1 min-w-0">
          <div className="text-[11.5px] font-medium truncate">{item.name}</div>
          <div className="text-[10px] font-mono text-muted-foreground truncate">{item.url || '—'}</div>
        </div>
        {item.response && (
          <span
            className={cn(
              'text-[9.5px] font-mono tabular-nums shrink-0',
              item.response.status >= 400 ? 'text-rose-500/90' : 'text-emerald-500/80',
            )}
          >
            {item.response.status || 'ERR'}
          </span>
        )}
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          className="opacity-0 group-hover:opacity-100 inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-destructive shrink-0"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>
    </li>
  )
}

function ScratchEditor({
  value,
  onChange,
  onRun,
  running,
  projectId,
}: {
  value: ScratchRequest
  onChange: (patch: Partial<ScratchRequest>) => void
  onRun: () => void
  running: boolean
  projectId: string | null
}) {
  const setHeader = (idx: number, patch: Partial<HeaderRow>) => {
    const next = value.headers.map((h, i) => (i === idx ? { ...h, ...patch } : h))
    onChange({ headers: next })
  }
  const addHeader = () =>
    onChange({ headers: [...value.headers, { key: '', value: '', enabled: true }] })
  const removeHeader = (idx: number) =>
    onChange({ headers: value.headers.filter((_, i) => i !== idx) })

  return (
    <>
      <div className="h-10 px-3 flex items-center gap-2 border-b border-border/40 shrink-0">
        <select
          value={value.method}
          onChange={(e) => onChange({ method: e.target.value })}
          className="h-7 text-[11.5px] font-mono bg-input/30 border border-border/50 rounded-md px-2 shrink-0"
        >
          {METHODS.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
        <Input
          value={value.url}
          onChange={(e) => onChange({ url: e.target.value })}
          placeholder="https://api.example.com/v1/users"
          className="flex-1 h-7 text-[12px] font-mono"
        />
        <Button size="sm" className="h-7 gap-1.5 shrink-0" onClick={onRun} disabled={running || !value.url.trim()}>
          {running ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3 fill-current" />}
          Run
        </Button>
      </div>

      <div className="flex-1 grid grid-cols-2 min-h-0">
        <section className="border-r border-border/40 flex flex-col min-h-0">
          <Tabs defaultValue="body" className="flex-1 flex flex-col min-h-0">
            <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
              {[
                { v: 'body', label: 'Body' },
                { v: 'headers', label: `Headers${value.headers.length > 0 ? ` · ${value.headers.length}` : ''}` },
              ].map((t) => (
                <TabsTrigger
                  key={t.v}
                  value={t.v}
                  className="text-[11.5px] px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent dark:data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
                >
                  {t.label}
                </TabsTrigger>
              ))}
            </TabsList>
            <TabsContent value="body" className="flex-1 flex flex-col p-3 overflow-hidden mt-0">
              <BodyEditor
                method={value.method}
                body={value.body}
                onBodyChange={(v) => onChange({ body: v })}
              />
            </TabsContent>
            <TabsContent value="headers" className="flex-1 p-3 overflow-auto mt-0">
              <HeadersEditor
                headers={value.headers}
                onAdd={addHeader}
                onChange={setHeader}
                onRemove={removeHeader}
                showAuth={false}
              />
            </TabsContent>
          </Tabs>
        </section>

        <section className="flex flex-col min-h-0">
          {value.response?.timeline && <TimelineStrip timeline={value.response.timeline} />}
          {value.response && value.response.status >= 400 && value.response.body && (
            <ExceptionPanel
              projectId={projectId}
              body={formatBody(value.response.body)}
              status={value.response.status}
            />
          )}
          <Tabs defaultValue="json" className="flex-1 flex flex-col min-h-0">
            <TabsList className="w-full justify-between border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0">
              <div className="flex items-center gap-4">
                {[
                  { v: 'json', label: 'JSON' },
                  { v: 'rheaders', label: 'Headers' },
                ].map((t) => (
                  <TabsTrigger
                    key={t.v}
                    value={t.v}
                    className="text-[11.5px] px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent dark:data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
                  >
                    {t.label}
                  </TabsTrigger>
                ))}
              </div>
              <div className="flex items-center gap-2">
                {value.response && (
                  <span className="text-[10px] font-mono tabular-nums text-muted-foreground">
                    <span className={cn(value.response.status >= 400 ? 'text-rose-500/90' : 'text-emerald-500/80')}>
                      {value.response.status || 'ERR'}
                    </span>
                    {' · '}
                    {value.response.durationMs}ms
                  </span>
                )}
                {value.response && (
                  <>
                    <SaveResponseButton text={value.response.body} method={value.method} path={value.url} />
                    <CopyButton text={value.response.body} title="Copy response" />
                  </>
                )}
              </div>
            </TabsList>
            <TabsContent value="json" className="flex-1 min-h-0 p-3 overflow-hidden mt-0">
              {!value.response ? (
                <p className="h-full flex items-center justify-center text-[11.5px] italic text-muted-foreground/70">
                  Run to see response
                </p>
              ) : (
                <ResponseBodyView raw={value.response.body} />
              )}
            </TabsContent>
            <TabsContent value="rheaders" className="flex-1 p-0 overflow-auto mt-0">
              <HeadersList headers={value.response?.headers} highlightImportant />
            </TabsContent>
          </Tabs>
        </section>
      </div>
    </>
  )
}

