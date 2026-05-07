import { ChevronRight, Shield, Bell, Palette } from 'lucide-react'

export function Settings() {
  const settingsSections = [
    {
      title: 'Appearance',
      icon: Palette,
      description: 'Customize how Spectra looks',
      items: [
        { label: 'Theme', value: 'System', action: '#' },
        { label: 'Font Size', value: 'Default', action: '#' },
      ]
    },
    {
      title: 'Security',
      icon: Shield,
      description: 'Manage authentication and security',
      items: [
        { label: 'Authentication Method', value: 'Current User', action: '#' },
        { label: 'Session Timeout', value: '1 hour', action: '#' },
      ]
    },
    {
      title: 'Notifications',
      icon: Bell,
      description: 'Configure notifications',
      items: [
        { label: 'Test Notifications', value: 'On', action: '#' },
        { label: 'Error Alerts', value: 'On', action: '#' },
      ]
    },
  ]

  return (
    <div className="h-full overflow-auto">
      <div className="max-w-2xl mx-auto p-8 space-y-8">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold text-foreground">Settings</h1>
          <p className="text-muted-foreground mt-2">Manage your Spectra preferences</p>
        </div>

        {/* Settings Sections */}
        <div className="space-y-6">
          {settingsSections.map((section) => {
            const Icon = section.icon
            return (
              <div key={section.title} className="border border-border/50 rounded-lg overflow-hidden">
                {/* Section Header */}
                <div className="p-4 bg-card/30 border-b border-border/50 flex items-center gap-3">
                  <Icon className="w-5 h-5 text-muted-foreground" />
                  <div className="flex-1">
                    <h2 className="font-semibold text-foreground">{section.title}</h2>
                    <p className="text-xs text-muted-foreground mt-0.5">{section.description}</p>
                  </div>
                </div>

                {/* Section Items */}
                <div className="divide-y divide-border/50">
                  {section.items.map((item, idx) => (
                    <button
                      key={idx}
                      className="w-full px-4 py-3 flex items-center justify-between hover:bg-card/20 transition-colors text-left group"
                    >
                      <span className="text-sm font-medium text-foreground">{item.label}</span>
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-muted-foreground">{item.value}</span>
                        <ChevronRight className="w-4 h-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )
          })}
        </div>

        {/* Footer */}
        <div className="pt-8 border-t border-border/50">
          <p className="text-xs text-muted-foreground">
            Version 0.1.0 • Settings are synced automatically
          </p>
        </div>
      </div>
    </div>
  )
}
