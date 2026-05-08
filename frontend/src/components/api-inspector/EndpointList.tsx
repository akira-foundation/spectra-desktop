import { useEffect, useMemo, useRef, useState } from 'react'
import { Search, X, Star, GripVertical } from 'lucide-react'
import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Separator } from '@/components/ui/separator'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { PINNED_CATEGORY, endpointKey } from '@/lib/group-endpoints'

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
  pinnedKeys?: string[]
  onTogglePin?: (key: string) => void
  onReorder?: (order: string[]) => void
}

export function EndpointList({
  endpoints,
  onSelectEndpoint,
  pinnedKeys = [],
  onTogglePin,
  onReorder,
}: EndpointListProps) {
  const { getMethodColor } = useHttpMethod()
  const [query, setQuery] = useState('')
  const pinnedSet = useMemo(() => new Set(pinnedKeys), [pinnedKeys])

  const filtered = useMemo(() => filterEndpoints(endpoints, query), [endpoints, query])
  const expanded = useMemo(() => filtered.map((c) => c.category.toLowerCase()), [filtered])
  const activeCategory = useMemo(() => {
    for (const c of filtered) {
      if (c.items.some((i) => i.active)) return c.category.toLowerCase()
    }
    return null
  }, [filtered])
  const activeRef = useRef<HTMLButtonElement | null>(null)
  useEffect(() => {
    if (activeRef.current) {
      activeRef.current.scrollIntoView({ block: 'nearest', behavior: 'auto' })
    }
  }, [activeCategory])
  const initialOpen = useMemo(() => {
    if (query) return expanded
    const seed: string[] = []
    if (activeCategory) seed.push(activeCategory)
    const first = filtered[0]?.category.toLowerCase()
    if (first && !seed.includes(first)) seed.push(first)
    return seed
  }, [query, expanded, activeCategory, filtered])
  const totalMatches = useMemo(
    () => filtered.reduce((sum, c) => sum + c.items.length, 0),
    [filtered],
  )

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }))

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const draggables = filtered.filter((c) => c.category !== PINNED_CATEGORY)
    const ids = draggables.map((c) => c.category)
    const oldIndex = ids.indexOf(String(active.id))
    const newIndex = ids.indexOf(String(over.id))
    if (oldIndex < 0 || newIndex < 0) return
    const next = arrayMove(ids, oldIndex, newIndex)
    onReorder?.(next)
  }

  const draggableCategories = filtered.filter((c) => c.category !== PINNED_CATEGORY)
  const pinnedCategory = filtered.find((c) => c.category === PINNED_CATEGORY)

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
            key={`${query}|${activeCategory ?? ''}`}
            type="multiple"
            defaultValue={initialOpen}
            className="w-full"
          >
            {pinnedCategory && (
              <CategoryItem
                category={pinnedCategory}
                getMethodColor={getMethodColor}
                onSelect={onSelectEndpoint}
                pinnedSet={pinnedSet}
                onTogglePin={onTogglePin}
                activeRef={activeRef}
                showSeparator={draggableCategories.length > 0}
                isPinnedSection
              />
            )}

            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext
                items={draggableCategories.map((c) => c.category)}
                strategy={verticalListSortingStrategy}
              >
                {draggableCategories.map((category, index) => (
                  <SortableCategoryItem
                    key={category.category}
                    id={category.category}
                    category={category}
                    getMethodColor={getMethodColor}
                    onSelect={onSelectEndpoint}
                    pinnedSet={pinnedSet}
                    onTogglePin={onTogglePin}
                    activeRef={activeRef}
                    showSeparator={index < draggableCategories.length - 1}
                  />
                ))}
              </SortableContext>
            </DndContext>
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

