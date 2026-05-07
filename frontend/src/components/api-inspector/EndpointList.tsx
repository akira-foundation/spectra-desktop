import { Search } from 'lucide-react'
import { Separator } from '@/components/ui/separator'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'

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
    <div className="w-64 shrink-0 border-r border-border/60 flex flex-col bg-sidebar/40 backdrop-blur-sm">
      <div className="p-2.5 border-b border-border/50">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Filter endpoints"
            className="w-full h-7 pl-7 pr-2 text-[12px] bg-input/60 border border-border/50 rounded-md focus:outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/70"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        <Accordion type="multiple" defaultValue={['auth']} className="w-full">
          {endpoints.map((category, index) => (
            <div key={category.category}>
              <AccordionItem value={category.category.toLowerCase()} className="border-b-0">
                <AccordionTrigger className="px-3 py-2 hover:no-underline cursor-pointer text-foreground/70 hover:text-foreground">
                  <div className="flex items-center gap-2 w-full pr-2">
                    <span className="text-[11px] font-semibold uppercase tracking-wider">
                      {category.category}
                    </span>
                    <span className="text-[10px] text-muted-foreground/70 font-mono">
                      {category.count}
                    </span>
                  </div>
                </AccordionTrigger>
                <AccordionContent className="pb-2">
                  <div className="space-y-px px-1">
                    {category.items.map((endpoint) => (
                      <button
                        key={endpoint.path + endpoint.method}
                        onClick={() => onSelectEndpoint(endpoint.tag)}
                        className={cn(
                          'group w-full text-left px-2 py-1.5 rounded-md transition-colors',
                          endpoint.active ? 'bg-accent/80 text-foreground' : 'hover:bg-accent/40',
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <span
                            className={cn(
                              'inline-flex w-10 shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1 py-0.5',
                              getMethodColor(endpoint.method),
                            )}
                          >
                            {endpoint.method}
                          </span>
                          <span className="text-[12px] font-medium truncate flex-1">
                            {endpoint.name}
                          </span>
                        </div>
                        <div className="ml-12 text-[10.5px] font-mono text-muted-foreground/80 truncate mt-0.5">
                          {endpoint.path}
                        </div>
                      </button>
                    ))}
                  </div>
                </AccordionContent>
              </AccordionItem>
              {index < endpoints.length - 1 && <Separator className="my-1 opacity-50" />}
            </div>
          ))}
        </Accordion>
      </div>
    </div>
  )
}
