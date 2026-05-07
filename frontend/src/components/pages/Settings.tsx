import { ChevronRight, Shield, Bell, Palette } from 'lucide-react'

export function Settings() {
  const sections = [
    {
      title: 'Appearance',
      icon: Palette,
      description: 'Customize how Spectra looks',
      items: [
        { label: 'Theme', value: 'System' },
        { label: 'Font size', value: 'Default' },
      ],
    },
    {
      title: 'Security',
      icon: Shield,
      description: 'Manage authentication and security',
      items: [
        { label: 'Authentication method', value: 'Current User' },
        { label: 'Session timeout', value: '1 hour' },
      ],
    },
    {
      title: 'Notifications',
      icon: Bell,
      description: 'Configure notifications',
      items: [
        { label: 'Test notifications', value: 'On' },
        { label: 'Error alerts', value: 'On' },
      ],
    },
  ]

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-2xl mx-auto p-6 space-y-6">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Settings</h1>
          <p className="text-muted-foreground text-[12.5px] mt-1">
            Manage your Spectra preferences
          </p>
        </div>

        <div className="space-y-4">
          {sections.map(({ title, description, icon: Icon, items }) => (
            <section key={title} className="border border-border/60 rounded-lg overflow-hidden bg-card/40">
              <header className="px-3.5 py-2.5 bg-card/60 border-b border-border/50 flex items-center gap-2.5">
                <Icon className="w-4 h-4 text-muted-foreground" />
                <div className="flex-1">
                  <h2 className="font-semibold text-[12.5px]">{title}</h2>
                  <p className="text-[11px] text-muted-foreground mt-0.5">{description}</p>
                </div>
              </header>
              <div className="divide-y divide-border/40">
                {items.map((item, idx) => (
                  <button
                    key={idx}
                    className="w-full px-3.5 py-2.5 flex items-center justify-between hover:bg-card/60 transition-colors text-left group"
                  >
                    <span className="text-[12.5px] font-medium">{item.label}</span>
                    <div className="flex items-center gap-1.5">
                      <span className="text-[11.5px] text-muted-foreground">{item.value}</span>
                      <ChevronRight className="w-3.5 h-3.5 text-muted-foreground group-hover:text-foreground transition-colors" />
                    </div>
                  </button>
                ))}
              </div>
            </section>
          ))}
        </div>

        <div className="pt-4 border-t border-border/50">
          <p className="text-[11px] text-muted-foreground">
            Version 0.1.0 · Settings sync automatically
          </p>
        </div>
      </div>
    </div>
  )
}
