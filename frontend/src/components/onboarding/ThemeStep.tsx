import { useState } from 'react'
import { Moon, Sun, Monitor, Check } from 'lucide-react'
import { useTheme } from '@/hooks/useTheme'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButton } from './OnboardingButton'

interface ThemeStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function ThemeStep({ onNext, onBack }: ThemeStepProps) {
  const setTheme = useTheme((state) => state.setTheme)

  const themes = [
    {
      id: 'light' as const,
      icon: Sun,
      title: 'Light',
      description: 'Clean and bright interface, perfect for daylight environments',
      gradient: 'from-amber-400 to-orange-500',
      preview: 'bg-white border-slate-200'
    },
    {
      id: 'dark' as const,
      icon: Moon,
      title: 'Dark',
      description: 'Easy on the eyes, ideal for long coding sessions',
      gradient: 'from-violet-500 to-purple-600',
      preview: 'bg-slate-900 border-slate-700'
    },
    {
      id: 'system' as const,
      icon: Monitor,
      title: 'System',
      description: 'Automatically matches your OS theme preferences',
      gradient: 'from-blue-500 to-cyan-500',
      preview: 'bg-gradient-to-r from-white to-slate-900'
    }
  ]

  const [selectedTheme, setSelectedTheme] = useState<'light' | 'dark' | 'system' | null>(null)

  const handleThemeSelect = (themeId: 'light' | 'dark' | 'system') => {
    setSelectedTheme(themeId)
    setTheme(themeId)
    setTimeout(() => onNext?.(), 300)
  }

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Choose Your Appearance
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            Pick a theme that suits your workflow and keeps you comfortable
          </p>
        </div>

        <div className="grid grid-cols-3 gap-4 w-full max-w-3xl">
          {themes.map((theme) => {
            const Icon = theme.icon
            const isSelected = selectedTheme === theme.id

            return (
              <button
                key={theme.id}
                onClick={() => handleThemeSelect(theme.id)}
                className={`
                  relative rounded-2xl border p-6 transition-all duration-300 text-left
                  ${
                    isSelected
                      ? 'border-violet-500 shadow-lg shadow-violet-200 dark:shadow-violet-900/20'
                      : 'border-slate-200 dark:border-white/10 hover:border-slate-300 dark:hover:border-white/20 hover:shadow-md'
                  }
                `}
              >
                {/* Preview */}
                <div className={`mb-4 h-16 rounded-lg border ${theme.preview} flex items-center justify-center overflow-hidden relative`}>
                  <div className="flex gap-2">
                    <div className="h-2 w-2 rounded-full bg-slate-300 dark:bg-slate-600" />
                    <div className="h-2 w-3 rounded-full bg-slate-400 dark:bg-slate-500" />
                    <div className="h-2 w-2 rounded-full bg-slate-300 dark:bg-slate-600" />
                  </div>
                </div>

                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex items-center gap-2">
                    <div className={`p-2 rounded-lg bg-gradient-to-br ${theme.gradient}`}>
                      <Icon className="w-5 h-5 text-white" />
                    </div>
                    <h3 className="font-semibold text-slate-900 dark:text-white">
                      {theme.title}
                    </h3>
                  </div>

                  {isSelected && (
                    <div className="flex h-6 w-6 items-center justify-center rounded-full bg-violet-600">
                      <Check className="h-4 w-4 text-white" />
                    </div>
                  )}
                </div>

                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
                  {theme.description}
                </p>
              </button>
            )
          })}
        </div>

        <p className="text-xs text-slate-500 dark:text-slate-400 pt-2">
          You can always change this in settings later
        </p>

        {onBack && (
          <div className="flex gap-3 justify-center w-full pt-4">
            <OnboardingButton variant="secondary" onClick={onBack}>
              Back
            </OnboardingButton>
          </div>
        )}
      </div>
    </OnboardingContainer>
  )
}
