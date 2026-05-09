import { Layers } from 'lucide-react'
import { Card } from './Card'

interface Props {
  env: { name: string; vars?: Record<string, string> } | null
  count: number
  onOpen: () => void
}

export function EnvCard({ env, count }: Props) {
  return (
    <Card title="Environment" icon={Layers}>
      {env ? (
        <div className="space-y-1">
          <p className="text-[13px] font-medium truncate">{env.name}</p>
          <p className="text-[10.5px] text-muted-foreground">
            {Object.keys(env.vars ?? {}).length} variable
            {Object.keys(env.vars ?? {}).length === 1 ? '' : 's'}
          </p>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">
          {count > 0 ? 'No active environment selected.' : 'No environments yet.'}
        </p>
      )}
    </Card>
  )
}
