import { useEffect, useState } from 'react'
import { FolderOpen, FolderSearch, Loader2, AlertTriangle } from 'lucide-react'
import { Drivers } from '../../../wailsjs/go/app/App'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useUIStore } from '@/store/uiStore'
import { useAddProject } from '@/hooks/useAddProject'
import { DetectionBadge } from './DetectionBadge'
import { ProjectAvatar } from './ProjectAvatar'
import { InspectionProgress } from './InspectionProgress'
import type { ProjectInfo } from '@/services/projectService'

const SUPPORTED_FRAMEWORKS = new Set(['laravel'])

export function AddProjectDialog() {
  const open = useUIStore((s) => s.isAddProjectOpen)
  const setOpen = useUIStore((s) => s.setAddProjectOpen)
  const close = () => setOpen(false)
  const {
    status,
    info,
    error,
    pipeline,
    pickFolder,
    confirm,
    reset,
  } = useAddProject(close)

  const handleOpenChange = (value: boolean) => {
    setOpen(value)
    if (!value) reset()
  }

  const isLoading = status === 'picking' || status === 'inspecting' || status === 'saving'
  const supported = info ? isSupported(info) : false
  const canConfirm = status === 'ready' && supported

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md gap-3">
        <DialogHeader>
          <DialogTitle className="text-base">Add Project</DialogTitle>
          <DialogDescription className="text-[12.5px]">
            Spectra inspects the folder and detects the framework.
          </DialogDescription>
        </DialogHeader>

        {!info && (
          <FolderPickerCard
            onPick={pickFolder}
            loading={status === 'picking'}
            label={status === 'picking' ? 'Opening...' : 'Choose folder'}
          />
        )}

        {info && status === 'inspecting' && (
          <InspectionProgress steps={pipeline} />
        )}

        {info && (status === 'ready' || status === 'saving' || status === 'error') && (
          <ProjectPreview info={info} onChange={pickFolder} disabled={isLoading} />
        )}

        {info && status === 'ready' && !supported && <UnsupportedWarning framework={info.framework} />}

        {!info && <SupportedFrameworks />}

        {error && (
          <p className="text-[12px] text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-3 py-2">
            {error}
          </p>
        )}

        <DialogFooter className="gap-2">
          <Button variant="outline" size="sm" onClick={close} disabled={status === 'saving'}>
            Cancel
          </Button>
          <Button
            size="sm"
            onClick={confirm}
            disabled={!canConfirm}
            className="min-w-[120px]"
          >
            {status === 'saving' ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                Adding...
              </>
            ) : (
              'Add Project'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function isSupported(info: ProjectInfo): boolean {
  return Boolean(info.detection?.detected) && SUPPORTED_FRAMEWORKS.has(info.framework)
}

interface FolderPickerCardProps {
  onPick: () => void
  loading: boolean
  label: string
}

function FolderPickerCard({ onPick, loading, label }: FolderPickerCardProps) {
  return (
    <button
      onClick={onPick}
      disabled={loading}
      className="w-full rounded-lg border border-dashed border-border/70 bg-card/40 hover:bg-card/70 hover:border-border transition-colors p-6 text-center disabled:opacity-60 disabled:cursor-progress"
    >
      <div className="flex flex-col items-center gap-2.5">
        <span className="inline-flex w-10 h-10 items-center justify-center rounded-md bg-primary/10 text-primary">
          {loading ? (
            <Loader2 className="w-5 h-5 animate-spin" />
          ) : (
            <FolderSearch className="w-5 h-5" />
          )}
        </span>
        <div>
          <p className="text-[13px] font-medium">{label}</p>
          <p className="text-[11px] text-muted-foreground mt-0.5">
            Pick your project root directory
          </p>
        </div>
      </div>
    </button>
  )
}

function SupportedFrameworks() {
  const [drivers, setDrivers] = useState<string[]>([])

  useEffect(() => {
    Drivers()
      .then((list) => setDrivers(list ?? []))
      .catch(() => setDrivers([]))
  }, [])

  if (drivers.length === 0) return null

  return (
    <div className="flex items-center justify-center gap-1.5 flex-wrap">
      <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground">Supported</span>
      {drivers.map((name) => (
        <span
          key={name}
          className="inline-flex items-center text-[11px] font-medium px-1.5 py-0.5 rounded border border-border/60 bg-muted/40 text-foreground/80 capitalize"
        >
          {name}
        </span>
      ))}
      <span className="text-[10.5px] text-muted-foreground/70">· more soon</span>
    </div>
  )
}

interface ProjectPreviewProps {
  info: ProjectInfo
  onChange: () => void
  disabled: boolean
}

function ProjectPreview({ info, onChange, disabled }: ProjectPreviewProps) {
  return (
    <div className="rounded-lg border border-border/60 bg-card/40 p-3.5 space-y-3">
      <div className="flex items-center gap-3">
        <ProjectAvatar name={info.name} size="md" />
        <div className="min-w-0 flex-1" title={info.path}>
          <p className="text-[13px] font-semibold truncate">{info.name}</p>
        </div>
        <button
          type="button"
          onClick={onChange}
          disabled={disabled}
          className="text-[11px] text-muted-foreground hover:text-foreground inline-flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <FolderOpen className="w-3 h-3" />
          Change
        </button>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground">Framework</span>
        <DetectionBadge
          framework={info.framework}
          detected={info.detection?.detected ?? false}
          confidence={info.detection?.confidence}
        />
      </div>
    </div>
  )
}

function UnsupportedWarning({ framework }: { framework: string }) {
  return (
    <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-[11.5px] text-amber-500">
      <AlertTriangle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
      <p className="leading-relaxed">
        <span className="font-semibold capitalize">{framework || 'Unknown'}</span> is not supported yet.
        Only Laravel projects can be added at this time.
      </p>
    </div>
  )
}
