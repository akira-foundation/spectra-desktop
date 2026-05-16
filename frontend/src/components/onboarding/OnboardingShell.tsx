import { useEffect, useMemo, useState } from 'react'
import type { CSSProperties } from 'react'
import { ArrowRight } from 'lucide-react'
import { useOnboardingStore, type OnboardingStep } from '@/store/onboardingStore'
import { useLicenseStore } from '@/store/licenseStore'
import { StepDots } from './parts/StepDots'
import { WelcomeStep } from './steps/WelcomeStep'
import { SignInStep } from './steps/SignInStep'
import { ProjectStep } from './steps/ProjectStep'
import { ReadyStep } from './steps/ReadyStep'
import { AddProjectDialog } from '@/components/projects/AddProjectDialog'
import { BillingIsConfigured } from '../../../wailsjs/go/app/App'

const drag = { '--wails-draggable': 'drag' } as CSSProperties
const noDrag = { '--wails-draggable': 'no-drag' } as CSSProperties

export function OnboardingShell() {
  const currentStep = useOnboardingStore((s) => s.currentStep)
  const setCurrentStep = useOnboardingStore((s) => s.setCurrentStep)
  const complete = useOnboardingStore((s) => s.complete)
  const authOnly = useOnboardingStore((s) => s.authOnly)
  const clearRequireAuth = useOnboardingStore((s) => s.clearRequireAuth)
  const initLicense = useLicenseStore((s) => s.init)
  const [billingConfigured, setBillingConfigured] = useState<boolean | null>(null)

  useEffect(() => {
    void initLicense()
    void BillingIsConfigured().then(setBillingConfigured).catch(() => setBillingConfigured(false))
  }, [initLicense])

  const steps: OnboardingStep[] = useMemo(() => {
    if (billingConfigured === false) return ['welcome', 'project', 'ready']
    return ['welcome', 'auth', 'project', 'ready']
  }, [billingConfigured])

  const stepIndex = useMemo(() => steps.indexOf(currentStep), [currentStep, steps])

  const goNext = () => {
    const next = steps[stepIndex + 1]
    if (next) setCurrentStep(next)
  }

  const onAuthenticated = () => {
    if (authOnly) {
      clearRequireAuth()
      complete()
      return
    }
    setCurrentStep('project')
  }

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col bg-sidebar text-foreground select-none"
      style={drag}
    >
      {/* Drag region — top 40px reserved for traffic lights + window dragging */}
      <div className="h-10 shrink-0" />

      <main
        className="flex-1 flex items-center justify-center px-6 overflow-auto"
        style={noDrag}
      >
        <div className="w-full max-w-md py-8">
          {currentStep === 'welcome' && <WelcomeStep />}
          {currentStep === 'auth' && (
            <SignInStep onAuthenticated={onAuthenticated} />
          )}
          {currentStep === 'project' && (
            <ProjectStep
              onAdded={() => setCurrentStep('ready')}
              onSkip={() => setCurrentStep('ready')}
            />
          )}
          {currentStep === 'ready' && <ReadyStep onEnter={complete} />}
        </div>
      </main>

      <footer
        className="h-20 shrink-0 px-8 flex items-center justify-between"
        style={noDrag}
      >
        <div className="w-32">
          <StepDots count={steps.length} active={stepIndex} />
        </div>

        <div className="flex items-center gap-3 w-32 justify-end">
          {currentStep === 'welcome' && (
            <button
              type="button"
              onClick={goNext}
              className="inline-flex items-center gap-2 h-10 px-5 rounded-full bg-foreground text-background hover:bg-foreground/90 transition-colors text-[13.5px] font-medium"
            >
              Continue
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </footer>

      <AddProjectDialog />
    </div>
  )
}
