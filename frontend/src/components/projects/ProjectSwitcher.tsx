import { ChevronsUpDown, Plus } from 'lucide-react'
import type { Project } from '@/types/project'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ProjectAvatar } from './ProjectAvatar'

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
        <button className="inline-flex items-center gap-1.5 h-7 px-1.5 rounded-md hover:bg-accent/60 transition-colors">
          <ProjectAvatar name={activeProject.name} />
          <span className="text-[12px] font-semibold tracking-tight truncate max-w-[160px]">
            {activeProject.name}
          </span>
          <ChevronsUpDown className="w-3 h-3 text-muted-foreground" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[14rem]">
        <DropdownMenuLabel className="text-[10.5px] uppercase tracking-wider text-muted-foreground">
          Projects
        </DropdownMenuLabel>
        {projects.map((p) => (
          <DropdownMenuItem
            key={p.id}
            onSelect={() => onSelect(p.id)}
            className="gap-2 text-[12.5px]"
          >
            <ProjectAvatar name={p.name} />
            <span className="font-medium truncate">{p.name}</span>
            {p.id === activeProject.id && (
              <span className="ml-auto text-[10px] text-muted-foreground">active</span>
            )}
          </DropdownMenuItem>
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
