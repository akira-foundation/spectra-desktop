import { useCallback, useState } from 'react'
import { projectService, type ProjectInfo } from '@/services/projectService'
import { useProjectStore } from '@/store/projectStore'
import { projectInputFromInfo } from '@/lib/project-factory'

export type AddProjectStatus = 'idle' | 'picking' | 'inspecting' | 'saving' | 'ready' | 'error'

export interface AddProjectState {
  status: AddProjectStatus
  info: ProjectInfo | null
  error: string | null
  pickFolder: () => Promise<void>
  setPath: (path: string) => Promise<void>
  confirm: () => Promise<void>
  reset: () => void
}

export function useAddProject(onSuccess?: () => void): AddProjectState {
  const addProjectFromInput = useProjectStore((s) => s.addProjectFromInput)
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
      setError(toMessage(err))
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
      setError(toMessage(err))
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

  const confirm = useCallback(async () => {
    if (!info) return
    setStatus('saving')
    setError(null)
    try {
      await addProjectFromInput(projectInputFromInfo(info))
      onSuccess?.()
      reset()
    } catch (err) {
      setError(toMessage(err))
      setStatus('ready')
    }
  }, [info, addProjectFromInput, onSuccess, reset])

  return { status, info, error, pickFolder, setPath, confirm, reset }
}

function toMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
