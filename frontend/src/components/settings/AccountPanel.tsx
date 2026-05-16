import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import {
  UserCircle2,
  LogOut,
  RefreshCw,
  ExternalLink,
  CheckCircle2,
  XCircle,
  ShieldCheck,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useLicenseStore } from '@/store/licenseStore'
import { useOnboardingStore } from '@/store/onboardingStore'
import { SettingsHeader } from './SettingsHeader'
import { SettingsCard, SettingsRow } from './SettingsRow'
import { LicenseActivationDialog } from '@/components/license/LicenseActivationDialog'
import { BrowserOpenURL } from '../../../wailsjs/runtime'

export function AccountPanel() {
  const license = useLicenseStore((s) => s.license)
  const load = useLicenseStore((s) => s.load)
  const verify = useLicenseStore((s) => s.verify)
  const refresh = useLicenseStore((s) => s.refresh)
  const clear = useLicenseStore((s) => s.clear)
  const portal = useLicenseStore((s) => s.portal)
  const requireAuth = useOnboardingStore((s) => s.requireAuth)

  const [activationOpen, setActivationOpen] = useState(false)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    void load()
  }, [load])

  const signOut = async () => {
    if (!confirm('Sign out and clear the local license?')) return
    setBusy(true)
    try {
      await clear()
      requireAuth()
      toast.success('Signed out')
    } finally {
      setBusy(false)
    }
  }

  const refreshLicense = async () => {
    setBusy(true)
    try {
      await refresh()
      toast.success('License refreshed')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const verifyLicense = async () => {
    setBusy(true)
    try {
      await verify()
      toast.success('License verified')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
  }

  const openPortal = async () => {
    try {
      const url = await portal()
      if (url) BrowserOpenURL(url)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    }
  }

  const signedIn = !!license?.customerEmail
  const featuresCount = Object.keys(license?.features ?? {}).length
  const activeFeatures = Object.entries(license?.features ?? {}).filter(([, v]) => v).length

  return (
    <div>
      <SettingsHeader
        icon={UserCircle2}
        title="Account"
        description="License, plan, and identity bound to this device."
      />

      <div className="space-y-4">
        <SettingsCard>
          <SettingsRow
            label="Signed in as"
            control={
              <span className="text-[12px] font-mono text-muted-foreground">
                {signedIn ? license!.customerEmail : 'Not signed in'}
              </span>
            }
          />
          <SettingsRow
            label="Plan"
            description={license?.cycle ? `Billed ${license.cycle}` : undefined}
            control={<PlanBadge plan={license?.plan || 'free'} />}
          />
          <SettingsRow
            label="Status"
            control={<StatusBadge status={license?.status || 'inactive'} grace={!!license?.gracePeriod} />}
          />
          {license?.validUntil && (
            <SettingsRow
              label="Valid until"
              control={
                <span className="text-[12px] font-mono text-muted-foreground">
                  {formatDate(license.validUntil)}
                </span>
              }
            />
          )}
          {license?.lastVerifiedAt && (
            <SettingsRow
              label="Last verified"
              control={
                <span className="text-[12px] font-mono text-muted-foreground">
                  {formatDate(license.lastVerifiedAt)}
                </span>
              }
            />
          )}
          {license?.deviceId && (
            <SettingsRow
              label="Device"
              control={
                <span className="text-[11.5px] font-mono text-muted-foreground truncate max-w-[260px]">
                  {license.deviceId}
                </span>
              }
            />
          )}
        </SettingsCard>

        {featuresCount > 0 && (
          <SettingsCard>
            <div className="px-4 py-3 space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-[13px] font-medium">Features</p>
                <span className="text-[11px] text-muted-foreground tabular-nums">
                  {activeFeatures}/{featuresCount} enabled
                </span>
              </div>
              <ul className="grid grid-cols-2 gap-1.5">
                {Object.entries(license!.features).map(([key, enabled]) => (
                  <li
                    key={key}
                    className="flex items-center gap-1.5 text-[11.5px] text-muted-foreground"
                  >
                    {enabled ? (
                      <CheckCircle2 className="h-3 w-3 text-emerald-500" />
                    ) : (
                      <XCircle className="h-3 w-3 text-muted-foreground/40" />
                    )}
                    <span className={enabled ? 'text-foreground/80' : 'line-through'}>{key}</span>
                  </li>
                ))}
              </ul>
            </div>
          </SettingsCard>
        )}

        <SettingsCard>
          <SettingsRow
            label="Activate device"
            description="Bind this machine to the active plan and issue a signed license snapshot."
            control={
              <Button size="sm" onClick={() => setActivationOpen(true)} disabled={busy}>
                <ShieldCheck className="h-3.5 w-3.5" />
                Activate
              </Button>
            }
          />
          <SettingsRow
            label="Refresh license"
            description="Pull the latest signed snapshot from the billing service."
            control={
              <Button size="sm" variant="outline" onClick={() => void refreshLicense()} disabled={busy}>
                <RefreshCw className="h-3.5 w-3.5" />
                Refresh
              </Button>
            }
          />
          <SettingsRow
            label="Re-verify locally"
            description="Re-check the signed snapshot against the embedded public key."
            control={
              <Button size="sm" variant="ghost" onClick={() => void verifyLicense()} disabled={busy}>
                Verify
              </Button>
            }
          />
          <SettingsRow
            label="Manage subscription"
            description="Opens the customer portal in your browser."
            control={
              <Button size="sm" variant="outline" onClick={() => void openPortal()}>
                <ExternalLink className="h-3.5 w-3.5" />
                Open portal
              </Button>
            }
          />
          <SettingsRow
            label="Sign out"
            description="Clears the local session and license snapshot."
            control={
              <Button
                size="sm"
                variant="ghost"
                onClick={() => void signOut()}
                disabled={busy}
                className="text-destructive hover:text-destructive"
              >
                <LogOut className="h-3.5 w-3.5" />
                Sign out
              </Button>
            }
          />
        </SettingsCard>
      </div>

      <LicenseActivationDialog open={activationOpen} onClose={() => setActivationOpen(false)} />
    </div>
  )
}

function PlanBadge({ plan }: { plan: string }) {
  const label = capitalize(plan)
  const isPaid = plan && plan !== 'free' && plan !== 'inactive'
  return (
    <span
      className={
        isPaid
          ? 'inline-flex items-center rounded border border-primary/40 bg-primary/10 px-1.5 py-px text-[10.5px] font-medium uppercase tracking-wider text-primary'
          : 'inline-flex items-center rounded border border-border/50 bg-muted px-1.5 py-px text-[10.5px] font-medium uppercase tracking-wider text-muted-foreground'
      }
    >
      {label}
    </span>
  )
}

function StatusBadge({ status, grace }: { status: string; grace: boolean }) {
  const tone =
    status === 'active' && !grace
      ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
      : status === 'active' && grace
        ? 'border-amber-500/40 bg-amber-500/10 text-amber-600 dark:text-amber-400'
        : status === 'expired'
          ? 'border-destructive/40 bg-destructive/10 text-destructive'
          : 'border-border/50 bg-muted text-muted-foreground'
  const label = grace && status === 'active' ? 'Grace period' : capitalize(status)
  return (
    <span className={`inline-flex items-center rounded border px-1.5 py-px text-[10.5px] font-medium uppercase tracking-wider ${tone}`}>
      {label}
    </span>
  )
}

function capitalize(s: string): string {
  return s ? s[0].toUpperCase() + s.slice(1) : s
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
