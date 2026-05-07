import { useState } from 'react'
import { ThemeStep } from './ThemeStep'
import { WelcomeStep } from './WelcomeStep'
import { FeaturesStep } from './FeaturesStep'
import { SetupStep } from './SetupStep'
import { LicenseStep } from './LicenseStep'
import { KeyboardShortcutsStep } from './KeyboardShortcutsStep'
import { IntegrationStep } from './IntegrationStep'
import { GetStartedStep } from './GetStartedStep'
import { CelebrateStep } from './CelebrateStep'

interface OnboardingFlowProps {
  onComplete: () => void
}

export function OnboardingFlow({ onComplete }: OnboardingFlowProps)  {
  const [currentStep, setCurrentStep] = useState(0)

  const steps = [
    { component: WelcomeStep, name: 'Welcome' },
    { component: ThemeStep, name: 'Appearance' },
    { component: FeaturesStep, name: 'Features' },
    { component: KeyboardShortcutsStep, name: 'Shortcuts' },
    { component: SetupStep, name: 'Setup' },
    { component: LicenseStep, name: 'License' },
    { component: CelebrateStep, name: 'Ready' },

  ]

  const CurrentStepComponent = steps[currentStep].component

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    } else {
      onComplete()
    }
  }

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  return (
    <div className="h-screen w-screen flex flex-col bg-background">
      {/* Step Content */}
      <div className="flex-1 overflow-auto">
        <CurrentStepComponent onNext={handleNext} onBack={handleBack} onComplete={onComplete} />
      </div>

      {/* Footer with Step Indicators */}
      <div className="py-6 px-8 border-t">
        <div className="max-w-2xl mx-auto">
          <div className="flex justify-center gap-2">
            {steps.map((_, index) => (
              <div
                key={index}
                className={`h-1.5 rounded-full transition-all ${
                  index === currentStep
                    ? 'w-8 bg-primary'
                    : index < currentStep
                    ? 'w-1.5 bg-primary/50'
                    : 'w-1.5 bg-muted'
                }`}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
