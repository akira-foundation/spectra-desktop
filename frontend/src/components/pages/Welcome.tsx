import { Zap, Shield, Cookie, Boxes, Share2 } from 'lucide-react'
import {Button} from "@/components/ui/button";

export function Welcome() {
  return (
    <div className="flex flex-col items-center justify-center h-full p-8">
      <div className="max-w-2xl w-full space-y-8">
        {/* Hero Section */}
        <div className="text-center space-y-4">
          <div className="inline-flex p-4 rounded-2xl bg-gradient-to-br from-primary/20 to-primary/10 ring-1 ring-primary/20">
            <Zap className="h-12 w-12 text-primary" />
          </div>
          <div className="space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Welcome to Spectra</h2>
            <p className="text-base text-foreground/60 max-w-lg mx-auto">
              Professional API testing and request building at your fingertips
            </p>
          </div>
        </div>

        {/* Quick Start Steps */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            { num: '1', title: 'Browse Endpoints', desc: 'Explore all available API endpoints in the left sidebar' },
            { num: '2', title: 'Select & Configure', desc: 'Click any endpoint to configure headers, body, and parameters' },
            { num: '3', title: 'Execute & View', desc: 'Execute your request and see the response instantly' },
          ].map((step) => (
            <div key={step.num} className="relative group">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/20 to-transparent opacity-0 group-hover:opacity-100 rounded-xl transition-opacity" />
              <div className="relative p-4 rounded-xl border border-border/50 bg-card/30 hover:bg-card/60 transition-all h-full">
                <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-primary/10 mb-3 mx-auto">
                  <span className="text-sm font-bold text-primary">{step.num}</span>
                </div>
                <h4 className="font-semibold text-sm text-center mb-2">{step.title}</h4>
                <p className="text-xs text-foreground/60 text-center leading-relaxed">
                  {step.desc}
                </p>
              </div>
            </div>
          ))}
        </div>

        {/* Features Section */}
        <div className="space-y-3 pt-4">
          <p className="text-xs font-semibold text-foreground/60 uppercase tracking-wide">Features</p>
          <div className="grid grid-cols-2 gap-3">
            {[
              { icon: Shield, label: 'Authentication', desc: 'Bearer, Basic, & more' },
              { icon: Cookie, label: 'Cookies', desc: 'Manage session data' },
              { icon: Boxes, label: 'Collections', desc: 'Save & organize' },
              { icon: Share2, label: 'Real-time', desc: 'Instant responses' },
            ].map((feature) => {
              const Icon = feature.icon
              return (
                <div key={feature.label} className="flex items-start gap-3 p-3 rounded-lg border border-border/50 bg-card/20 hover:bg-card/40 transition-colors">
                  <Icon className="h-4 w-4 text-primary flex-shrink-0 mt-1" />
                  <div>
                    <p className="text-xs font-medium">{feature.label}</p>
                    <p className="text-[10px] text-foreground/60 mt-0.5">{feature.desc}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* CTA */}
        <div className="text-center pt-4">
          <p className="text-sm text-foreground/60">
            <Button className="font-medium " variant="gradient">Add Project</Button>
          </p>
        </div>
      </div>
    </div>
  )
}
