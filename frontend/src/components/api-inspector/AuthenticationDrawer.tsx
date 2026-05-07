import { User, Key, Lock, Shield } from 'lucide-react'
import { Drawer, DrawerContent, DrawerHeader, DrawerTitle, DrawerDescription } from '@/components/ui/drawer'
import { useUIStore } from '@/store/uiStore'

interface AuthenticationDrawerProps {
  activeMethod: string
  onMethodChange: (method: string) => void
}

export function AuthenticationDrawer({ activeMethod, onMethodChange }: AuthenticationDrawerProps) {
  const isOpen = useUIStore((state) => state.isAuthDrawerOpen)
  const setAuthDrawerOpen = useUIStore((state) => state.setAuthDrawerOpen)

  const authMethods = [
    { id: 'current-user', icon: User, label: 'Current User', description: 'Use the authenticated user context' },
    { id: 'impersonate', icon: Key, label: 'Impersonate', description: 'Impersonate another user' },
    { id: 'bearer-token', icon: Lock, label: 'Bearer Token', description: 'Use a bearer token for authentication' },
    { id: 'basic-auth', icon: Shield, label: 'Basic Auth', description: 'Use basic HTTP authentication' },
  ]

  const handleSelectMethod = (methodId: string) => {
    onMethodChange(methodId)
    setAuthDrawerOpen(false)
  }

  return (
    <Drawer open={isOpen} onOpenChange={setAuthDrawerOpen}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>Authentication Method</DrawerTitle>
          <DrawerDescription>Select how to authenticate requests</DrawerDescription>
        </DrawerHeader>

        <div className="space-y-2 px-4 pb-6">
          {authMethods.map((method) => {
            const Icon = method.icon
            const isActive = activeMethod === method.id

            return (
              <button
                key={method.id}
                onClick={() => handleSelectMethod(method.id)}
                className={`w-full flex items-start gap-3 p-3 rounded-lg border transition-colors ${
                  isActive
                    ? 'border-primary bg-primary/5'
                    : 'border-border/50 hover:border-border'
                }`}
              >
                <Icon className={`w-4 h-4 mt-0.5 flex-shrink-0 ${isActive ? 'text-primary' : 'text-muted-foreground'}`} />
                <div className="flex-1 text-left">
                  <div className={`text-sm font-medium ${isActive ? 'text-foreground' : 'text-foreground'}`}>
                    {method.label}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {method.description}
                  </div>
                </div>
              </button>
            )
          })}
        </div>
      </DrawerContent>
    </Drawer>
  )
}
