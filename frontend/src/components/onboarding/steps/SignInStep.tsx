import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Mail, ArrowLeft, RefreshCw } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { useLicenseStore } from '@/store/licenseStore'
import { OtpInput } from '../parts/OtpInput'
import { OauthProviderButton } from '../parts/OauthProviderButton'
import { BillingOauthProviders } from '../../../../wailsjs/go/app/App'

interface Props {
  onAuthenticated: () => void
}

type Mode = 'choose' | 'email' | 'code'

export function SignInStep({ onAuthenticated }: Props) {
  const requestOtp = useLicenseStore((s) => s.requestOtp)
  const verifyOtp = useLicenseStore((s) => s.verifyOtp)
  const oauthLogin = useLicenseStore((s) => s.oauthLogin)
  const cancelOauth = useLicenseStore((s) => s.cancelOauth)

  const [mode, setMode] = useState<Mode>('choose')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [providers, setProviders] = useState<string[] | null>(null)
  const [providersError, setProvidersError] = useState<string | null>(null)
  const [resendIn, setResendIn] = useState(0)

  useEffect(() => {
    void BillingOauthProviders()
      .then((list) => {
        setProviders(Array.isArray(list) ? list : [])
        setProvidersError(null)
      })
      .catch((err) => {
        setProviders([])
        setProvidersError(err instanceof Error ? err.message : String(err))
      })
  }, [])

  useEffect(() => {
    if (resendIn <= 0) return
    const t = setInterval(() => setResendIn((v) => Math.max(0, v - 1)), 1000)
    return () => clearInterval(t)
  }, [resendIn])

  const sendCode = async () => {
    setError(null)
    if (!email.includes('@')) {
      setError('Enter a valid email address')
      return
    }
    setBusy(true)
    try {
      await requestOtp(email)
      setMode('code')
      setResendIn(30)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const verify = async (full: string) => {
    setError(null)
    setBusy(true)
    try {
      await verifyOtp(email, full)
      onAuthenticated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid code')
      setBusy(false)
    }
  }

  const startOauth = async (provider: string) => {
    setError(null)
    setBusy(true)
    try {
      await oauthLogin(provider)
      onAuthenticated()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      void cancelOauth()
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="w-full max-w-sm mx-auto">
      <header className="text-center mb-8">
        <h1 className="text-[28px] font-semibold tracking-tight leading-tight">Sign in</h1>
        <p className="mt-2 text-[13.5px] text-muted-foreground">
          Sync projects, accounts, and licences across machines.
        </p>
      </header>

      {mode === 'choose' && (
        <div className="space-y-2.5">
          {providers === null ? (
            <div className="h-11 rounded-xl bg-card/40 animate-pulse" />
          ) : providers.length === 0 ? (
            <div className="rounded-xl border border-amber-500/40 bg-amber-500/5 px-3.5 py-2.5 text-[11.5px] text-amber-600 dark:text-amber-400 leading-snug">
              No OAuth providers reachable.
              {providersError && (
                <span className="block mt-1 font-mono text-[10.5px] opacity-80 truncate">
                  {providersError}
                </span>
              )}
            </div>
          ) : (
            providers.map((p) => (
              <OauthProviderButton
                key={p}
                provider={p}
                label={capitalize(p)}
                disabled={busy}
                onClick={() => void startOauth(p)}
              />
            ))
          )}
          <button
            type="button"
            onClick={() => setMode('email')}
            disabled={busy}
            className="w-full h-11 inline-flex items-center justify-center gap-2.5 rounded-xl border border-border/60 bg-card/40 hover:bg-accent/40 transition-colors text-[13.5px] font-medium"
          >
            <Mail className="h-4 w-4" />
            Continue with email
          </button>
        </div>
      )}

      {mode === 'email' && (
        <div className="space-y-4">
          <button
            type="button"
            onClick={() => setMode('choose')}
            className="inline-flex items-center gap-1 text-[12px] text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-3 w-3" />
            Back
          </button>
          <Input
            type="email"
            autoFocus
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void sendCode()
            }}
            placeholder="you@example.com"
            className="h-11 text-[14px] rounded-xl"
          />
          <button
            type="button"
            onClick={() => void sendCode()}
            disabled={busy}
            className="w-full h-11 rounded-xl bg-foreground text-background hover:bg-foreground/90 disabled:opacity-50 transition-colors text-[13.5px] font-medium"
          >
            {busy ? 'Sending…' : 'Continue with email'}
          </button>
        </div>
      )}

      {mode === 'code' && (
        <div className="space-y-4">
          <button
            type="button"
            onClick={() => setMode('email')}
            className="inline-flex items-center gap-1 text-[12px] text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-3 w-3" />
            Back
          </button>
          <p className="text-[12.5px] text-muted-foreground text-center">
            We sent a code to{' '}
            <span className="font-medium text-foreground/90">{email}</span>
          </p>
          <div className="flex justify-center">
            <OtpInput
              value={code}
              onChange={setCode}
              onComplete={(v) => void verify(v)}
              disabled={busy}
              autoFocus
            />
          </div>
          <div className="text-center">
            <button
              type="button"
              onClick={() => {
                if (resendIn === 0) {
                  void sendCode()
                  toast.success('Code resent')
                }
              }}
              disabled={resendIn > 0 || busy}
              className="inline-flex items-center gap-1.5 text-[11.5px] text-muted-foreground hover:text-foreground disabled:opacity-50"
            >
              <RefreshCw className="h-3 w-3" />
              {resendIn > 0 ? `Resend in ${resendIn}s` : 'Resend code'}
            </button>
          </div>
        </div>
      )}

      {error && (
        <p className="mt-4 text-[12px] text-destructive text-center">{error}</p>
      )}
    </div>
  )
}

function capitalize(s: string): string {
  return s ? s[0].toUpperCase() + s.slice(1) : s
}
