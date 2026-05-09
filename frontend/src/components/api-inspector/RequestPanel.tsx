import { Play, Send, FileCheck, Sparkles, Shuffle, FileX2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useState } from 'react'
import { ParamsEditor } from './ParamsEditor'
import { JsonEditor } from './JsonEditor'
import { HeadersEditor, type HeaderRow } from './HeadersEditor'
import { FormBodyEditor } from './FormBodyEditor'
import { TestsEditor } from './TestsEditor'
import { CapturesEditor } from './CapturesEditor'
import type { TestResult } from '@/services/testsService'
import type { CapturedValue } from '@/services/capturesService'
import type { QueryParam } from '@/lib/route-params'
import type { RequestSchema } from '@/lib/request-schema'
import { sourceLabel } from '@/lib/request-schema'
import { cn } from '@/lib/utils'

const NO_BODY_METHODS = new Set(['GET', 'HEAD', 'DELETE', 'OPTIONS'])

interface RequestPanelProps {
  method?: string
  requestBody: string
  onRequestBodyChange: (value: string) => void
  onResetBody: () => void
  bodyTouched: boolean
  schema: RequestSchema | null
  routeParams: string[]
  routeValues: string[]
  onRouteValueChange: (index: number, value: string) => void
  queryParams: QueryParam[]
  onQueryAdd: () => void
  onQueryChange: (index: number, patch: Partial<QueryParam>) => void
  onQueryRemove: (index: number) => void
  headers: HeaderRow[]
  onHeaderAdd: () => void
  onHeaderChange: (index: number, patch: Partial<HeaderRow>) => void
  onHeaderRemove: (index: number) => void
  onExecute: () => void
  executing?: boolean
  variables?: Record<string, string>
  scope?: string
  projectId?: string | null
  endpointId?: string | null
  endpointPath?: string | null
  testResults?: TestResult[]
  responseBody?: string
  responseHeaders?: Record<string, string[]>
  autoAuth?: { scheme?: string; tokenPreview?: string } | null
  onOpenAuth?: () => void
  capturedValues?: CapturedValue[]
  onCapturedChange?: (values: CapturedValue[]) => void
}

