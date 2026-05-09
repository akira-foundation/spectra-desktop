export type DashboardTab = 'overview' | 'discovery' | 'activity'

interface Props {
  active: DashboardTab
  onChange: (t: DashboardTab) => void
}

export function DashboardTabs({ active, onChange }: Props) {
  const tabs: { v: DashboardTab; label: string }[] = [
    { v: 'overview', label: 'Overview' },
    { v: 'discovery', label: 'Discovery' },
    { v: 'activity', label: 'Activity' },
  ]
  return (
    <div className="flex items-center gap-4 border-b border-border/40">
      {tabs.map((t) => (
        <button
          key={t.v}
          type="button"
          onClick={() => onChange(t.v)}
          className={`text-[12px] font-medium px-0 pb-2 -mb-px border-b-2 transition-colors ${
            active === t.v
              ? 'border-primary text-foreground'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          {t.label}
        </button>
      ))}
    </div>
  )
}
