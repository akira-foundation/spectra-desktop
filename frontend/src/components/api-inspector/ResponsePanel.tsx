import { useState, useEffect } from 'react'
import { Copy, Download } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'
import materialLight from 'react-syntax-highlighter/dist/esm/styles/prism/material-light'
import {Button} from "@/components/ui/button";

interface ResponsePanelProps {
  responseData: any
}

export function ResponsePanel({ responseData }: ResponsePanelProps) {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    const checkDarkMode = () => {
      setIsDark(document.documentElement.classList.contains('dark'))
    }
    
    checkDarkMode()
    
    const observer = new MutationObserver(checkDarkMode)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
    
    return () => observer.disconnect()
  }, [])

  return (
    <div className="flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2  border-border/50">
        <div className="flex items-center gap-2">
          <Download className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold text-foreground">Response</h3>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8 p-1 hover:bg-muted rounded transition-colors">
          <Copy className="w-4 h-4 text-muted-foreground" />
        </Button>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="json" className="flex-1 flex flex-col">
        <TabsList className="w-full justify-start border-b border-border/50 rounded-none bg-transparent px-4 h-auto py-1">
          <TabsTrigger value="json" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">JSON</TabsTrigger>
          <TabsTrigger value="raw" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Raw</TabsTrigger>
          <TabsTrigger value="headers" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Headers</TabsTrigger>
          <TabsTrigger value="history" className="text-xs py-2 px-3 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">History (1)</TabsTrigger>
        </TabsList>

        {/* Response Content */}
        <TabsContent value="json" className="flex-1 p-4 overflow-auto">
          <SyntaxHighlighter
            language="json"
            style={isDark ? vscDarkPlus : materialLight}
            customStyle={{
              background: 'transparent',
              margin: 0,
              fontSize: '0.75rem',
              padding: 0,
            }}
            showLineNumbers={false}
            wrapLines={true}
          >
            {JSON.stringify(responseData, null, 2)}
          </SyntaxHighlighter>
        </TabsContent>

        <TabsContent value="raw" className="flex-1 p-4 overflow-auto">
          <pre className="text-xs text-foreground font-mono">{JSON.stringify(responseData, null, 2)}</pre>
        </TabsContent>

        <TabsContent value="headers" className="flex-1 p-4 overflow-auto">
          <div className="text-center py-8 text-sm text-muted-foreground">No headers</div>
        </TabsContent>

        <TabsContent value="history" className="flex-1 p-4 overflow-auto">
          <div className="text-center py-8 text-sm text-muted-foreground">No history</div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
