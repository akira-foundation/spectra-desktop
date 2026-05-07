import { Plug, Github, Code2, Check, ArrowRight } from 'lucide-react'
import { useState } from 'react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface IntegrationStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function IntegrationStep({ onNext, onBack }: IntegrationStepProps) {
  const [selectedIntegrations, setSelectedIntegrations] = useState<string[]>([])

  const integrations = [
    {
      id: 'github',
      name: 'GitHub',
      description: 'Sync with GitHub repositories',
      icon: <Github className="w-6 h-6" />,
      color: 'from-gray-600 to-gray-900'
    },
    {
      id: 'webhooks',
      name: 'Webhooks',
      description: 'Send requests to external services',
      icon: <Plug className="w-6 h-6" />,
      color: 'from-blue-500 to-cyan-500'
    },
    {
      id: 'rest',
      name: 'REST APIs',
      description: 'Test third-party REST endpoints',
      icon: <Code2 className="w-6 h-6" />,
      color: 'from-purple-500 to-pink-500'
    }
  ]

  const toggleIntegration = (id: string) => {
    setSelectedIntegrations(prev =>
      prev.includes(id)
        ? prev.filter(i => i !== id)
        : [...prev, id]
    )
  }

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Integrations & Connectivity
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            Connect with external services and APIs to enhance your workflow
          </p>
        </div>

        <div className="grid grid-cols-3 gap-4 w-full max-w-3xl">
          {integrations.map((integration, idx) => {
            const isSelected = selectedIntegrations.includes(integration.id)

            return (
              <button
                key={integration.id}
                onClick={() => toggleIntegration(integration.id)}
                className={`
                  relative rounded-2xl border p-6 transition-all duration-300 text-left
                  ${
                    isSelected
                      ? 'border-violet-500 shadow-lg shadow-violet-200 dark:shadow-violet-900/20'
                      : 'border-slate-200 dark:border-white/10 hover:border-slate-300 dark:hover:border-white/20 hover:shadow-md'
                  }
                `}
                style={{
                  animation: `slide-up 0.6s ease-out ${idx * 0.1}s both`
                }}
              >
                <div className={`flex items-start justify-between gap-3 mb-3`}>
                  <div className={`p-3 rounded-lg bg-gradient-to-br ${integration.color}`}>
                    <div className="text-white">
                      {integration.icon}
                    </div>
                  </div>

                  {isSelected && (
                    <div className="flex h-6 w-6 items-center justify-center rounded-full bg-violet-600">
                      <Check className="h-4 w-4 text-white" />
                    </div>
                  )}
                </div>

                <h3 className="font-semibold text-slate-900 dark:text-white mb-1">
                  {integration.name}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">
                  {integration.description}
                </p>
              </button>
            )
          })}
        </div>

        <div className="rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/20 p-4 w-full max-w-3xl">
          <p className="text-sm text-amber-900 dark:text-amber-200">
            ℹ️ <strong>Note:</strong> You can enable or disable integrations anytime in settings
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
                Continue
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
