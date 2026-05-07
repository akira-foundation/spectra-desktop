import { useState, useEffect } from 'react'
import { RotateCcw, Play, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'
import materialLight from 'react-syntax-highlighter/dist/esm/styles/prism/material-light'

interface RequestPanelProps {
  requestBody: any
  onExecute: () => void
}

export function RequestPanel({ requestBody, onExecute }: RequestPanelProps) {
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
    <div className="flex-1 border-r border-border/50 flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-border/50">
        <div className="flex items-center gap-2 ">
          <Send className="w-4 h-4 text-muted-foreground " />
          <h3 className="text-sm font-semibold text-foreground">Request</h3>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8 p-1 hover:bg-muted rounded transition-colors">
          <RotateCcw className="w-4 h-4 text-muted-foreground" />
        </Button>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="body" className="flex-1 flex flex-col ">
        <TabsList className="w-full justify-start border-b border-border/50 rounded-none bg-transparent px-4 h-auto py-1">
          <TabsTrigger value="body" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Body</TabsTrigger>
          <TabsTrigger value="params" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Params</TabsTrigger>
          <TabsTrigger value="headers" className="text-xs py-2 px-3 mr-8 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Headers</TabsTrigger>
          <TabsTrigger value="cookies" className="text-xs py-2 px-3 cursor-pointer data-[state=active]:border-b-2 data-[state=active]:border-primary">Cookies</TabsTrigger>
        </TabsList>

        <TabsContent value="body" className="flex-1 flex flex-col p-4 overflow-hidden">
          <div className="flex items-center justify-between mb-3 gap-2">
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-medium text-muted-foreground">Format:</span>
              <div className="flex items-center gap-1">
                <button className="px-2.5 py-1 text-xs bg-primary/10 text-primary rounded hover:bg-primary/20 transition">JSON</button>
                <button className="px-2.5 py-1 text-xs text-muted-foreground hover:bg-muted rounded transition">Form</button>
              </div>
            </div>
          </div>
          <div className="flex-1 overflow-auto">
            <SyntaxHighlighter
              language="json"
              style={isDark ? vscDarkPlus : materialLight}
              customStyle={{
                margin: 0,
                fontSize: '0.75rem',
                background: 'transparent',
                padding: 0,
              }}
              showLineNumbers={false}
              wrapLines={true}
            >
              {JSON.stringify(requestBody, null, 2)}
            </SyntaxHighlighter>
          </div>
        </TabsContent>

        <TabsContent value="params" className="flex-1 p-4 overflow-auto">
          <div className="text-center py-8 text-sm text-muted-foreground">No parameters</div>
        </TabsContent>

        <TabsContent value="headers" className="flex-1 p-4 overflow-auto">
          <div className="text-center py-8 text-sm text-muted-foreground">No headers</div>
        </TabsContent>

        <TabsContent value="cookies" className="flex-1 p-4 overflow-auto">
          <div className="text-center py-8 text-sm text-muted-foreground">No cookies</div>
        </TabsContent>
      </Tabs>

      {/* Execute Button */}
      <div className="px-4 py-3 border-t border-border/50 bg-background">
        <Button
          onClick={onExecute}
          className="w-full bg-gradient-to-r from-violet-600 to-blue-600 hover:from-violet-700 hover:to-blue-700 text-white font-semibold h-9"
        >
          <Play className="w-4 h-4 mr-2" />
          Execute
          <span className="ml-auto text-xs opacity-80">⌘↵</span>
        </Button>
      </div>
    </div>
  )
}
