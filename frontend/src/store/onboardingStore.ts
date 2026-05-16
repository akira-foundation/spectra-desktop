import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

export type OnboardingStep = 'welcome' | 'auth' | 'project' | 'ready'

interface OnboardingState {
  completed: boolean
  authOnly: boolean
  currentStep: OnboardingStep
  setCurrentStep: (step: OnboardingStep) => void
  complete: () => void
  reset: () => void
  requireAuth: () => void
  clearRequireAuth: () => void
}

export const useOnboardingStore = create<OnboardingState>()(
  persist(
    (set) => ({
      completed: false,
      authOnly: false,
      currentStep: 'welcome',
      setCurrentStep: (step) => set({ currentStep: step }),
      complete: () => set({ completed: true, authOnly: false, currentStep: 'ready' }),
      reset: () => set({ completed: false, authOnly: false, currentStep: 'welcome' }),
      requireAuth: () => set({ authOnly: true, currentStep: 'auth' }),
      clearRequireAuth: () => set({ authOnly: false }),
    }),
    {
      name: 'spectra:onboarding',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ completed: state.completed }),
    },
  ),
)
