import { useEffect, useState } from 'react'
import { useLicenseStore } from '@/store/licenseStore'

interface Props {
  onEnter: () => void
}

export function ReadyStep({ onEnter }: Props) {
  const license = useLicenseStore((s) => s.license)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const t = setTimeout(() => setVisible(true), 50)
    return () => clearTimeout(t)
  }, [])

  const greeting = license?.customerName || license?.customerEmail?.split('@')[0] || null
  const planLabel = license?.plan ? capitalize(license.plan) : null

  return (
    <div
      className={`flex flex-col items-center text-center transition-all duration-700 ease-out ${
        visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'
      }`}
    >
      <div className="relative h-20 w-20 mb-8">
        <div className="absolute inset-0 rounded-full bg-emerald-500/20 animate-pulse" />
        <div className="relative h-20 w-20 rounded-full bg-emerald-500/10 border border-emerald-500/40 flex items-center justify-center">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            className="h-9 w-9 text-emerald-500"
            strokeWidth={1.5}
            strokeLinecap="round"
            strokeLinejoin="round"
            stroke="currentColor"
          >
            <path d="M5 12l5 5L20 7" />
          </svg>
        </div>
      </div>

      <h1 className="text-[36px] font-semibold tracking-tight leading-none">
        {greeting ? `Hello, ${greeting}` : 'All set'}
      </h1>

      <p className="mt-4 text-[14px] text-muted-foreground max-w-xs leading-relaxed">
        {license?.customerEmail
          ? `Signed in${planLabel ? ` · ${planLabel} plan` : ''}.`
          : 'Running in local-only mode. Sign in any time from Settings.'}
      </p>

      <button
        type="button"
        onClick={onEnter}
        className="mt-10 h-11 px-8 rounded-full bg-foreground text-background hover:bg-foreground/90 transition-colors text-[13.5px] font-medium"
      >
        Start using Spectra
      </button>
    </div>
  )
}

function capitalize(s: string): string {
  return s ? s[0].toUpperCase() + s.slice(1) : s
}
