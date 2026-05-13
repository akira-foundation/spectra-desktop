import { create } from 'zustand'
import {
  CheckForUpdates,
  InstallUpdate,
  AppVersion,
} from '../../wailsjs/go/app/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'

export type UpdatePhase = 'idle' | 'checking' | 'available' | 'downloading' | 'ready' | 'error'

export interface UpdateInfo {
  version: string
  currentVersion: string
  notes: string
}

interface UpdatesState {
  phase: UpdatePhase
  info: UpdateInfo | null
  currentVersion: string
  progress: { downloaded: number; total: number } | null
  error: string | null
  lastCheckedAt: string | null
  initialized: boolean
  unsubscribe: (() => void) | null

  init: () => Promise<void>
  check: () => Promise<void>
  install: () => Promise<void>
  dismiss: () => void
}

export const useUpdatesStore = create<UpdatesState>((set, get) => ({
  phase: 'idle',
  info: null,
  currentVersion: '',
  progress: null,
  error: null,
  lastCheckedAt: null,
  initialized: false,
  unsubscribe: null,

  init: async () => {
    if (get().initialized) return
    set({ initialized: true })

    const onProgress = (p: { downloaded: number; total: number }) => {
      set({
        progress: p,
        phase: p.total > 0 && p.downloaded >= p.total ? 'ready' : 'downloading',
      })
    }
    const onError = (msg: string) => {
      set({ error: msg, phase: 'error' })
    }
    const onInstalled = () => {
      set({ phase: 'ready' })
    }
    EventsOn('update:progress', onProgress)
    EventsOn('update:error', onError)
    EventsOn('update:installed', onInstalled)

    set({
      unsubscribe: () => {
        EventsOff('update:progress')
        EventsOff('update:error')
        EventsOff('update:installed')
      },
    })

    try {
      const version = await AppVersion()
      set({ currentVersion: version })
    } catch {}

    await get().check()
  },

  check: async () => {
    set({ phase: 'checking', error: null })
    const started = Date.now()
    try {
      const result = (await CheckForUpdates()) as UpdateInfo | null
      const elapsed = Date.now() - started
      if (elapsed < 600) {
        await new Promise((r) => setTimeout(r, 600 - elapsed))
      }
      set({
        info: result ?? null,
        phase: result ? 'available' : 'idle',
        lastCheckedAt: new Date().toISOString(),
      })
    } catch (err) {
      const elapsed = Date.now() - started
      if (elapsed < 600) {
        await new Promise((r) => setTimeout(r, 600 - elapsed))
      }
      set({
        error: err instanceof Error ? err.message : String(err),
        phase: 'error',
        lastCheckedAt: new Date().toISOString(),
      })
    }
  },

  install: async () => {
    set({ phase: 'downloading', progress: { downloaded: 0, total: -1 }, error: null })
    try {
      await InstallUpdate()
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : String(err),
        phase: 'error',
      })
    }
  },

  dismiss: () => {
    set({ info: null, phase: 'idle' })
  },
}))

export function isUpdateActionable(phase: UpdatePhase): boolean {
  return phase === 'available' || phase === 'ready'
}
