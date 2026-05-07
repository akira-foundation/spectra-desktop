import { useCallback, useState } from 'react'
import { projectService, type ProjectInfo } from '@/services/projectService'
import { useProjectStore } from '@/store/projectStore'
import { projectFromInfo } from '@/lib/project-factory'

export type AddProjectStatus = 'idle' | 'picking' | 'inspecting' | 'ready' | 'error'

export interface AddProjectState {
  status: AddProjectStatus
  info: ProjectInfo | null
  error: string | null
  pickFolder: () => Promise<void>
  setPath: (path: string) => Promise<void>
  confirm: () => void
  reset: () => void
}

export function useAddProject(onSuccess?: () => void): AddProjectState {
  const addProject = useProjectStore((s) => s.addProject)
  const [status, setStatus] = useState<AddProjectStatus>('idle')
  const [info, setInfo] = useState<ProjectInfo | null>(null)
  const [error, setError] = useState<string | null>(null)

  const reset = useCallback(() => {
    setStatus('idle')
    setInfo(null)
    setError(null)
  }, [])

  const inspect = useCallback(async (path: string) => {
    setStatus('inspecting')
    setError(null)
    try {
      const result = await projectService.inspect(path)
      setInfo(result)
      setStatus('ready')
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setStatus('error')
    }
  }, [])

  const pickFolder = useCallback(async () => {
    setStatus('picking')
    setError(null)
    try {
      const path = await projectService.pickFolder()
      if (!path) {
        setStatus('idle')
        return
      }
      await inspect(path)
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setStatus('error')
    }
  }, [inspect])

  const setPath = useCallback(
    async (path: string) => {
      if (!path.trim()) return
      await inspect(path.trim())
    },
    [inspect],
  )

  const confirm = useCallback(() => {
    if (!info) return
    addProject(projectFromInfo(info))
    onSuccess?.()
    reset()
  }, [info, addProject, onSuccess, reset])

  return { status, info, error, pickFolder, setPath, confirm, reset }
}
