import { useEffect } from 'react'
// import { useState } from 'react'
import { useProjectStore } from '@/store/projectStore'
import { useUIStore } from '@/store/uiStore'
import { AppShell } from '@/components/app/AppShell'
import { Dashboard } from '@/components/pages/Dashboard'
// import { Welcome } from '@/components/pages/Welcome'
import { APIInspector } from '@/components/pages/APIInspector'
import { Settings } from '@/components/pages/Settings'
// import { OnboardingFlow } from '@/components/onboarding'
import '@/styles/globals.css'

function App() {
  const loadFromStorage = useProjectStore((state) => state.loadFromStorage)
  const currentPage = useUIStore((state) => state.currentPage)

  // const [showOnboarding, setShowOnboarding] = useState(() => {
  //   return localStorage.getItem('spectra-onboarded') !== 'true'
  // })

  useEffect(() => {
    loadFromStorage()
  }, [loadFromStorage])

  // const handleOnboardingComplete = () => {
  //   setShowOnboarding(false)
  // }

  // if (showOnboarding) {
  //   return <OnboardingFlow onComplete={handleOnboardingComplete} />
  // }

  return (
    <AppShell>
      {currentPage === 'inspector' && <APIInspector />}
      {currentPage === 'dashboard' && <Dashboard />}
      {currentPage === 'settings' && <Settings />}
    </AppShell>
  )
}

export default App