export function RequestPanel({
  method,
  requestBody,
  onRequestBodyChange,
  onResetBody,
  bodyTouched,
  schema,
  routeParams,
  routeValues,
  onRouteValueChange,
  queryParams,
  onQueryAdd,
  onQueryChange,
  onQueryRemove,
  headers,
  onHeaderAdd,
  onHeaderChange,
  onHeaderRemove,
  onExecute,
  executing = false,
  variables,
  projectId,
  endpointId,
  endpointPath,
  testResults,
  responseBody,
  responseHeaders,
  autoAuth,
  onOpenAuth,
  capturedValues,
  onCapturedChange,
}: RequestPanelProps) {
  const [bodyMode, setBodyMode] = useState<'json' | 'form'>('json')
  const requiredCount = schema?.fields.filter((f) => f.required).length ?? 0

  return (
    <div className="flex flex-col min-w-0 min-h-0 h-full border-r border-border bg-transparent">
      <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
        <div className="flex items-center gap-1.5">
          <Send className="w-3.5 h-3.5 text-muted-foreground" />
          <h3 className="text-[11.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            Request
          </h3>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="h-6 px-2 text-[10.5px] gap-1"
          onClick={onResetBody}
          disabled={!schema || schema.fields.length === 0}
        >
          <Shuffle className="w-3 h-3" />
          Re-roll values
        </Button>
      </div>

      <Tabs defaultValue="body" className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {(() => {
            const passCount = testResults?.filter((r) => r.pass).length ?? 0
            const failCount = testResults?.filter((r) => !r.pass).length ?? 0
            const endpointKey = method && endpointPath ? `${method.toUpperCase()} ${endpointPath}` : null
            const endpointCaptures = endpointKey
              ? (capturedValues ?? []).filter((c) => c.endpointKey === endpointKey)
              : []
            const tabs = [
              { v: 'body', label: 'Body' },
              { v: 'params', label: 'Params' },
              { v: 'headers', label: 'Headers' },
              { v: 'tests', label: testResults && testResults.length > 0 ? `Tests · ${passCount}/${passCount + failCount}` : 'Tests' },
              { v: 'captures', label: endpointCaptures.length > 0 ? `Captures · ${endpointCaptures.length}` : 'Captures' },
              { v: 'cookies', label: 'Cookies' },
            ]
            return tabs.map((t) => (
              <TabsTrigger
                key={t.v}
                value={t.v}
                className="text-[11.5px] capitalize px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
              >
                {t.label}
              </TabsTrigger>
            ))
          })()}
        </TabsList>

        <TabsContent value="body" className="flex-1 flex flex-col p-3 overflow-hidden mt-0">
          {method && NO_BODY_METHODS.has(method.toUpperCase()) ? (
            <NoBodyState method={method.toUpperCase()} />
          ) : (
          <>
          <div className="flex items-center gap-2 mb-2">
            <button
              type="button"
              onClick={() => setBodyMode('json')}
              className={cn(
                'px-2 py-0.5 text-[10.5px] rounded-sm transition-colors',
                bodyMode === 'json'
                  ? 'bg-primary/15 text-primary hover:bg-primary/25'
                  : 'text-muted-foreground hover:bg-accent/60',
              )}
            >
              JSON
            </button>
            <button
              type="button"
              onClick={() => setBodyMode('form')}
              disabled={!schema || schema.fields.length === 0}
              className={cn(
                'px-2 py-0.5 text-[10.5px] rounded-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed',
                bodyMode === 'form'
                  ? 'bg-primary/15 text-primary hover:bg-primary/25'
                  : 'text-muted-foreground hover:bg-accent/60',
              )}
            >
              Form
            </button>
            {schema && schema.fields.length > 0 && (
              <SchemaBadge schema={schema} requiredCount={requiredCount} touched={bodyTouched} />
            )}
          </div>
          <div className="flex-1 min-h-0 overflow-auto">
            {bodyMode === 'form' && schema && schema.fields.length > 0 ? (
              <FormBodyEditor
                value={requestBody}
                schema={schema}
                onChange={onRequestBodyChange}
                variables={variables}
              />
            ) : (
              <JsonEditor
                value={requestBody}
                onChange={onRequestBodyChange}
                placeholder="{}"
                variables={variables}
              />
            )}
          </div>
          </>
          )}
        </TabsContent>

        <TabsContent value="params" className="flex-1 p-3 overflow-auto mt-0">
          <ParamsEditor
            routeParams={routeParams}
            routeValues={routeValues}
            onRouteValueChange={onRouteValueChange}
            queryParams={queryParams}
            onQueryAdd={onQueryAdd}
            onQueryChange={onQueryChange}
            onQueryRemove={onQueryRemove}
            variables={variables}
          />
        </TabsContent>
        <TabsContent value="headers" className="flex-1 p-3 overflow-auto mt-0">
          <HeadersEditor
            headers={headers}
            onAdd={onHeaderAdd}
            onChange={onHeaderChange}
            onRemove={onHeaderRemove}
            variables={variables}
            autoAuth={autoAuth}
            onOpenAuth={onOpenAuth}
          />
        </TabsContent>
        <TabsContent value="tests" className="flex-1 p-3 overflow-auto mt-0">
          <TestsEditor
            projectId={projectId ?? null}
            method={method ?? null}
            path={endpointPath ?? null}
            results={testResults}
            responseBody={responseBody}
            responseHeaders={responseHeaders}
          />
        </TabsContent>
        <TabsContent value="captures" className="flex-1 p-3 overflow-auto mt-0">
          <CapturesEditor
            projectId={projectId ?? null}
            method={method ?? null}
            path={endpointPath ?? null}
            responseBody={responseBody}
            responseHeaders={responseHeaders}
            capturedValues={capturedValues}
            onCapturedChange={onCapturedChange}
          />
        </TabsContent>
        <TabsContent value="cookies" className="flex-1 p-4 text-center text-[11.5px] text-muted-foreground mt-0">
          No cookies
        </TabsContent>
      </Tabs>

      <div className="px-3 py-2 border-t border-border/40">
        <button
          onClick={onExecute}
          disabled={executing}
          className="group w-full h-8 inline-flex items-center rounded-md border border-border/60 bg-card hover:bg-accent/60 active:bg-accent text-foreground text-[12px] font-medium transition-colors px-2.5 disabled:opacity-60 disabled:cursor-progress"
        >
          {executing ? (
            <span className="w-3.5 h-3.5 rounded-full border-2 border-emerald-500 border-t-transparent animate-spin" />
          ) : (
            <Play className="w-3.5 h-3.5 fill-emerald-500 text-emerald-500 shrink-0" />
          )}
          <span className="ml-2">{executing ? 'Executing...' : 'Execute'}</span>
          <span className="ml-auto flex items-center gap-1 text-muted-foreground group-hover:text-foreground/70">
            <kbd className="inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded border border-border/60 bg-muted/50 text-[10.5px] font-sans leading-none">
              ⌘
            </kbd>
            <kbd className="inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded border border-border/60 bg-muted/50 text-[10.5px] font-sans leading-none">
              ⏎
            </kbd>
          </span>
        </button>
      </div>
    </div>
  )
}

interface SchemaBadgeProps {
  schema: RequestSchema
  requiredCount: number
  touched: boolean
}