interface CategoryItemProps {
  category: EndpointCategory
  getMethodColor: (method: string) => string
  onSelect: (tag: string) => void
  pinnedSet: Set<string>
  onTogglePin?: (key: string) => void
  activeRef: React.MutableRefObject<HTMLButtonElement | null>
  showSeparator?: boolean
  isPinnedSection?: boolean
  dragHandleProps?: Record<string, unknown>
  setNodeRef?: (node: HTMLElement | null) => void
  style?: React.CSSProperties
  isDragging?: boolean
}

function CategoryItem({
  category,
  getMethodColor,
  onSelect,
  pinnedSet,
  onTogglePin,
  activeRef,
  showSeparator,
  isPinnedSection,
  dragHandleProps,
  setNodeRef,
  style,
  isDragging,
}: CategoryItemProps) {
  return (
    <div ref={setNodeRef} style={style} className={cn(isDragging && 'opacity-60')}>
      <AccordionItem value={category.category.toLowerCase()} className="border-b-0">
        <div className="flex items-center group/cat">
          {!isPinnedSection && dragHandleProps && (
            <button
              type="button"
              {...dragHandleProps}
              aria-label="Drag to reorder"
              className="ml-1.5 inline-flex h-5 w-5 items-center justify-center text-muted-foreground/40 hover:text-muted-foreground cursor-grab active:cursor-grabbing"
              onClick={(e) => e.stopPropagation()}
            >
              <GripVertical className="w-3 h-3" />
            </button>
          )}
          <AccordionTrigger
            className={cn(
              'flex-1 px-3 py-2 hover:no-underline cursor-pointer text-foreground/70 hover:text-foreground',
              isPinnedSection && 'text-amber-500/90 hover:text-amber-500',
            )}
          >
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <span className="text-[11.5px] font-medium capitalize truncate">
                {category.category.toLowerCase()}
              </span>
              <span className="text-[10px] text-muted-foreground/70 font-mono shrink-0">
                {category.items.length}
              </span>
            </div>
          </AccordionTrigger>
        </div>
        <AccordionContent className="pb-2">
          <div className="space-y-px px-1">
            {category.items.map((endpoint) => {
              const key = endpointKey(endpoint.method, endpoint.path)
              const isPinned = pinnedSet.has(key)
              return (
                <button
                  key={endpoint.path + endpoint.method + endpoint.name}
                  ref={endpoint.active ? activeRef : undefined}
                  onClick={() => onSelect(endpoint.tag)}
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
                    {onTogglePin && (
                      <span
                        role="button"
                        tabIndex={0}
                        onClick={(e) => {
                          e.stopPropagation()
                          onTogglePin(key)
                        }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.stopPropagation()
                            e.preventDefault()
                            onTogglePin(key)
                          }
                        }}
                        aria-label={isPinned ? 'Unpin' : 'Pin'}
                        className={cn(
                          'inline-flex h-5 w-5 items-center justify-center rounded transition-opacity',
                          isPinned
                            ? 'text-amber-500 opacity-100'
                            : 'text-muted-foreground/40 opacity-0 group-hover:opacity-100 hover:text-amber-500',
                        )}
                      >
                        <Star className={cn('w-3 h-3', isPinned && 'fill-amber-500')} />
                      </span>
                    )}
                  </div>
                  <div className="ml-12 text-[10.5px] font-mono text-muted-foreground truncate mt-0.5">
                    {endpoint.path}
                  </div>
                </button>
              )
            })}
          </div>
        </AccordionContent>
      </AccordionItem>
      {showSeparator && <Separator className="my-1 opacity-50" />}
    </div>
  )
}

interface SortableProps extends Omit<CategoryItemProps, 'dragHandleProps' | 'setNodeRef' | 'style' | 'isDragging'> {
  id: string
}

function SortableCategoryItem(props: SortableProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: props.id,
  })
  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
  }
  return (
    <CategoryItem
      {...props}
      setNodeRef={setNodeRef}
      style={style}
      isDragging={isDragging}
      dragHandleProps={{ ...attributes, ...listeners }}
    />
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
