import { Zap, Shield, Cookie, Boxes, Share2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useUIStore } from '@/store/uiStore'

export function Welcome() {
  const setAddProjectOpen = useUIStore((s) => s.setAddProjectOpen)
  return (
    <div className="flex flex-col items-center justify-center h-full p-8">
      <div className="max-w-xl w-full space-y-8">
        <div className="text-center space-y-3">
          <div className="inline-flex p-3 rounded-xl bg-primary/10 ring-1 ring-primary/20">
            <Zap className="h-8 w-8 text-primary" />
          </div>
          <div className="space-y-1.5">
            <h2 className="text-2xl font-semibold tracking-tight">Welcome to Spectra</h2>
            <p className="text-sm text-foreground/60 max-w-md mx-auto">
              Local-first API testing for Laravel and beyond
            </p>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2.5">
          {[
            { num: '1', title: 'Browse', desc: 'Explore endpoints in the sidebar' },
            { num: '2', title: 'Configure', desc: 'Set headers, body, and params' },
            { num: '3', title: 'Execute', desc: 'Send and inspect the response' },
          ].map((step) => (
            <div
              key={step.num}
              className="rounded-md border border-border/60 bg-card/40 p-3 hover:bg-card/70 transition-colors"
            >
              <div className="flex items-center justify-center w-7 h-7 rounded-md bg-primary/10 mb-2 mx-auto">
                <span className="text-[11px] font-semibold text-primary">{step.num}</span>
              </div>
              <h4 className="font-medium text-[12.5px] text-center mb-1">{step.title}</h4>
              <p className="text-[11px] text-foreground/60 text-center leading-relaxed">{step.desc}</p>
            </div>
          ))}
        </div>

        <div className="space-y-2">
          <p className="text-[10px] font-semibold text-foreground/50 uppercase tracking-wider">
            Features
          </p>
          <div className="grid grid-cols-2 gap-2">
            {[
              { icon: Shield, label: 'Authentication', desc: 'Bearer, Basic, Sanctum' },
              { icon: Cookie, label: 'Cookies', desc: 'Manage session data' },
              { icon: Boxes, label: 'Collections', desc: 'Save and organize' },
              { icon: Share2, label: 'Real-time', desc: 'Instant responses' },
            ].map((feature) => {
              const Icon = feature.icon
              return (
                <div
                  key={feature.label}
                  className="flex items-start gap-2.5 p-2.5 rounded-md border border-border/60 bg-card/30 hover:bg-card/60 transition-colors"
                >
                  <Icon className="h-3.5 w-3.5 text-primary shrink-0 mt-0.5" />
                  <div>
                    <p className="text-[12px] font-medium">{feature.label}</p>
                    <p className="text-[10.5px] text-foreground/60 mt-0.5">{feature.desc}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        <div className="text-center pt-2">
          <Button size="sm" className="font-medium" onClick={() => setAddProjectOpen(true)}>
            Add Project
          </Button>
        </div>
      </div>
    </div>
  )
}
