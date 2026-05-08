import { useCallback, useMemo, useState } from 'react'
import {
  projectService,
  type ProjectInfo,
  type APIDetection,
  type APIFilterMode,
} from '@/services/projectService'
import { useProjectStore } from '@/store/projectStore'
import { projectInputFromInfo } from '@/lib/project-factory'
import { useInspectionPipeline, delay, type InspectionStep } from './useInspectionPipeline'

export type AddProjectStatus = 'idle' | 'picking' | 'inspecting' | 'ready' | 'saving' | 'error'

export interface AddProjectState {
  status: AddProjectStatus
  info: ProjectInfo | null
  detection: APIDetection | null
  filterMode: APIFilterMode
  filterValue: string
  baseUrl: string
  previewing: boolean
  error: string | null
  pipeline: ReturnType<typeof useInspectionPipeline>['state']
  pipelineRunning: boolean
  pickFolder: () => Promise<void>
  setFilterMode: (mode: APIFilterMode) => void
  setFilterValue: (value: string) => void
  setBaseUrl: (value: string) => void
  applyFilter: () => Promise<void>
  confirm: () => Promise<void>
  reset: () => void
}

export function useAddProject(onSuccess?: () => void): AddProjectState {
  const addProjectFromInput = useProjectStore((s) => s.addProjectFromInput)
  const [status, setStatus] = useState<AddProjectStatus>('idle')
  const [info, setInfo] = useState<ProjectInfo | null>(null)
  const [detection, setDetection] = useState<APIDetection | null>(null)
  const [filterMode, setFilterMode] = useState<APIFilterMode>('auto')
  const [filterValue, setFilterValue] = useState<string>('')
  const [baseUrl, setBaseUrl] = useState<string>('')
  const [previewing, setPreviewing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const inspectStepsRef = useMemo<InspectionStep[]>(
    () => [
      { id: 'detect', label: 'Detect framework', run: () => delay(120) },
      { id: 'routes', label: 'Scan routes', run: () => delay(420) },
      { id: 'filter', label: 'Filter API routes', run: () => delay(280) },
      { id: 'middleware', label: 'Resolve middleware', run: () => delay(260) },
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
    setDetection(null)
    setFilterMode('auto')
    setFilterValue('')
    setBaseUrl('')
    setPreviewing(false)
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
      setDetection(result.apiDetection ?? null)
      const resolvedMode = (result.apiDetection?.mode ?? 'auto') as APIFilterMode
      setFilterMode(resolvedMode === 'auto' ? 'auto' : resolvedMode)
      setFilterValue(result.apiDetection?.value ?? '')
      setBaseUrl(result.defaultBaseUrl ?? '')
      setStatus('inspecting')
      void pipeline.run()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setStatus('error')
    }
  }, [pipeline])

  const applyFilter = useCallback(async () => {
    if (!info) return
    setPreviewing(true)
    setError(null)
    try {
      const result = await projectService.previewRoutes(info.path, filterMode, filterValue)
      setDetection(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setPreviewing(false)
    }
  }, [info, filterMode, filterValue])

  const confirm = useCallback(async () => {
    if (!info) return
    setStatus('saving')
    setError(null)
    try {
      const base = projectInputFromInfo(info)
      const effectiveMode = detection?.mode ?? filterMode
      const effectiveValue = detection?.value ?? filterValue
      await addProjectFromInput({
        ...base,
        apiFilterMode: effectiveMode,
        apiFilterValue: effectiveValue,
        baseUrl: baseUrl.trim() || base.baseUrl,
      })
      onSuccess?.()
      reset()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setStatus('ready')
    }
  }, [info, detection, filterMode, filterValue, addProjectFromInput, onSuccess, reset])

  return {
    status,
    info,
    detection,
    filterMode,
    filterValue,
    baseUrl,
    previewing,
    error,
    pipeline: pipeline.state,
    pipelineRunning: pipeline.running,
    pickFolder,
    setFilterMode,
    setFilterValue,
    setBaseUrl,
    applyFilter,
    confirm,
    reset,
  }
}
