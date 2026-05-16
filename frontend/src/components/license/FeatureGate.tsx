import { ReactNode } from 'react'
import { useFeature } from '@/hooks/useFeature'
import { useUpgradeModalStore, type FreeTierLimitType, type TargetPlan } from '@/store/upgradeModalStore'

interface Props {
  feature: string
  fallback?: ReactNode
  limitType?: FreeTierLimitType
  targetPlan?: TargetPlan
  children: ReactNode
}

export function FeatureGate({ feature, fallback, limitType, targetPlan, children }: Props) {
  const allowed = useFeature(feature)
  if (allowed) return <>{children}</>
  if (fallback !== undefined) return <>{fallback}</>
  return <LockedShim feature={feature} limitType={limitType} targetPlan={targetPlan} />
}

function LockedShim({
  feature,
  limitType,
  targetPlan,
}: {
  feature: string
  limitType?: FreeTierLimitType
  targetPlan?: TargetPlan
}) {
  const show = useUpgradeModalStore((s) => s.show)
  return (
    <button
      type="button"
      onClick={() => show(limitType ?? 'mock_requires_pro', targetPlan ?? 'pro')}
      className="inline-flex items-center text-[11px] text-muted-foreground/70 italic"
      title={`Feature "${feature}" requires upgrade`}
    >
      Pro feature — click to upgrade
    </button>
  )
}
