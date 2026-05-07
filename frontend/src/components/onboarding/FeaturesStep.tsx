import { Terminal, History, Zap, Globe, ArrowRight } from 'lucide-react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface FeaturesStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function FeaturesStep({ onNext, onBack }: FeaturesStepProps) {
  const features = [
    {
      icon: Terminal,
      title: 'Smart Request Builder',
      description: 'Intuitive interface with auto-complete, syntax highlighting, and request history at your fingertips',
      color: 'from-violet-500 to-purple-500'
    },
    {
      icon: History,
      title: 'Real-Time Sync',
      description: 'Automatic synchronization with your Laravel application endpoints in real-time',
      color: 'from-blue-500 to-cyan-500'
    },
    {
      icon: Zap,
      title: 'Lightning Fast',
      description: 'Native desktop performance with minimal resource usage and instant responses',
      color: 'from-amber-500 to-orange-500'
    },
    {
      icon: Globe,
      title: 'Environment Management',
      description: 'Effortlessly switch between local development, staging, and production environments',
      color: 'from-emerald-500 to-green-500'
    }
  ]

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Built for Developers
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            Everything you need to test, debug, and document your Laravel APIs efficiently
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4 w-full max-w-3xl">
          {features.map((feature, index) => {
            const Icon = feature.icon

            return (
              <div
                key={index}
                className="rounded-2xl border border-slate-200 dark:border-white/10 p-6 text-left hover:border-slate-300 dark:hover:border-white/20 hover:shadow-md transition-all duration-300 animate-in fade-in slide-in-from-bottom-4"
                style={{
                  animationDelay: `${index * 100}ms`,
                  animationFillMode: 'both'
                }}
              >
                <div className={`mb-4 inline-flex rounded-xl bg-gradient-to-br ${feature.color} p-3`}>
                  <Icon className="w-6 h-6 text-white" />
                </div>
                <h3 className="font-semibold text-slate-900 dark:text-white mb-2">
                  {feature.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
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
                Continue
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </span>
            </OnboardingButton>
          )}
        </OnboardingButtonGroup>
      </div>
    </OnboardingContainer>
  )
}
