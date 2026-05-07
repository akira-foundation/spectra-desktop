import { Zap, Sparkles, Code2, Rocket, ArrowRight } from 'lucide-react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButton } from './OnboardingButton'

interface WelcomeStepProps {
  onNext?: () => void;
}

export function WelcomeStep({ onNext }: WelcomeStepProps) {
  return (
    <OnboardingContainer variant="centered">
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        {/* Animated Logo */}
        <div className="relative mb-4">
          <div className="absolute inset-0 rounded-2xl bg-gradient-to-r from-violet-500 to-purple-600 opacity-30 blur-2xl animate-pulse" />
          <div className="relative w-20 h-20 rounded-2xl bg-gradient-to-br from-violet-500 via-purple-500 to-purple-600 flex items-center justify-center shadow-2xl">
            <Zap className="w-10 h-10 text-white animate-bounce" style={{ animationDelay: '0s' }} />
          </div>
        </div>

        {/* Welcome text */}
        <div className="space-y-4 max-w-xl">
          <h1 className="text-5xl font-bold bg-gradient-to-r from-slate-900 via-violet-700 to-slate-900 dark:from-white dark:via-violet-300 dark:to-white bg-clip-text text-transparent">
            Welcome to Spectra
          </h1>
          <p className="text-lg text-slate-600 dark:text-slate-300 font-medium">
            Your powerful API testing companion for Laravel applications
          </p>
          <p className="text-sm text-slate-500 dark:text-slate-400">
            Let's set up everything to get you started in just a few minutes.
          </p>
        </div>

        {/* Feature highlights with better cards */}
        <div className="grid grid-cols-3 gap-4 w-full max-w-2xl py-4">
          {[
            { icon: Sparkles, label: 'Auto Discovery', desc: 'Find all endpoints automatically' },
            { icon: Code2, label: 'Live Testing', desc: 'Test in real-time with ease' },
            { icon: Rocket, label: 'Fast & Native', desc: 'Desktop-grade performance' }
          ].map((feature, idx) => {
            const Icon = feature.icon;
            return (
              <div
                key={idx}
                className="group rounded-xl border border-slate-200 dark:border-white/10 p-4 transition-all duration-300 hover:border-slate-300 dark:hover:border-white/20 hover:shadow-lg hover:shadow-violet-200 dark:hover:shadow-violet-900/20"
                style={{
                  animation: `slide-up 0.6s ease-out ${idx * 0.1}s both`
                }}
              >
                <div className="flex justify-center mb-3">
                  <div className="p-3 rounded-lg bg-gradient-to-br from-violet-100 to-purple-100 dark:from-violet-900/40 dark:to-purple-900/40 group-hover:scale-110 transition-transform">
                    <Icon className="w-6 h-6 text-violet-600 dark:text-violet-400" />
                  </div>
                </div>
                <h3 className="font-semibold text-sm text-slate-900 dark:text-white mb-1">
                  {feature.label}
                </h3>
                <p className="text-xs text-slate-600 dark:text-slate-400">
                  {feature.desc}
                </p>
              </div>
            );
          })}
        </div>

        {/* CTA Button */}
        {onNext && (
          <div className="pt-4">
            <OnboardingButton onClick={onNext} className="group">
              <span className="flex items-center gap-2">
                Get Started
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </span>
            </OnboardingButton>
          </div>
        )}

        {/* Footer hint */}
        <p className="text-xs text-slate-500 dark:text-slate-500 pt-4">
          Takes about 5 minutes to complete
        </p>
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
