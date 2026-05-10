import type { LucideIcon } from 'lucide-react'

interface Props {
  icon: LucideIcon
  title: string
  description: string
}

export function SettingsHeader({ icon: Icon, title, description }: Props) {
  return (
    <div className="flex items-center gap-3.5 mb-6">
      <div className="h-12 w-12 rounded-xl flex items-center justify-center bg-gradient-to-br from-primary/20 to-primary/5 border border-primary/20 shrink-0">
        <Icon className="h-5 w-5 text-primary" strokeWidth={1.75} />
      </div>
      <div className="min-w-0">
        <h1 className="text-[20px] font-semibold tracking-tight leading-tight">{title}</h1>
        <p className="text-[12.5px] text-muted-foreground mt-0.5">{description}</p>
      </div>
    </div>
  )
}
