import { useState, useEffect } from 'react'
import { RotateCcw, Play, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'
import oneLight from 'react-syntax-highlighter/dist/esm/styles/prism/one-light'
import { ParamsEditor } from './ParamsEditor'
import type { QueryParam } from '@/lib/route-params'

interface RequestPanelProps {
  requestBody: any
  routeParams: string[]
  routeValues: string[]
  onRouteValueChange: (index: number, value: string) => void
  queryParams: QueryParam[]
  onQueryAdd: () => void
  onQueryChange: (index: number, patch: Partial<QueryParam>) => void
  onQueryRemove: (index: number) => void
  onExecute: () => void
  executing?: boolean
}

export function RequestPanel({
  requestBody,
  routeParams,
  routeValues,
  onRouteValueChange,
  queryParams,
  onQueryAdd,
  onQueryChange,
  onQueryRemove,
  onExecute,
  executing = false,
}: RequestPanelProps) {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    const check = () => setIsDark(document.documentElement.classList.contains('dark'))
    check()
    const observer = new MutationObserver(check)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
    return () => observer.disconnect()
  }, [])

  return (
    <div className="flex flex-col min-w-0 border-r border-border bg-transparent">
      <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
        <div className="flex items-center gap-1.5">
          <Send className="w-3.5 h-3.5 text-muted-foreground" />
          <h3 className="text-[11.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            Request
          </h3>
        </div>
        <Button variant="ghost" size="icon-sm" className="h-6 w-6">
          <RotateCcw className="w-3 h-3 text-muted-foreground" />
        </Button>
      </div>

      <Tabs defaultValue="body" className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {['body', 'params', 'headers', 'cookies'].map((v) => (
            <TabsTrigger
              key={v}
              value={v}
              className="text-[11.5px] capitalize px-0 h-full rounded-none bg-transparent border-0 border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none text-muted-foreground data-[state=active]:text-foreground"
            >
              {v}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="body" className="flex-1 flex flex-col p-3 overflow-hidden mt-0">
          <div className="flex items-center gap-1 mb-2">
            <button className="px-2 py-0.5 text-[10.5px] bg-primary/15 text-primary rounded-sm hover:bg-primary/25 transition-colors">
              JSON
            </button>
            <button className="px-2 py-0.5 text-[10.5px] text-muted-foreground hover:bg-accent/60 rounded-sm transition-colors">
              Form
            </button>
          </div>
          <div className="flex-1 overflow-auto rounded-md border border-border/40 bg-muted/20 p-2">
            <SyntaxHighlighter
              language="json"
              style={isDark ? vscDarkPlus : oneLight}
              customStyle={{
                margin: 0,
                fontSize: '11.5px',
                background: 'transparent',
                padding: 0,
                fontFamily: 'var(--font-mono)',
              }}
              codeTagProps={{
                style: {
                  background: 'transparent',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11.5px',
                },
              }}
              showLineNumbers={false}
              wrapLines
            >
              {JSON.stringify(requestBody, null, 2)}
            </SyntaxHighlighter>
          </div>
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
          />
        </TabsContent>
        <TabsContent value="headers" className="flex-1 p-4 text-center text-[11.5px] text-muted-foreground mt-0">
          No headers
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