function SchemaBadge({ schema, requiredCount, touched }: SchemaBadgeProps) {
  const Icon = schema.confidence === 'high' ? FileCheck : Sparkles
  const tone =
    schema.confidence === 'high'
      ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/15'
      : 'border-amber-500/30 bg-amber-500/10 text-amber-500 hover:bg-amber-500/15'
  return (
    <Dialog>
      <DialogTrigger asChild>
        <button
          type="button"
          className={cn(
            'ml-auto inline-flex items-center gap-1 text-[10px] font-medium px-1.5 py-0.5 rounded border transition-colors cursor-pointer',
            tone,
          )}
          title={`${sourceLabel(schema.source)} · ${schema.confidence} confidence`}
        >
          <Icon className="w-3 h-3" />
          {sourceLabel(schema.source)}
          <span className="text-muted-foreground/80">·</span>
          <span>{requiredCount} required</span>
          {touched && <span className="text-muted-foreground/70 ml-0.5">· edited</span>}
        </button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-3 shrink-0 border-b border-border/40">
          <DialogTitle className="text-base flex items-center gap-2">
            <Icon className={cn('w-4 h-4', schema.confidence === 'high' ? 'text-emerald-500' : 'text-amber-500')} />
            Validation rules
          </DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Detected from {sourceLabel(schema.source)} with {schema.confidence} confidence.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4 space-y-3">
          <div className="grid grid-cols-3 gap-2 text-center">
            <Stat label="Fields" value={schema.fields.length} />
            <Stat label="Required" value={requiredCount} accent="rose" />
            <Stat label="Optional" value={schema.fields.length - requiredCount} />
          </div>
          {schema.fields.length === 0 ? (
            <div className="rounded-md border border-border/60 bg-card/40 px-4 py-8 text-center">
              <p className="text-[11.5px] italic text-muted-foreground/70">No fields detected.</p>
            </div>
          ) : (
            <ul className="m-0 p-0 list-none space-y-2">
              {schema.fields.map((f) => (
                <li key={f.name} className="rounded-md border border-border/50 bg-card/40 overflow-hidden">
                  <div className="px-3 py-2 flex items-center gap-2 bg-muted/20 border-b border-border/30">
                    <code className="text-[12px] font-mono font-semibold text-foreground">
                      {f.name}
                    </code>
                    {f.required && (
                      <span className="inline-flex items-center text-[9px] font-bold uppercase tracking-wider text-rose-500 px-1.5 py-0.5 rounded bg-rose-500/10">
                        required
                      </span>
                    )}
                    <span
                      className={cn(
                        'ml-auto text-[9.5px] font-mono px-1.5 py-0.5 rounded',
                        typeTone(f.type),
                      )}
                    >
                      {f.type || 'mixed'}
                    </span>
                  </div>
                  {f.rules && f.rules.length > 0 ? (
                    <div className="px-3 py-2 flex flex-wrap gap-1.5">
                      {f.rules.map((r, i) => (
                        <code
                          key={i}
                          className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-muted/40 text-muted-foreground/90 border border-border/30"
                        >
                          {r}
                        </code>
                      ))}
                    </div>
                  ) : (
                    <div className="px-3 py-1.5 text-[10px] italic text-muted-foreground/50">No rules</div>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>

      </DialogContent>
    </Dialog>
  )
}

function Stat({ label, value, accent }: { label: string; value: number; accent?: 'rose' }) {
  return (
    <div className="rounded-md border border-border/60 bg-card/40 px-3 py-2">
      <div className={cn('text-[16px] font-semibold tabular-nums', accent === 'rose' && 'text-rose-500')}>
        {value}
      </div>
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground/70">{label}</div>
    </div>
  )
}

function typeTone(t: string): string {
  const k = (t || '').toLowerCase()
  if (k.startsWith('string') || k === 'email' || k === 'uuid') return 'bg-blue-500/15 text-blue-400'
  if (k.startsWith('int') || k === 'numeric' || k === 'number') return 'bg-purple-500/15 text-purple-400'
  if (k.startsWith('bool')) return 'bg-amber-500/15 text-amber-500'
  if (k === 'date' || k === 'datetime' || k === 'timestamp') return 'bg-pink-500/15 text-pink-400'
  if (k === 'array' || k === 'object') return 'bg-teal-500/15 text-teal-400'
  return 'bg-muted/40 text-muted-foreground'
}

function NoBodyState({ method }: { method: string }) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center text-center px-6 py-8 gap-2">
      <span className="inline-flex w-9 h-9 items-center justify-center rounded-md bg-muted/40 text-muted-foreground">
        <FileX2 className="w-4 h-4" />
      </span>
      <p className="text-[12.5px] font-medium text-foreground/85">No body needed</p>
      <p className="text-[11px] text-muted-foreground max-w-xs leading-relaxed">
        {method} requests don&apos;t typically carry a payload. Use Params or Headers if you need
        to send data.
      </p>
    </div>
  )
}
