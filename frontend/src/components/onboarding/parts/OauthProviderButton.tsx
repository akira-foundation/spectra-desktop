import { ProviderIcon } from './ProviderIcon'
import { cn } from '@/lib/utils'

interface Props {
  provider: string
  label: string
  disabled?: boolean
  onClick: () => void
}

export function OauthProviderButton({ provider, label, disabled, onClick }: Props) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'w-full h-11 px-3 inline-flex items-center justify-center gap-2.5 rounded-xl border border-border/60 bg-card/40 transition-colors',
        'hover:bg-accent/40 hover:border-border',
        disabled && 'opacity-50 cursor-not-allowed',
      )}
    >
      <ProviderIcon provider={provider} />
      <span className="text-[13.5px] font-medium">Continue with {label}</span>
    </button>
  )
}
