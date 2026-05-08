import { MousePointer } from 'lucide-react'
import { EmptyState, Kbd } from '@/components/common/EmptyState'

export function EndpointEmptyState() {
  return (
    <EmptyState
      icon={MousePointer}
      title="Select an endpoint"
      description="Pick a route from the sidebar to inspect its request and response."
      hint={
        <span className="inline-flex items-center gap-1.5">
          Use <Kbd>↑</Kbd> <Kbd>↓</Kbd> to navigate the list
        </span>
      }
    />
  )
}
