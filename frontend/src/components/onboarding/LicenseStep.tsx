import { useState } from 'react'
import { Key, ExternalLink, Globe, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { OnboardingContainer } from './OnboardingContainer'
import { OnboardingButtonGroup } from './OnboardingButtonGroup'
import { OnboardingButton } from './OnboardingButton'

interface LicenseStepProps {
  onBack?: () => void;
  onNext?: () => void;
}

export function LicenseStep({ onBack, onNext }: LicenseStepProps) {
  const [licenseKey, setLicenseKey] = useState('')
  const [isValidating, setIsValidating] = useState(false)

  const handleActivate = async () => {
    if (!licenseKey.trim()) return

    setIsValidating(true)

    // Simulate license validation
    await new Promise(resolve => setTimeout(resolve, 1500))

    // Save license
    localStorage.setItem('spectra-license', licenseKey)

    setIsValidating(false)

    // Go to celebrate step
    if (onNext) {
      setTimeout(() => onNext(), 500)
    }
  }

  const handleSkip = () => {
    localStorage.setItem('spectra-license-skipped', 'true')

    // Go to celebrate step
    if (onNext) {
      onNext()
    }
  }

  return (
    <OnboardingContainer>
      <div className="flex flex-col items-center text-center">
        <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mb-6">
          <Key className="w-8 h-8 text-primary" />
        </div>

        <h2 className="text-2xl font-semibold mb-3">Activate Spectra</h2>
        <p className="text-muted-foreground mb-8 max-w-md">
          Enter your license key to unlock all features
        </p>

        <div className="w-full max-w-md space-y-4 mb-8">
          <div className="space-y-2 text-left">
            <Label htmlFor="license-key">License Key</Label>
            <Input
              id="license-key"
              placeholder="XXXX-XXXX-XXXX-XXXX"
              value={licenseKey}
              onChange={(e) => setLicenseKey(e.target.value.toUpperCase())}
              className="h-12 font-mono text-center tracking-widest"
            />
          </div>

          <Button 
            onClick={handleActivate}
            disabled={!licenseKey.trim() || isValidating}
            variant="gradient"
            className="w-full h-12"
          >
            {isValidating ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                Validating...
              </>
            ) : (
              <>
                <Check className="w-4 h-4 mr-2" />
                Activate License
              </>
            )}
          </Button>
        </div>

        <div className="w-full max-w-md">
          <div className="relative my-6">
            <Separator />
            <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-background px-3 text-xs text-muted-foreground">
              Don't have a license?
            </span>
          </div>

          <div className="grid grid-cols-2 gap-3 mb-8">
            <Button
              variant="outline"
              size="sm"
              className="h-auto py-4 flex flex-col items-center gap-2"
              onClick={() => window.open('https://spectra.dev/buy', '_blank')}
            >
              <ExternalLink className="w-5 h-5 text-primary" />
              <div className="text-xs">
                <div className="font-medium">Purchase</div>
              </div>
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-auto py-4 flex flex-col items-center gap-2"
              onClick={() => window.open('https://spectra.dev', '_blank')}
            >
              <Globe className="w-5 h-5 text-primary" />
              <div className="text-xs">
                <div className="font-medium">Web Version</div>
              </div>
            </Button>
          </div>
        </div>

        <OnboardingButtonGroup>
          <OnboardingButton variant="secondary" onClick={onBack}>
            Back
          </OnboardingButton>
        </OnboardingButtonGroup>
      </div>
    </OnboardingContainer>
  )
}
