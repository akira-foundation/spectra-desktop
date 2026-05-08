import { Skeleton } from '@/components/ui/skeleton'

export function EndpointListSkeleton() {
  return (
    <div className="w-full p-2 space-y-3">
      {Array.from({ length: 6 }).map((_, gi) => (
        <div key={gi} className="space-y-1.5">
          <div className="flex items-center gap-2 px-1">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-3 w-6" />
          </div>
          {Array.from({ length: 3 }).map((_, ri) => (
            <div key={ri} className="flex items-center gap-2 px-2 py-1.5">
              <Skeleton className="h-3.5 w-10" />
              <Skeleton className="h-3 flex-1" />
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
