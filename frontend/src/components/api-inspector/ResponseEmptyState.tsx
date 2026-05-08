import { Inbox } from 'lucide-react'
import { EmptyState, Kbd } from '@/components/common/EmptyState'

export function ResponseEmptyState() {
  return (
    <EmptyState
      size="sm"
      icon={Inbox}
      title="No response yet"
      description="Run the request to see the response inline."
      hint={
        <span className="inline-flex items-center gap-1.5">
          Press <Kbd>⌘</Kbd> <Kbd>⏎</Kbd> to execute
        </span>
      }
    />
  )
}
