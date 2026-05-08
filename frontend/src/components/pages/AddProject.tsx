import { useState } from 'react'
import { useProjectStore } from '@/store/projectStore'
import { ChevronRight, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

type Step = 'framework' | 'folder' | 'install' | 'test' | 'sync' | 'success'

export function AddProject() {
  const [currentStep, setCurrentStep] = useState<Step>('framework')
  const [projectData, setProjectData] = useState({
    name: '',
    path: '',
    framework: '' as 'laravel' | 'symfony' | 'other',
    frameworkVersion: '',
  })
  const addProjectFromInput = useProjectStore((state) => state.addProjectFromInput)

  const steps: { id: Step; title: string; description: string }[] = [
    { id: 'framework', title: 'Choose Framework', description: 'Select your project framework' },
    { id: 'folder', title: 'Select Folder', description: 'Choose your project directory' },
    { id: 'install', title: 'Install SDK', description: 'Installing Spectra SDK...' },
    { id: 'test', title: 'Test Connection', description: 'Testing SDK connection...' },
    { id: 'sync', title: 'Initial Sync', description: 'Syncing project data...' },
    { id: 'success', title: 'Success', description: 'Project added successfully!' },
  ]

  const currentStepIndex = steps.findIndex((s) => s.id === currentStep)
  const stepProgress = ((currentStepIndex + 1) / steps.length) * 100

  const handleNext = () => {
    const stepOrder: Step[] = ['framework', 'folder', 'install', 'test', 'sync', 'success']
    const nextIndex = stepOrder.indexOf(currentStep) + 1
    if (nextIndex < stepOrder.length) {
      setCurrentStep(stepOrder[nextIndex])
    }
  }

  const handleFrameworkSelect = (framework: 'laravel' | 'symfony' | 'other') => {
    setProjectData({ ...projectData, framework, frameworkVersion: framework === 'laravel' ? '11' : '6' })
    handleNext()
  }

  const handleAddProject = async () => {
    await addProjectFromInput({
      id: '',
      name: projectData.name || 'New Project',
      path: projectData.path,
      framework: projectData.framework,
      frameworkVersion: projectData.frameworkVersion,
      apiFilterMode: 'auto',
      apiFilterValue: '',
      baseUrl: '',
    })
    setCurrentStep('success')
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="glass-card w-full max-w-md p-8 space-y-6">
        {/* Progress Bar */}
        <div className="space-y-2">
          <div className="h-1 bg-background rounded-full overflow-hidden">
            <div className="h-full bg-primary transition-all" style={{ width: `${stepProgress}%` }} />
          </div>
          <p className="text-xs text-foreground/50 text-right">
            Step {currentStepIndex + 1} of {steps.length}
          </p>
        </div>

        {/* Title */}
        <div className="space-y-1">
          <h2 className="text-2xl font-bold">{steps[currentStepIndex].title}</h2>
          <p className="text-sm text-foreground/60">{steps[currentStepIndex].description}</p>
        </div>

        {/* Content */}
        <div className="space-y-4">
          {currentStep === 'framework' && (
            <div className="space-y-3">
              {[
                { label: 'Laravel', value: 'laravel' as const },
                { label: 'Symfony', value: 'symfony' as const },
                { label: 'Other PHP Framework', value: 'other' as const },
              ].map((option) => (
                <Button
                  key={option.value}
                  variant="outline"
                  onClick={() => handleFrameworkSelect(option.value)}
                  className="w-full justify-start h-auto p-4"
                >
                  <p className="font-medium">{option.label}</p>
                </Button>
              ))}
            </div>
          )}

          {currentStep === 'folder' && (
            <div className="space-y-3">
              <Input
                type="text"
                placeholder="Project name"
                value={projectData.name}
                onChange={(e) => setProjectData({ ...projectData, name: e.target.value })}
              />
              <Input
                type="text"
                placeholder="/path/to/project"
                value={projectData.path}
                onChange={(e) => setProjectData({ ...projectData, path: e.target.value })}
              />
              <div className="flex gap-2">
                <Button
                  onClick={() => setCurrentStep('install')}
                  className="flex-1 gap-2"
                >
                  <span>Continue</span>
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}

          {currentStep === 'install' && (
            <div className="space-y-3">
              <div className="h-2 bg-background rounded-full overflow-hidden">
                <div className="h-full bg-primary animate-pulse" style={{ width: '66%' }} />
              </div>
              <p className="text-sm text-foreground/60">Installing Spectra SDK package...</p>
              <p className="text-xs text-foreground/40">This may take a moment</p>
              <Button
                onClick={handleNext}
                className="w-full"
              >
                Next
              </Button>
            </div>
          )}

          {currentStep === 'test' && (
            <div className="space-y-3">
              <div className="p-4 bg-background rounded-lg space-y-2">
                <div className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-500" />
                  <span className="text-sm">SDK package verified</span>
                </div>
                <div className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-500" />
                  <span className="text-sm">Service provider detected</span>
                </div>
                <div className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-500" />
                  <span className="text-sm">Connection established</span>
                </div>
              </div>
              <Button
                onClick={handleNext}
                className="w-full"
              >
                Next
              </Button>
            </div>
          )}

          {currentStep === 'sync' && (
            <div className="space-y-3">
              <div className="h-2 bg-background rounded-full overflow-hidden">
                <div className="h-full bg-primary animate-pulse" style={{ width: '75%' }} />
              </div>
              <p className="text-sm text-foreground/60">Syncing project data...</p>
              <div className="text-xs text-foreground/40 space-y-1">
                <p>✓ Scanning routes: 45 found</p>
                <p>✓ Analyzing models: 12 found</p>
                <p>✓ Extracting middleware: 8</p>
              </div>
              <Button
                onClick={handleAddProject}
                className="w-full"
              >
                Finish
              </Button>
            </div>
          )}

          {currentStep === 'success' && (
            <div className="space-y-4 text-center">
              <div className="w-16 h-16 rounded-full bg-green-500/10 flex items-center justify-center mx-auto">
                <Check className="w-8 h-8 text-green-500" />
              </div>
              <div>
                <p className="font-semibold">{projectData.name}</p>
                <p className="text-sm text-foreground/60">{projectData.framework}</p>
              </div>
              <Button className="w-full">
                Go to Dashboard
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
