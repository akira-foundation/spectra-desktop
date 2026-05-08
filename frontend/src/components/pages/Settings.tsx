import { Palette, Folder, Info } from 'lucide-react'
import { useTheme } from '@/hooks/useTheme'
import { useProjectStore } from '@/store/projectStore'
import { cn } from '@/lib/utils'

export function Settings() {
  const theme = useTheme((s) => s.theme)
  const setTheme = useTheme((s) => s.setTheme)
  const projects = useProjectStore((s) => s.projects)

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-2xl mx-auto p-6 space-y-4">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Settings</h1>
          <p className="text-muted-foreground text-[12.5px] mt-1">
            Spectra preferences
          </p>
        </div>

        <Section title="Appearance" icon={Palette}>
          <div className="px-3.5 py-2.5 flex items-center justify-between text-[12.5px]">
            <span className="text-foreground/80">Theme</span>
            <div className="inline-flex items-center gap-1 rounded-md border border-border/40 p-0.5">
              {(['light', 'dark', 'system'] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setTheme(t)}
                  className={cn(
                    'px-2 py-0.5 text-[11px] rounded capitalize',
                    theme === t ? 'bg-accent text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>
        </Section>

        <Section title="Projects" icon={Folder}>
          {projects.length === 0 ? (
            <p className="px-3.5 py-3 text-[11.5px] italic text-muted-foreground">
              No projects yet.
            </p>
          ) : (
            <ul className="divide-y divide-border/40">
              {projects.map((p) => (
                <li
                  key={p.id}
                  className="px-3.5 py-2 flex items-center justify-between gap-3 text-[12px]"
                >
                  <div className="min-w-0">
                    <p className="font-medium truncate capitalize">{p.name}</p>
                    <p className="text-[10.5px] text-muted-foreground font-mono truncate">
                      {p.path}
                    </p>
                  </div>
                  <span className="text-[10.5px] uppercase tracking-wider text-muted-foreground shrink-0">
                    {p.framework}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Section>

        <Section title="About" icon={Info}>
          <div className="px-3.5 py-2.5 space-y-1 text-[12px]">
            <Row label="Version" value="0.1.0" />
            <Row label="Build" value="local" />
            <Row label="License" value="Beta · all features unlocked" />
          </div>
        </Section>
      </div>
    </div>
  )
}

interface SectionProps {
  title: string
  icon: React.ComponentType<{ className?: string }>
  children: React.ReactNode
}

function Section({ title, icon: Icon, children }: SectionProps) {
  return (
    <section className="border border-border/60 rounded-lg overflow-hidden bg-card/40">
      <header className="px-3.5 py-2 bg-card/60 border-b border-border/50 flex items-center gap-2">
        <Icon className="w-3.5 h-3.5 text-muted-foreground" />
        <h2 className="font-semibold text-[11.5px] uppercase tracking-wider text-muted-foreground">
          {title}
        </h2>
      </header>
      {children}
    </section>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-foreground/80">{label}</span>
      <span className="text-muted-foreground font-mono">{value}</span>
    </div>
  )
}
