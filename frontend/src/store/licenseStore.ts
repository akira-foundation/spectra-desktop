import { create } from 'zustand'
import {
  BillingGetLicense,
  BillingVerifyLicense,
  BillingActivateLicense,
  BillingRefreshLicense,
  BillingLogout,
  BillingIsAuthenticated,
  BillingRequestOTP,
  BillingVerifyOTP,
  BillingOauthLogin,
  BillingCancelOauth,
  BillingPlans,
  BillingPortal,
} from '../../wailsjs/go/app/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'
import type { app } from '../../wailsjs/go/models'

export type Plan = 'free' | 'pro' | 'ultimate' | string
export type LicenseStatus = 'inactive' | 'active' | 'expired' | 'invalid'

export interface LicenseDTO {
  customerId: string
  customerEmail: string
  customerName: string
  plan: string
  cycle: string
  status: string
  validUntil: string
  activatedAt: string
  lastVerifiedAt: string
  features: Record<string, boolean>
  deviceId: string
  cancelAtPeriodEnd: boolean
  cancelAt?: string
  targetPlan?: string
  gracePeriod: boolean
}

export interface OauthEntitlement {
  planKey?: string
  source: string
  endsAt?: string
}

export interface OauthLoginResult {
  customerId: string
  customerEmail: string
  customerName: string
  entitlement?: OauthEntitlement
  requiresPlanSelection: boolean
}

interface LicenseState {
  license: LicenseDTO | null
  authenticated: boolean
  loading: boolean
  activating: boolean
  initialized: boolean
  unsubscribe: (() => void) | null

  init: () => Promise<void>
  load: () => Promise<void>
  verify: () => Promise<void>
  activate: (deviceName?: string) => Promise<void>
  refresh: () => Promise<void>
  clear: () => Promise<void>
  plan: () => Plan
  hasFeature: (key: string) => boolean

  requestOtp: (email: string) => Promise<void>
  verifyOtp: (email: string, code: string) => Promise<LicenseDTO | null>
  oauthLogin: (provider: string) => Promise<OauthLoginResult>
  cancelOauth: () => Promise<void>

  plans: () => Promise<Record<string, unknown>>
  portal: (returnURL?: string) => Promise<string>
}

export const useLicenseStore = create<LicenseState>((set, get) => ({
  license: null,
  authenticated: false,
  loading: false,
  activating: false,
  initialized: false,
  unsubscribe: null,

  init: async () => {
    if (get().initialized) return
    set({ initialized: true })

    const onLicense = (dto: LicenseDTO | null) => set({ license: dto })
    const onSession = () => void get().load()
    EventsOn('billing:license-changed', onLicense)
    EventsOn('billing:session-changed', onSession)
    set({
      unsubscribe: () => {
        EventsOff('billing:license-changed')
        EventsOff('billing:session-changed')
      },
    })

    await get().load()
  },

  load: async () => {
    set({ loading: true })
    try {
      const [dto, authed] = await Promise.all([
        BillingGetLicense() as Promise<LicenseDTO | null>,
        BillingIsAuthenticated() as Promise<boolean>,
      ])
      set({ license: dto, authenticated: !!authed })
    } finally {
      set({ loading: false })
    }
  },

  verify: async () => {
    set({ loading: true })
    try {
      const dto = (await BillingVerifyLicense()) as LicenseDTO | null
      set({ license: dto })
    } finally {
      set({ loading: false })
    }
  },

  activate: async (deviceName) => {
    set({ activating: true })
    try {
      const dto = (await BillingActivateLicense({
        deviceName: deviceName ?? '',
      } as unknown as app.BillingActivationInput)) as LicenseDTO
      set({ license: dto })
    } finally {
      set({ activating: false })
    }
  },

  refresh: async () => {
    const dto = (await BillingRefreshLicense()) as LicenseDTO
    set({ license: dto })
  },

  clear: async () => {
    await BillingLogout()
    set({ license: null, authenticated: false })
  },

  plan: (): Plan => {
    const license = get().license
    if (!license || license.status !== 'active') return 'free'
    const p = (license.plan || 'free').toLowerCase()
    return p as Plan
  },

  hasFeature: (key) => {
    const license = get().license
    if (!license) return false
    return !!license.features?.[key]
  },

  requestOtp: async (email) => {
    await BillingRequestOTP({ email } as unknown as app.BillingOtpRequestInput)
  },

  verifyOtp: async (email, code) => {
    const dto = (await BillingVerifyOTP({
      email,
      code,
    } as unknown as app.BillingOtpVerifyInput)) as LicenseDTO | null
    if (dto) {
      set({ license: dto, authenticated: true })
    }
    return dto
  },

  oauthLogin: async (provider) => {
    const result = (await BillingOauthLogin(provider)) as unknown as OauthLoginResult
    set({ authenticated: true })
    await get().load()
    return result
  },

  cancelOauth: async () => {
    await BillingCancelOauth()
  },

  plans: async () => {
    return (await BillingPlans()) as Record<string, unknown>
  },

  portal: async (returnURL) => {
    return (await BillingPortal(returnURL ?? '')) as string
  },
}))
