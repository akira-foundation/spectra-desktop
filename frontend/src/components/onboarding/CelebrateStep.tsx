import { Sparkles, Rocket, Star } from 'lucide-react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButton } from './OnboardingButton'

interface CelebrateStepProps {
  onComplete?: () => void;
}

export function CelebrateStep({ onComplete }: CelebrateStepProps) {
  const handleComplete = () => {
    localStorage.setItem('spectra-onboarded', 'true')
    if (onComplete) {
      onComplete()
    }
  }

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        {/* Animated celebration icon */}
        <div className="relative">
          <div className="absolute inset-0 rounded-full bg-gradient-to-r from-violet-500 to-purple-600 opacity-20 blur-3xl animate-pulse" />
          <div className="relative flex items-center justify-center w-24 h-24 rounded-full bg-gradient-to-br from-violet-500 via-purple-500 to-violet-600 shadow-2xl">
            <Sparkles className="w-12 h-12 text-white animate-bounce" />
          </div>
        </div>

        {/* Success message */}
        <div className="space-y-4">
          <h2 className="text-5xl font-bold text-slate-900 dark:text-white">
            You're All Set!
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-300 max-w-lg">
            Welcome to Spectra. Your API testing journey starts now.
          </p>
        </div>

        {/* Features unlocked */}
        <div className="space-y-3 w-full max-w-2xl">
          <p className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
            You now have access to:
          </p>
          <div className="grid grid-cols-3 gap-3">
            {[
              { icon: <Rocket className="w-5 h-5" />, label: 'Auto-Discovery' },
              { icon: <Star className="w-5 h-5" />, label: 'Live Testing' },
              { icon: <Sparkles className="w-5 h-5" />, label: 'Real-Time Sync' }
            ].map((item, idx) => (
              <div
                key={idx}
                className="rounded-lg border border-slate-200 dark:border-white/10 p-4 flex flex-col items-center gap-2"
                style={{
                  animation: `slide-up 0.6s ease-out ${idx * 0.1}s both`
                }}
              >
                <div className="p-2 rounded-lg bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-400">
                  {item.icon}
                </div>
                <span className="text-xs font-medium text-slate-700 dark:text-slate-300">
                  {item.label}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* CTA */}
        <div className="pt-4">
          <OnboardingButton onClick={handleComplete} className="group">
            <span className="flex items-center gap-2">
              Start Using Spectra
            </span>
          </OnboardingButton>
        </div>

        {/* Footer hint */}
        <p className="text-xs text-slate-500 dark:text-slate-500 pt-4">
          💡 Tip: Visit the settings anytime to adjust your preferences
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
