import { useMemo, useState } from 'react'
import { Search, X } from 'lucide-react'
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
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => filterEndpoints(endpoints, query), [endpoints, query])
  const expanded = useMemo(() => filtered.map((c) => c.category.toLowerCase()), [filtered])
  const totalMatches = useMemo(
    () => filtered.reduce((sum, c) => sum + c.items.length, 0),
    [filtered],
  )

  return (
    <div className="w-64 shrink-0 border-r border-border flex flex-col bg-transparent">
      <div className="h-10 px-1.5 flex items-center border-b border-border/60 shrink-0">
        <div className="relative w-full">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter endpoints"
            className="w-full h-7 pl-7 pr-7 text-[12px] bg-input/60 border border-border/50 rounded-md focus:outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/70"
          />
          {query && (
            <button
              type="button"
              onClick={() => setQuery('')}
              aria-label="Clear filter"
              className="absolute right-1.5 top-1/2 -translate-y-1/2 inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent/50"
            >
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {filtered.length === 0 ? (
          <p className="px-3 py-6 text-[11.5px] text-muted-foreground text-center">
            {query ? 'No endpoints match.' : 'No endpoints yet.'}
          </p>
        ) : (
          <Accordion
            key={query}
            type="multiple"
            defaultValue={query ? expanded : [filtered[0]?.category.toLowerCase()].filter(Boolean) as string[]}
            className="w-full"
          >
            {filtered.map((category, index) => (
              <div key={category.category}>
                <AccordionItem value={category.category.toLowerCase()} className="border-b-0">
                  <AccordionTrigger className="px-3 py-2 hover:no-underline cursor-pointer text-foreground/70 hover:text-foreground">
                    <div className="flex items-center gap-2 w-full pr-2">
                      <span className="text-[11px] font-semibold uppercase tracking-wider">
                        {category.category}
                      </span>
                      <span className="text-[10px] text-muted-foreground/70 font-mono">
                        {category.items.length}
                      </span>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent className="pb-2">
                    <div className="space-y-px px-1">
                      {category.items.map((endpoint) => (
                        <button
                          key={endpoint.path + endpoint.method + endpoint.name}
                          onClick={() => onSelectEndpoint(endpoint.tag)}
                          className={cn(
                            'group relative w-full text-left pl-2.5 pr-2 py-1.5 rounded-md transition-colors duration-150',
                            'hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring/40',
                            endpoint.active && 'bg-accent text-foreground',
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
                            <span
                              className={cn(
                                'text-[12px] truncate flex-1',
                                endpoint.active ? 'font-semibold' : 'font-medium',
                              )}
                            >
                              {endpoint.name}
                            </span>
                          </div>
                          <div className="ml-12 text-[10.5px] font-mono text-muted-foreground truncate mt-0.5">
                            {endpoint.path}
                          </div>
                        </button>
                      ))}
                    </div>
                  </AccordionContent>
                </AccordionItem>
                {index < filtered.length - 1 && <Separator className="my-1 opacity-50" />}
              </div>
            ))}
          </Accordion>
        )}
      </div>

      {query && (
        <div className="border-t border-border/60 px-3 py-1.5 text-[10.5px] text-muted-foreground tabular-nums">
          {totalMatches} match{totalMatches === 1 ? '' : 'es'}
        </div>
      )}
    </div>
  )
}

function filterEndpoints(categories: EndpointCategory[], query: string): EndpointCategory[] {
  const q = query.trim().toLowerCase()
  if (!q) return categories

  return categories
    .map((c) => {
      if (c.category.toLowerCase().includes(q)) {
        return c
      }
      const items = c.items.filter((item) => matchesEndpoint(item, q))
      if (items.length === 0) return null
      return { ...c, items, count: items.length }
    })
    .filter((c): c is EndpointCategory => c !== null)
}

function matchesEndpoint(endpoint: Endpoint, q: string): boolean {
  return (
    endpoint.method.toLowerCase().includes(q) ||
    endpoint.path.toLowerCase().includes(q) ||
    endpoint.name.toLowerCase().includes(q) ||
    endpoint.tag.toLowerCase().includes(q)
  )
}
