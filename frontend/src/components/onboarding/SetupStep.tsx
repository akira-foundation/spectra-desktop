import { useState } from 'react'
import { FolderOpen } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface SetupStepProps {
  onNext?: () => void;
  onBack?: () => void;
}

export function SetupStep({ onNext, onBack }: SetupStepProps) {
  const [projectName, setProjectName] = useState('')
  const [apiUrl, setApiUrl] = useState('http://localhost:8000')
  const [projectPath, setProjectPath] = useState('')
  const [isSelecting, setIsSelecting] = useState(false)

  const handleOpenFolder = async () => {
    setIsSelecting(true)
    try {
      const selected = window.prompt(
        'Enter Laravel project absolute path:',
        projectPath || '/Users/you/projects/my-laravel-app',
      )
      if (selected && selected.trim()) {
        const path = selected.trim()
        setProjectPath(path)
        const folderName = path.split(/[/\\]/).pop() || 'My Laravel Project'
        setProjectName(folderName)
      }
    } catch (error) {
      console.error('Error opening folder:', error)
    } finally {
      setIsSelecting(false)
    }
  }

  const handleContinue = () => {
    if (projectName.trim()) {
      localStorage.setItem('spectra-project', JSON.stringify({
        name: projectName,
        apiUrl,
        path: projectPath
      }))
      onNext?.()
    }
  }

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center justify-center text-center space-y-8">
        <div className="space-y-2">
          <h2 className="text-4xl font-bold text-slate-900 dark:text-white">
            Connect Your Project
          </h2>
          <p className="text-base text-slate-600 dark:text-slate-300 max-w-lg">
            Select your Laravel project folder and we'll configure everything automatically
          </p>
        </div>

        <div className="w-full max-w-2xl space-y-6">
          {/* Project Folder Card */}
          <button
            onClick={handleOpenFolder}
            disabled={isSelecting}
            className="w-full group relative rounded-2xl border-2 border-dashed border-slate-300 dark:border-slate-700 hover:border-violet-400 dark:hover:border-violet-500 p-8 transition-all duration-300 text-left bg-white dark:bg-transparent hover:bg-slate-50 dark:hover:bg-slate-800/50"
          >
            <div className="flex flex-col items-center justify-center gap-4 text-center">
              <div className="p-4 rounded-xl bg-gradient-to-br from-violet-100 to-purple-100 dark:from-violet-900/40 dark:to-purple-900/40 group-hover:scale-110 transition-transform">
                <FolderOpen className="w-8 h-8 text-violet-600 dark:text-violet-400" />
              </div>
              <div>
                <h3 className="font-semibold text-base text-slate-900 dark:text-white mb-1">
                  {projectPath ? 'Change Project Folder' : 'Select Project Folder'}
                </h3>
                {projectPath ? (
                  <p className="text-sm font-mono text-slate-600 dark:text-slate-400 break-all">
                    {projectPath}
                  </p>
                ) : (
                  <p className="text-sm text-slate-600 dark:text-slate-400">
                    Click to browse your Laravel project directory
                  </p>
                )}
              </div>
            </div>
          </button>

          {/* Form Fields */}
          <div className="space-y-5">
            {/* Project Name */}
            <div className="space-y-2">
              <Label htmlFor="project-name" className="text-sm font-semibold text-slate-900 dark:text-white">
                Project Name
              </Label>
              <Input
                id="project-name"
                placeholder="My Laravel Project"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                className="h-11 rounded-lg border-slate-300 dark:border-slate-700 text-base"
              />
            </div>

            {/* API URL */}
            <div className="space-y-2">
              <Label htmlFor="api-url" className="text-sm font-semibold text-slate-900 dark:text-white">
                API Base URL
              </Label>
              <Input
                id="api-url"
                placeholder="http://localhost:8000"
                value={apiUrl}
                onChange={(e) => setApiUrl(e.target.value)}
                className="h-11 rounded-lg border-slate-300 dark:border-slate-700 font-mono text-sm"
              />
              <p className="text-xs text-slate-600 dark:text-slate-400">
                Automatically detected from your <code className="bg-slate-200 dark:bg-slate-800 px-2 py-1 rounded">.env</code> file
              </p>
            </div>
          </div>

        </div>

        <OnboardingButtonGroup>
          {onBack && (
            <OnboardingButton variant="secondary" onClick={onBack}>
              Back
            </OnboardingButton>
          )}
          <OnboardingButton
            onClick={handleContinue}
            disabled={!projectName.trim()}
            className="disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Continue
          </OnboardingButton>
        </OnboardingButtonGroup>
      </div>
    </OnboardingContainer>
  )
}
