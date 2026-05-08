import { useState, useEffect } from 'react'
import { Copy, Download } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'
import materialLight from 'react-syntax-highlighter/dist/esm/styles/prism/material-light'
import { Button } from '@/components/ui/button'

interface ResponsePanelProps {
  responseData: any
}

export function ResponsePanel({ responseData }: ResponsePanelProps) {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    const check = () => setIsDark(document.documentElement.classList.contains('dark'))
    check()
    const observer = new MutationObserver(check)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
    return () => observer.disconnect()
  }, [])

  return (
    <div className="flex flex-col min-w-0 bg-sidebar">
      <div className="h-9 px-3 flex items-center justify-between border-b border-border/40">
        <div className="flex items-center gap-1.5">
          <Download className="w-3.5 h-3.5 text-muted-foreground" />
          <h3 className="text-[11.5px] font-semibold uppercase tracking-wider text-muted-foreground">
            Response
          </h3>
        </div>
        <Button variant="ghost" size="icon-sm" className="h-6 w-6">
          <Copy className="w-3 h-3 text-muted-foreground" />
        </Button>
      </div>

      <Tabs defaultValue="json" className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full justify-start border-b border-border/40 rounded-none bg-transparent px-3 h-8 py-0 gap-4">
          {[
            { v: 'json', label: 'JSON' },
            { v: 'raw', label: 'Raw' },
            { v: 'headers', label: 'Headers' },
            { v: 'history', label: 'History' },
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

        <TabsContent value="json" className="flex-1 p-3 overflow-auto mt-0">
          <div className="rounded-md border border-border/40 bg-muted/20 p-2">
            <SyntaxHighlighter
              language="json"
              style={isDark ? vscDarkPlus : materialLight}
              customStyle={{
                background: 'transparent',
                margin: 0,
                fontSize: '11.5px',
                padding: 0,
                fontFamily: 'var(--font-mono)',
              }}
              showLineNumbers={false}
              wrapLines
            >
              {JSON.stringify(responseData, null, 2)}
            </SyntaxHighlighter>
          </div>
        </TabsContent>

        <TabsContent value="raw" className="flex-1 p-3 overflow-auto mt-0">
          <pre className="text-[11.5px] text-foreground font-mono whitespace-pre-wrap">
            {JSON.stringify(responseData, null, 2)}
          </pre>
        </TabsContent>
        <TabsContent value="headers" className="flex-1 p-4 text-center text-[11.5px] text-muted-foreground mt-0">
          No headers
        </TabsContent>
        <TabsContent value="history" className="flex-1 p-4 text-center text-[11.5px] text-muted-foreground mt-0">
          No history
        </TabsContent>
      </Tabs>
    </div>
  )
}
