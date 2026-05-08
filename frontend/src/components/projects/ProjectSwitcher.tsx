import { Check, ChevronsUpDown, Plus } from 'lucide-react'
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
  if (!activeProject) {
    return <AddProjectButton onClick={onAddProject} />
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="inline-flex items-center gap-1.5 h-7 px-1.5 rounded-md hover:bg-accent/60 transition-colors outline-none focus:outline-none focus-visible:outline-none focus-visible:ring-0 data-[state=open]:bg-accent/60">
          <ProjectAvatar name={activeProject.name} />
          <span className="text-[12px] font-semibold tracking-tight truncate max-w-[160px]">
            {activeProject.name}
          </span>
          <FrameworkChip framework={activeProject.framework} />
          <ChevronsUpDown className="w-3 h-3 text-muted-foreground" />
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
  )
}

interface ProjectRowProps {
  project: Project
  active: boolean
  onSelect: () => void
}

function ProjectRow({ project, active, onSelect }: ProjectRowProps) {
  return (
    <Tooltip delayDuration={400}>
      <TooltipTrigger asChild>
        <DropdownMenuItem
          onSelect={onSelect}
          className={cn('gap-2 text-[12.5px] py-1.5', active && 'bg-accent/40')}
        >
          <ProjectAvatar name={project.name} />
          <span className="font-medium truncate flex-1 min-w-0">{project.name}</span>
          <div className="flex items-center gap-1.5 shrink-0">
            <FrameworkChip framework={project.framework} />
            {active && <Check className="w-3.5 h-3.5 text-primary" />}
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
    <span className="inline-flex items-center text-[10px] font-medium px-1.5 py-0.5 rounded border border-border/60 bg-muted/40 text-foreground/70 capitalize">
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
