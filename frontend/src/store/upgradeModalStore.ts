import { create } from 'zustand'

export type FreeTierLimitType =
  | 'request_limit_reached'
  | 'project_limit_reached'
  | 'account_limit_reached'
  | 'mock_requires_pro'
  | 'multi_account_requires_pro'
  | 'archive_requires_pro'
  | 'beta_requires_entitlement'

export type TargetPlan = 'pro' | 'ultimate'

interface UpgradeModalState {
  open: boolean
  limitType: FreeTierLimitType | null
  targetPlan: TargetPlan
  show: (limitType: FreeTierLimitType, targetPlan?: TargetPlan) => void
  close: () => void
}

export const useUpgradeModalStore = create<UpgradeModalState>((set) => ({
  open: false,
  limitType: null,
  targetPlan: 'pro',
  show: (limitType, targetPlan = 'pro') =>
    set({ open: true, limitType, targetPlan }),
  close: () => set({ open: false, limitType: null }),
}))
