interface Props {
  label: string
  description?: string
  control: React.ReactNode
}

export function SettingsRow({ label, description, control }: Props) {
  return (
    <div className="px-4 py-3.5 flex items-start justify-between gap-4">
      <div className="min-w-0">
        <p className="text-[13px] text-foreground/85 font-medium">{label}</p>
        {description && (
          <p className="text-[11.5px] text-muted-foreground mt-0.5 leading-snug">{description}</p>
        )}
      </div>
      <div className="shrink-0">{control}</div>
    </div>
  )
}

export function SettingsCard({ children }: { children: React.ReactNode }) {
  return (
    <section className="rounded-lg border border-border/50 bg-card/30 divide-y divide-border/40 overflow-hidden">
      {children}
    </section>
  )
}
