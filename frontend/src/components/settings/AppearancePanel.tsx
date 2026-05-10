import { Palette } from 'lucide-react'
import { useTheme } from '@/hooks/useTheme'
import { useUIStore } from '@/store/uiStore'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { cn } from '@/lib/utils'

export function AppearancePanel() {
  const theme = useTheme((s) => s.theme)
  const setTheme = useTheme((s) => s.setTheme)
  const compactToolbar = useUIStore((s) => s.compactToolbar)
  const setCompactToolbar = useUIStore((s) => s.setCompactToolbar)

  return (
    <div>
      <SettingsHeader
        icon={Palette}
        title="Appearance"
        description="Match Spectra to your editor and OS theme."
      />

      <SettingsCard>
        <SettingsRow
          label="Theme"
          description="Light, dark, or follow your system preference."
          control={
            <div className="inline-flex items-center gap-1 rounded-md border border-border/40 p-0.5">
              {(['light', 'dark', 'system'] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setTheme(t)}
                  className={cn(
                    'px-2.5 py-0.5 text-[11px] rounded capitalize',
                    theme === t
                      ? 'bg-accent text-foreground'
                      : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          }
        />

        <SettingsRow
          label="Compact toolbar"
          description="Collapse Inspector toolbar pills to icon-only. Tooltip on hover."
          control={
            <button
              type="button"
              onClick={() => setCompactToolbar(!compactToolbar)}
              className={cn(
                'h-6 w-11 rounded-full transition-colors relative',
                compactToolbar ? 'bg-primary' : 'bg-muted-foreground/30',
              )}
            >
              <span
                className={cn(
                  'absolute top-0.5 h-5 w-5 rounded-full bg-background shadow transition-all',
                  compactToolbar ? 'left-[22px]' : 'left-0.5',
                )}
              />
            </button>
          }
        />
      </SettingsCard>
    </div>
  )
}
