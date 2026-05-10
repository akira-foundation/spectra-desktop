import { useEffect, useState } from 'react'
import { Palette, Info, Package, Database } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { Island, IslandBody } from '@/components/app/Island'
import {
  SettingsSidebar,
  type SettingsNavGroup,
} from '@/components/settings/SettingsSidebar'
import { AppearancePanel } from '@/components/settings/AppearancePanel'
import { ArchivesPanel } from '@/components/settings/ArchivesPanel'
import { DatabasePanel } from '@/components/settings/DatabasePanel'
import { AboutPanel } from '@/components/settings/AboutPanel'

type SectionId = 'appearance' | 'archives' | 'database' | 'about'

const NAV: SettingsNavGroup[] = [
  {
    heading: 'Application',
    items: [
      { id: 'appearance', label: 'Appearance', icon: Palette },
      { id: 'about', label: 'About', icon: Info },
    ],
  },
  {
    heading: 'Data',
    items: [
      { id: 'archives', label: 'Project archives', icon: Package },
      { id: 'database', label: 'Database backup', icon: Database },
    ],
  },
]

export function Settings() {
  const [section, setSection] = useState<SectionId>('appearance')
  const pendingAction = useUIStore((s) => s.pendingArchiveAction)
  const clearPending = useUIStore((s) => s.setPendingArchiveAction)

  useEffect(() => {
    if (!pendingAction) return
    if (pendingAction === 'export' || pendingAction === 'import') {
      setSection('archives')
    } else if (pendingAction === 'backup' || pendingAction === 'restore') {
      setSection('database')
    }
    // pending action is consumed by the destination panel.
    // We do NOT clear it here so the panel's effect can act on it.
    void clearPending
  }, [pendingAction, clearPending])

  return (
    <div className="h-full flex gap-2 p-2 min-h-0">
      <Island as="aside" className="w-64 shrink-0">
        <SettingsSidebar
          groups={NAV}
          activeId={section}
          onSelect={(id) => setSection(id as SectionId)}
        />
      </Island>
      <Island as="main" className="flex-1">
        <IslandBody>
          <div className="max-w-2xl mx-auto px-10 py-10">
            {section === 'appearance' && <AppearancePanel />}
            {section === 'archives' && <ArchivesPanel />}
            {section === 'database' && <DatabasePanel />}
            {section === 'about' && <AboutPanel />}
          </div>
        </IslandBody>
      </Island>
    </div>
  )
}
