import { User, Key, Lock, Shield } from 'lucide-react'

interface AuthenticationTabsProps {
  activeMethod: string
  onMethodChange: (method: string) => void
}

export function AuthenticationTabs({ activeMethod, onMethodChange }: AuthenticationTabsProps) {
  const authMethods = [
    { id: 'current-user', icon: User, label: 'Current User' },
    { id: 'impersonate', icon: Key, label: 'Impersonate' },
    { id: 'bearer-token', icon: Lock, label: 'Bearer Token' },
    { id: 'basic-auth', icon: Shield, label: 'Basic Auth' },
  ]

  return (
    <div className="flex items-center gap-6 px-4 py-3 border-b border-border/50">
      <span className="text-xs font-medium text-muted-foreground uppercase">Auth</span>
      <div className="flex items-center gap-6">
        {authMethods.map((method) => {
          const Icon = method.icon
          const isActive = activeMethod === method.id
          return (
            <button
              key={method.id}
              onClick={() => onMethodChange(method.id)}
              className={`flex items-center gap-1.5 text-sm py-1 transition-colors ${
                isActive
                  ? 'text-foreground font-medium border-b-2 border-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <Icon className="w-4 h-4" />
              <span className="text-xs">{method.label}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
