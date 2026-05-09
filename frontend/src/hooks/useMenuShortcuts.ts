import { useEffect } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'
import { useUIStore, type PageType } from '@/store/uiStore'
import { useProjectStore } from '@/store/projectStore'
import { useEndpointsStore } from '@/store/endpointsStore'
import { useMockStore } from '@/store/mockStore'
import { SaveOpenAPIToFile } from '../../wailsjs/go/app/App'
import toast from 'react-hot-toast'

export function useMenuShortcuts() {
  useEffect(() => {
    const handlers: Array<[string, (...args: any[]) => void]> = [
      ['menu:navigate', (page: PageType) => useUIStore.getState().setCurrentPage(page)],
      ['menu:new-project', () => useUIStore.getState().setAddProjectOpen(true)],
      ['menu:sync', () => {
        const id = useProjectStore.getState().activeProjectId
        if (id) void useEndpointsStore.getState().scan(id)
      }],
      ['menu:export-openapi', async () => {
        const id = useProjectStore.getState().activeProjectId
        if (!id) return
        try {
          const path = await SaveOpenAPIToFile(id)
          if (path) toast.success(`OpenAPI saved: ${path}`)
        } catch {
          toast.error('Export failed')
        }
      }],
      ['menu:command-palette', () => useUIStore.getState().setCommandPaletteOpen(true)],
      ['menu:auth-drawer', () => useUIStore.getState().setAuthDrawerOpen(true)],
      ['menu:toggle-compact-toolbar', () => {
        const s = useUIStore.getState()
        s.setCompactToolbar(!s.compactToolbar)
      }],
      ['menu:mock-start', async () => {
        const id = useProjectStore.getState().activeProjectId
        if (!id) return
        await useMockStore.getState().start(id, 4001)
        toast.success('Mock server started')
      }],
      ['menu:mock-stop', async () => {
        await useMockStore.getState().stop()
        toast.success('Mock server stopped')
      }],
    ]
    for (const [event, fn] of handlers) EventsOn(event, fn)
    return () => {
      for (const [event] of handlers) EventsOff(event)
    }
  }, [])
}
