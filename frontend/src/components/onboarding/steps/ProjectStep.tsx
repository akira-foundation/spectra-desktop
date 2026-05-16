import { FolderPlus, CheckCircle2 } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'

interface Props {
  onAdded: () => void
  onSkip: () => void
}

export function ProjectStep({ onAdded, onSkip }: Props) {
  const setAddProjectOpen = useUIStore((s) => s.setAddProjectOpen)
  const projects = useProjectStore((s) => s.projects)
  const openPicker = () => setAddProjectOpen(true)

  return (
    <div className="w-full max-w-sm mx-auto">
      <header className="text-center mb-8">
        <h1 className="text-[28px] font-semibold tracking-tight leading-tight">Add a project</h1>
        <p className="mt-2 text-[13.5px] text-muted-foreground">
          Point Spectra at a backend folder. Framework and routes detected automatically.
        </p>
      </header>

      <button
        type="button"
        onClick={openPicker}
        className="w-full rounded-xl border border-dashed border-border/60 bg-card/30 px-6 py-8 hover:bg-accent/30 hover:border-border transition-colors text-center"
      >
        <div className="h-12 w-12 rounded-2xl bg-primary/15 border border-primary/30 flex items-center justify-center mx-auto">
          <FolderPlus className="h-5 w-5 text-primary" strokeWidth={1.75} />
        </div>
        <p className="mt-4 text-[14px] font-medium">Choose folder</p>
        <p className="mt-1 text-[12px] text-muted-foreground">
          Laravel today. More frameworks soon.
        </p>
      </button>

      {projects.length > 0 && (
        <button
          type="button"
          onClick={onAdded}
          className="mt-4 w-full h-11 rounded-xl bg-foreground text-background hover:bg-foreground/90 transition-colors text-[13.5px] font-medium inline-flex items-center justify-center gap-2"
        >
          <CheckCircle2 className="h-4 w-4" />
          Continue · {projects.length} {projects.length === 1 ? 'project' : 'projects'}
        </button>
      )}

      <div className="mt-10 text-center">
        <button
          type="button"
          onClick={onSkip}
          className="text-[12px] text-muted-foreground/80 hover:text-foreground transition-colors"
        >
          I'll add one later
        </button>
      </div>
    </div>
  )
}
