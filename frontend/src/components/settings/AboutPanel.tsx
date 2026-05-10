import { Info } from 'lucide-react'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'

export function AboutPanel() {
  return (
    <div>
      <SettingsHeader icon={Info} title="About" description="What is running locally." />

      <SettingsCard>
        <SettingsRow label="Version" control={<Mono value="0.1.0" />} />
        <SettingsRow label="Build" control={<Mono value="local" />} />
        <SettingsRow label="License" control={<Mono value="Beta · all features unlocked" />} />
      </SettingsCard>
    </div>
  )
}

function Mono({ value }: { value: string }) {
  return <span className="font-mono text-[12px] text-muted-foreground">{value}</span>
}
