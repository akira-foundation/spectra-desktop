import { useEffect } from 'react'
import { useProjectStore } from '@/store/projectStore'
import { useUIStore } from '@/store/uiStore'
import { useMenuShortcuts } from '@/hooks/useMenuShortcuts'
import { useUpdatesStore } from '@/store/updatesStore'
import { AppShell } from '@/components/app/AppShell'
import { Dashboard } from '@/components/pages/Dashboard'
import { APIInspector } from '@/components/pages/APIInspector'
import { Settings } from '@/components/pages/Settings'
import { Changelog } from '@/components/pages/Changelog'
import { CollectionsPage } from '@/components/pages/CollectionsPage'
import { ScratchPage } from '@/components/pages/ScratchPage'
import { AccountsPage } from '@/components/pages/AccountsPage'
import { MockPage } from '@/components/pages/MockPage'
import { EmptyWorkspace } from '@/components/projects/EmptyWorkspace'
// import { OnboardingFlow } from '@/components/onboarding'
import '@/styles/globals.css'

function App() {
  const loadFromStorage = useProjectStore((state) => state.loadFromStorage)
  const projects = useProjectStore((state) => state.projects)
  const isLoading = useProjectStore((state) => state.isLoading)
  const currentPage = useUIStore((state) => state.currentPage)

  const initUpdates = useUpdatesStore((s) => s.init)
  useEffect(() => {
    void loadFromStorage()
    void initUpdates()
  }, [loadFromStorage, initUpdates])

  useMenuShortcuts()

  const noProjects = !isLoading && projects.length === 0
  const pageNeedsProject = currentPage !== 'settings'

  return (
    <AppShell>
      {noProjects && pageNeedsProject ? (
        <EmptyWorkspace />
      ) : (
        <>
          {currentPage === 'inspector' && <APIInspector />}
          {currentPage === 'collections' && <CollectionsPage />}
          {currentPage === 'scratch' && <ScratchPage />}
          {currentPage === 'accounts' && <AccountsPage />}
          {currentPage === 'mock' && <MockPage />}
          {currentPage === 'dashboard' && <Dashboard />}
          {currentPage === 'changelog' && <Changelog />}
          {currentPage === 'settings' && <Settings />}
        </>
      )}
    </AppShell>
  )
}

export default App
