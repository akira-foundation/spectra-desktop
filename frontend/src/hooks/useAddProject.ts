import { useCallback, useMemo, useState } from 'react'
import { projectService, type ProjectInfo } from '@/services/projectService'
import { useProjectStore } from '@/store/projectStore'
import { projectInputFromInfo } from '@/lib/project-factory'
import { useInspectionPipeline, delay, type InspectionStep } from './useInspectionPipeline'

export type AddProjectStatus = 'idle' | 'picking' | 'inspecting' | 'ready' | 'saving' | 'error'

export interface AddProjectState {
  status: AddProjectStatus
  info: ProjectInfo | null
  error: string | null
  pipeline: ReturnType<typeof useInspectionPipeline>['state']
  pipelineRunning: boolean
  pickFolder: () => Promise<void>
  confirm: () => Promise<void>
  reset: () => void
}

export function useAddProject(onSuccess?: () => void): AddProjectState {
  const addProjectFromInput = useProjectStore((s) => s.addProjectFromInput)
  const [status, setStatus] = useState<AddProjectStatus>('idle')
  const [info, setInfo] = useState<ProjectInfo | null>(null)
  const [error, setError] = useState<string | null>(null)

  const inspectStepsRef = useMemo<InspectionStep[]>(
    () => [
      {
        id: 'detect',
        label: 'Detect framework',
        run: async () => {
          // handled outside pipeline (real backend call done in pickFolder)
          await delay(120)
        },
      },
      { id: 'routes', label: 'Scan routes', run: () => delay(420) },
      { id: 'middleware', label: 'Resolve middleware', run: () => delay(320) },
      { id: 'controllers', label: 'Map controllers', run: () => delay(360) },
    ],
    [],
  )

  const pipeline = useInspectionPipeline(inspectStepsRef, {
    onComplete: () => setStatus('ready'),
    onError: (err) => {
      setError(err instanceof Error ? err.message : String(err))
      setStatus('error')
    },
  })

  const reset = useCallback(() => {
    setStatus('idle')
    setInfo(null)
    setError(null)
    pipeline.reset()
  }, [pipeline])

  const pickFolder = useCallback(async () => {
    setStatus('picking')
    setError(null)
    try {
      const path = await projectService.pickFolder()
      if (!path) {
        setStatus('idle')
        return
      }
      const result = await projectService.inspect(path)
      setInfo(result)
      setStatus('inspecting')
      void pipeline.run()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setStatus('error')
    }
  }, [pipeline])

  const confirm = useCallback(async () => {
    if (!info) return
    setStatus('saving')
    setError(null)
    try {
      await addProjectFromInput(projectInputFromInfo(info))
      onSuccess?.()
      reset()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setStatus('ready')
    }
  }, [info, addProjectFromInput, onSuccess, reset])

  return {
    status,
    info,
    error,
    pipeline: pipeline.state,
    pipelineRunning: pipeline.running,
    pickFolder,
    confirm,
    reset,
  }
}
