import { useState } from 'react'
import { Check, ChevronsUpDown, Plus, Trash2 } from 'lucide-react'
import type { Project } from '@/types/project'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { useProjectStore } from '@/store/projectStore'
import { ProjectAvatar } from './ProjectAvatar'
import { cn } from '@/lib/utils'

interface ProjectSwitcherProps {
  projects: Project[]
  activeProject?: Project
  onSelect: (id: string) => void
  onAddProject: () => void
}

export function ProjectSwitcher({
  projects,
  activeProject,
  onSelect,
  onAddProject,
}: ProjectSwitcherProps) {
  const removeProject = useProjectStore((s) => s.removeProject)
  const [pendingDelete, setPendingDelete] = useState<Project | null>(null)

  if (!activeProject) {
    return <AddProjectButton onClick={onAddProject} />
  }

  const handleConfirmDelete = async () => {
    if (!pendingDelete) return
    await removeProject(pendingDelete.id)
    setPendingDelete(null)
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button className="inline-flex items-center gap-1.5 h-7 px-1.5 rounded-md hover:bg-foreground/10 dark:hover:bg-white/10 transition-colors outline-none focus:outline-none focus-visible:outline-none focus-visible:ring-0 data-[state=open]:bg-foreground/10 dark:data-[state=open]:bg-white/10 text-foreground dark:text-white/90">
            <ProjectAvatar name={activeProject.name} />
            <span className="text-[12px] font-semibold tracking-tight truncate max-w-[160px]">
              {activeProject.name}
            </span>
            <FrameworkChip framework={activeProject.framework} />
            <ChevronsUpDown className="w-3 h-3 text-muted-foreground dark:text-white/60" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="min-w-[18rem]">
          <DropdownMenuLabel className="text-[10.5px] uppercase tracking-wider text-muted-foreground">
            Projects
          </DropdownMenuLabel>
          {projects.map((p) => (
            <ProjectRow
              key={p.id}
              project={p}
              active={p.id === activeProject.id}
              onSelect={() => onSelect(p.id)}
              onDelete={() => setPendingDelete(p)}
            />
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onSelect={onAddProject}
            className="gap-2 text-[12.5px] text-emerald-500 focus:text-emerald-500"
          >
            <Plus className="w-3.5 h-3.5" />
            <span className="font-medium">Add Project</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ConfirmDialog
        open={!!pendingDelete}
        onOpenChange={(v) => !v && setPendingDelete(null)}
        variant="destructive"
        title={`Remove ${pendingDelete?.name ?? 'project'}?`}
        description="This unlinks the project from Spectra. The folder on disk is not touched."
        confirmLabel="Remove"
        onConfirm={handleConfirmDelete}
      />
    </>
  )
}

interface ProjectRowProps {
  project: Project
  active: boolean
  onSelect: () => void
  onDelete: () => void
}

function ProjectRow({ project, active, onSelect, onDelete }: ProjectRowProps) {
  return (
    <Tooltip delayDuration={400}>
      <TooltipTrigger asChild>
        <DropdownMenuItem
          onSelect={onSelect}
          className={cn('group gap-2 text-[12.5px] py-1.5 pr-1.5', active && 'bg-accent/40')}
        >
          <ProjectAvatar name={project.name} />
          <span className="font-medium truncate flex-1 min-w-0">{project.name}</span>
          <div className="flex items-center gap-1 shrink-0">
            <FrameworkChip framework={project.framework} />
            {active && <Check className="w-3.5 h-3.5 text-primary shrink-0" />}
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
                onDelete()
              }}
              className="opacity-0 group-hover:opacity-100 transition-opacity inline-flex items-center justify-center w-5 h-5 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10"
              aria-label="Remove project"
            >
              <Trash2 className="w-3 h-3" />
            </button>
          </div>
        </DropdownMenuItem>
      </TooltipTrigger>
      <TooltipContent side="right" className="font-mono">
        {project.path}
      </TooltipContent>
    </Tooltip>
  )
}

function FrameworkChip({ framework }: { framework: string }) {
  if (!framework || framework === 'other') return null
  return (
    <span className="inline-flex items-center text-[10px] font-medium px-1.5 py-0.5 rounded border border-foreground/15 dark:border-white/15 bg-foreground/10 dark:bg-white/10 text-foreground/85 dark:text-white/85 capitalize">
      {framework}
    </span>
  )
}

interface AddProjectButtonProps {
  onClick: () => void
}

export function AddProjectButton({ onClick }: AddProjectButtonProps) {
  return (
    <button
      onClick={onClick}
      className="inline-flex items-center gap-1.5 h-7 px-2 rounded-md border border-dashed border-border/70 text-muted-foreground hover:text-foreground hover:bg-accent/40 hover:border-border transition-colors"
    >
      <Plus className="w-3.5 h-3.5" />
      <span className="text-[11.5px] font-medium">Add Project</span>
    </button>
  )
}
