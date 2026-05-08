import { FolderPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useUIStore } from '@/store/uiStore'

export function EmptyWorkspace() {
  const setAddProjectOpen = useUIStore((s) => s.setAddProjectOpen)

  return (
    <div className="flex h-full w-full items-center justify-center p-8">
      <div className="max-w-sm w-full text-center space-y-4">
        <div className="inline-flex w-12 h-12 items-center justify-center rounded-xl bg-primary/10 ring-1 ring-primary/20 text-primary">
          <FolderPlus className="w-6 h-6" />
        </div>
        <div className="space-y-1">
          <h2 className="text-[15px] font-semibold tracking-tight">No project yet</h2>
          <p className="text-[12.5px] text-muted-foreground leading-relaxed">
            Add a backend folder to start exploring its API. Spectra detects the framework
            automatically.
          </p>
        </div>
        <Button size="sm" onClick={() => setAddProjectOpen(true)}>
          <FolderPlus className="w-3.5 h-3.5" />
          Add Project
        </Button>
        <p className="text-[10.5px] text-muted-foreground/70">
          Currently supports Laravel · more frameworks coming
        </p>
      </div>
    </div>
  )
}
