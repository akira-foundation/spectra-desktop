import { Zap, Play, ArrowRight } from 'lucide-react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface GetStartedStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function GetStartedStep({ onNext, onBack }: GetStartedStepProps) {
  const actions = [
    {
      title: 'Create Your First Request',
      description: 'Build and test your first API request in seconds',
      icon: <Zap className="w-8 h-8" />
    },
    {
      title: 'Explore Auto-Discovered Endpoints',
      description: 'See all your Laravel routes automatically loaded',
      icon: <Play className="w-8 h-8" />
    }
  ]

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Ready to Get Started?
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            You're all set! Here's what you can do next
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4 w-full max-w-2xl">
          {actions.map((action, idx) => (
            <div
              key={idx}
              className="rounded-xl border border-slate-200 dark:border-white/10 p-6 text-left hover:border-slate-300 dark:hover:border-white/20 hover:shadow-md transition-all duration-300"
              style={{
                animation: `slide-up 0.6s ease-out ${idx * 0.1}s both`
              }}
            >
              <div className="p-3 rounded-lg bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-400 w-fit mb-4">
                {action.icon}
              </div>
              <h3 className="font-semibold text-slate-900 dark:text-white mb-2">
                {action.title}
              </h3>
              <p className="text-sm text-slate-600 dark:text-slate-400">
                {action.description}
              </p>
            </div>
          ))}
        </div>

        <div className="text-center space-y-2 pt-4">
          <p className="text-sm text-slate-600 dark:text-slate-400">
            You can always access the settings to modify your configuration
          </p>
        </div>

        <OnboardingButtonGroup>
          {onBack && (
            <OnboardingButton variant="secondary" onClick={onBack}>
              Back
            </OnboardingButton>
          )}
          {onNext && (
            <OnboardingButton onClick={onNext} className="group">
              <span className="flex items-center gap-2">
                Let's Go!
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </span>
            </OnboardingButton>
          )}
        </OnboardingButtonGroup>
      </div>

      <style>{`
        @keyframes slide-up {
          from {
            opacity: 0;
            transform: translateY(20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </OnboardingContainer>
  )
}
