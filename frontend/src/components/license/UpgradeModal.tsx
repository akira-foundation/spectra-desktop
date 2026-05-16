import toast from 'react-hot-toast'
import { Sparkles, ExternalLink } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  useUpgradeModalStore,
  type FreeTierLimitType,
  type TargetPlan,
} from '@/store/upgradeModalStore'
import { useLicenseStore } from '@/store/licenseStore'
import { BrowserOpenURL } from '../../../wailsjs/runtime'

interface Copy {
  title: string
  description: string
}

const COPY: Record<FreeTierLimitType, Copy> = {
  request_limit_reached: {
    title: 'Daily request limit reached',
    description: 'Free plan tracks 200 inspector requests per day. Upgrade for unlimited.',
  },
  project_limit_reached: {
    title: 'Project limit reached',
    description: 'Free plan caps active projects. Pro lifts the cap and adds remote sync.',
  },
  account_limit_reached: {
    title: 'Account slot limit reached',
    description: 'Switch between identities without limits on Pro.',
  },
  mock_requires_pro: {
    title: 'Mock server is a Pro feature',
    description: 'Replay history offline and generate fake responses with the embedded server.',
  },
  multi_account_requires_pro: {
    title: 'Multi-account requires Pro',
    description: 'Save credentials per identity and switch in 1 click.',
  },
  archive_requires_pro: {
    title: 'Project archives require Pro',
    description: 'Export and import .spectra bundles with encryption.',
  },
  beta_requires_entitlement: {
    title: 'Beta channel locked',
    description: 'Beta builds are reserved for entitled accounts.',
  },
}

export function UpgradeModal() {
  const open = useUpgradeModalStore((s) => s.open)
  const limitType = useUpgradeModalStore((s) => s.limitType)
  const targetPlan = useUpgradeModalStore((s) => s.targetPlan)
  const close = useUpgradeModalStore((s) => s.close)
  const portal = useLicenseStore((s) => s.portal)
  const license = useLicenseStore((s) => s.license)

  const copy = limitType ? COPY[limitType] : null

  const openPortal = async () => {
    try {
      const url = await portal()
      if (url) {
        BrowserOpenURL(url)
        close()
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && close()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-primary" />
            {copy?.title ?? 'Upgrade required'}
          </DialogTitle>
          <DialogDescription>{copy?.description ?? ''}</DialogDescription>
        </DialogHeader>

        <PlanCompareCard targetPlan={targetPlan} currentPlan={license?.plan || 'free'} />

        <DialogFooter>
          <Button variant="ghost" onClick={close}>
            Later
          </Button>
          <Button onClick={() => void openPortal()}>
            <ExternalLink className="h-3.5 w-3.5" />
            Upgrade to {capitalize(targetPlan)}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function PlanCompareCard({
  targetPlan,
  currentPlan,
}: {
  targetPlan: TargetPlan
  currentPlan: string
}) {
  return (
    <div className="rounded-md border border-border/40 bg-card/30 px-3.5 py-3 text-[12px] space-y-1.5">
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground">Current</span>
        <span className="font-medium">{capitalize(currentPlan)}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground">Upgrading to</span>
        <span className="font-semibold text-primary">{capitalize(targetPlan)}</span>
      </div>
    </div>
  )
}

function capitalize(s: string): string {
  return s ? s[0].toUpperCase() + s.slice(1) : s
}
