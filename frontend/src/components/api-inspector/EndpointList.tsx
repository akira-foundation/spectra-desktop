import { Search } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { useHttpMethod } from '@/hooks/useHttpMethod'

interface Endpoint {
  method: string
  name: string
  path: string
  tag: string
  active?: boolean
}

interface EndpointCategory {
  category: string
  count: number
  items: Endpoint[]
}

interface EndpointListProps {
  endpoints: EndpointCategory[]
  onSelectEndpoint: (tag: string) => void
}

export function EndpointList({ endpoints, onSelectEndpoint }: EndpointListProps) {
  const { getMethodColor } = useHttpMethod()

  return (
    <div className="w-80 border-r border-border/50 flex flex-col bg-card/30">
      {/* Search */}
      <div className="p-4 border-b border-border/50">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search endpoints"
            className="w-full pl-9 pr-3 py-2 text-sm bg-background border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
        </div>
      </div>

      {/* Endpoint List */}
      <div className="flex-1 overflow-y-auto">
        <Accordion type="multiple" defaultValue={["auth"]} className="w-full">
          {endpoints.map((category, index) => (
            <div key={category.category}>
              <AccordionItem value={category.category.toLowerCase()} className="border-b-0">
                <AccordionTrigger className="px-4 py-3 hover:no-underline hover:text-primary cursor-pointer">
                  <div className="flex items-center gap-2 w-full pr-2">
                    <span className="text-base font-semibold">{category.category}</span>
                    <Badge variant="secondary" className="h-5 px-2 text-xs bg-muted/50 text-muted-foreground/70">
                      {category.count}
                    </Badge>
                  </div>
                </AccordionTrigger>
                <AccordionContent className="pb-4">
                  <div className="space-y-1 px-2">
                    {category.items.map((endpoint) => (
                      <button
                        key={endpoint.path}
                        onClick={() => onSelectEndpoint(endpoint.tag)}
                        className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                          endpoint.active
                            ? 'bg-primary/10 border border-primary/20'
                            : 'hover:bg-muted/50'
                        }`}
                      >
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`text-xs font-bold px-2 py-0.5 rounded ${getMethodColor(endpoint.method)}`}>
                            {endpoint.method}
                          </span>
                          <span className="font-medium">{endpoint.name}</span>
                        </div>
                        <div className="text-xs text-muted-foreground truncate">{endpoint.path}</div>
                        <div className="text-xs text-muted-foreground/60 mt-1">{endpoint.tag}</div>
                      </button>
                    ))}
                  </div>
                </AccordionContent>
              </AccordionItem>
              {index < endpoints.length - 1 && <Separator className="my-2" />}
            </div>
          ))}
        </Accordion>
      </div>
    </div>
  )
}
