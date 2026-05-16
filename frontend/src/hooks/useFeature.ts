import { useLicenseStore } from '@/store/licenseStore'

export function useFeature(key: string): boolean {
  return useLicenseStore((s) => !!s.license?.features?.[key])
}

export function usePlan(): string {
  return useLicenseStore((s) => s.plan())
}
