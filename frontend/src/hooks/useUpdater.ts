import { useCallback, useEffect, useState } from 'react'
import { CheckForUpdates, InstallUpdate, AppVersion } from '../../wailsjs/go/app/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'

export interface UpdateInfo {
  version: string
  currentVersion: string
  notes: string
}

export interface UpdateProgress {
  downloaded: number
  total: number
}

export interface UseUpdaterReturn {
  update: UpdateInfo | null
  currentVersion: string
  checking: boolean
  installing: boolean
  progress: UpdateProgress | null
  error: string | null
  check: () => Promise<void>
  install: () => Promise<void>
  dismiss: () => void
}

export function useUpdater(): UseUpdaterReturn {
  const [update, setUpdate] = useState<UpdateInfo | null>(null)
  const [currentVersion, setCurrentVersion] = useState('')
  const [checking, setChecking] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [progress, setProgress] = useState<UpdateProgress | null>(null)
  const [error, setError] = useState<string | null>(null)

  const check = useCallback(async () => {
    setChecking(true)
    setError(null)
    try {
      const result = await CheckForUpdates()
      setUpdate(result as UpdateInfo | null)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setChecking(false)
    }
  }, [])

  const install = useCallback(async () => {
    setInstalling(true)
    setError(null)
    setProgress({ downloaded: 0, total: -1 })
    try {
      await InstallUpdate()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
      setInstalling(false)
    }
  }, [])

  const dismiss = useCallback(() => {
    setUpdate(null)
  }, [])

  useEffect(() => {
    AppVersion().then(setCurrentVersion).catch(() => {})

    const onProgress = (p: UpdateProgress) => setProgress(p)
    const onError = (msg: string) => {
      setError(msg)
      setInstalling(false)
    }
    EventsOn('update:progress', onProgress)
    EventsOn('update:error', onError)

    check()

    return () => {
      EventsOff('update:progress')
      EventsOff('update:error')
    }
  }, [check])

  return {
    update,
    currentVersion,
    checking,
    installing,
    progress,
    error,
    check,
    install,
    dismiss,
  }
}
