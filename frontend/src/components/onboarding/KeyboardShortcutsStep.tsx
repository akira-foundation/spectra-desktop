import { Keyboard, Copy, Send, Save, ArrowRight } from 'lucide-react'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface KeyboardShortcutsStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function KeyboardShortcutsStep({ onNext, onBack }: KeyboardShortcutsStepProps) {
  const shortcuts = [
    {
      keys: ['⌘', 'K'],
      description: 'Open command palette',
      icon: <Keyboard className="w-5 h-5" />
    },
    {
      keys: ['⌘', 'Enter'],
      description: 'Send request',
      icon: <Send className="w-5 h-5" />
    },
    {
      keys: ['⌘', 'S'],
      description: 'Save request',
      icon: <Save className="w-5 h-5" />
    },
    {
      keys: ['⌘', 'C'],
      description: 'Copy response',
      icon: <Copy className="w-5 h-5" />
    }
  ]

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Keyboard Shortcuts
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            Speed up your workflow with these essential shortcuts
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4 w-full max-w-2xl">
          {shortcuts.map((shortcut, idx) => (
            <div
              key={idx}
              className="group rounded-xl border border-slate-200 dark:border-white/10 p-4 hover:border-slate-300 dark:hover:border-white/20 hover:shadow-md transition-all duration-300"
              style={{
                animation: `slide-up 0.6s ease-out ${idx * 0.1}s both`
              }}
            >
              <div className="flex items-start gap-3">
                <div className="mt-0.5 p-2 rounded-lg bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-400 group-hover:scale-110 transition-transform">
                  {shortcut.icon}
                </div>
                <div className="flex-1 text-left">
                  <div className="flex gap-1 mb-2">
                    {shortcut.keys.map((key, i) => (
                      <div key={i}>
                        <kbd className="px-2 py-1 rounded border border-slate-300 dark:border-slate-600 bg-slate-100 dark:bg-slate-800 text-xs font-semibold text-slate-700 dark:text-slate-300">
                          {key}
                        </kbd>
                        {i < shortcut.keys.length - 1 && (
                          <span className="mx-1 text-slate-400">+</span>
                        )}
                      </div>
                    ))}
                  </div>
                  <p className="text-sm text-slate-600 dark:text-slate-400 font-medium">
                    {shortcut.description}
                  </p>
                </div>
              </div>
            </div>
          ))}
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
