import { useState } from 'react'
import { Copy, Trash2, KeyRound, UserRound, Mail, Hash, Shield } from 'lucide-react'
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerDescription,
} from '@/components/ui/drawer'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useUIStore } from '@/store/uiStore'
import { useAuthStore } from '@/store/authStore'
import { useProjectStore } from '@/store/projectStore'
import { SetProjectAuthManual } from '../../../wailsjs/go/app/App'
import { app } from '../../../wailsjs/go/models'
import { cn } from '@/lib/utils'

interface AuthenticationDrawerProps {
  activeMethod: string
  onMethodChange: (method: string) => void
}

export function AuthenticationDrawer(_props: AuthenticationDrawerProps) {
  const isOpen = useUIStore((s) => s.isAuthDrawerOpen)
  const setAuthDrawerOpen = useUIStore((s) => s.setAuthDrawerOpen)
  const activeProjectId = useProjectStore((s) => s.activeProjectId)
  const auth = useAuthStore((s) =>
    activeProjectId ? s.byProject[activeProjectId] ?? null : null,
  )
  const refreshAuth = useAuthStore((s) => s.refresh)
  const clearAuth = useAuthStore((s) => s.clear)
  const [manualToken, setManualToken] = useState('')
  const [pasting, setPasting] = useState(false)

  const handlePaste = async () => {
    if (!activeProjectId || !manualToken.trim()) return
    setPasting(true)
    try {
      await SetProjectAuthManual(
        app.SetProjectAuthInput.createFrom({
          projectID: activeProjectId,
          scheme: 'bearer',
          token: manualToken.trim(),
        }),
      )
      setManualToken('')
      await refreshAuth(activeProjectId)
    } finally {
      setPasting(false)
    }
  }

  const handleClear = async () => {
    if (!activeProjectId) return
    await clearAuth(activeProjectId)
  }

  const handleCopy = async () => {
    if (!auth?.tokenPreview) return
    try {
      await navigator.clipboard.writeText(auth.tokenPreview)
    } catch {}
  }

  return (
    <Drawer open={isOpen} onOpenChange={setAuthDrawerOpen} direction="right">
      <DrawerContent className="data-[vaul-drawer-direction=right]:sm:max-w-md !bg-sidebar">
        <DrawerHeader className="border-b border-border/60">
          <DrawerTitle className="text-[13px] font-semibold tracking-tight">
            Authentication
          </DrawerTitle>
          <DrawerDescription className="text-[11.5px]">
            Project-scoped credentials. Auto-captured from login responses or pasted manually.
          </DrawerDescription>
        </DrawerHeader>

        <div className="flex-1 overflow-auto px-4 py-4 space-y-5">
          <Section title="Authenticated user">
            {auth?.user ? (
              <div className="space-y-1.5">
                {auth.user.name && <UserRow icon={UserRound} label="Name" value={auth.user.name} />}
                {auth.user.username && <UserRow icon={Hash} label="Username" value={auth.user.username} />}
                {auth.user.email && <UserRow icon={Mail} label="Email" value={auth.user.email} />}
                {auth.user.role && <UserRow icon={Shield} label="Role" value={auth.user.role} />}
                {auth.user.id && <UserRow icon={Hash} label="ID" value={auth.user.id} />}
              </div>
            ) : (
              <Empty label="No user captured" />
            )}
          </Section>

          <Section title="Token">
            {auth?.hasToken ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <KeyRound className="w-3.5 h-3.5 text-emerald-500" />
                  <code className="font-mono text-[11.5px] text-foreground/85 truncate flex-1">
                    {auth.tokenPreview}
                  </code>
                  <Button variant="ghost" size="icon-sm" className="h-6 w-6" onClick={handleCopy}>
                    <Copy className="w-3 h-3" />
                  </Button>
                </div>
                {auth.tokenPath && (
                  <p className="text-[10.5px] text-muted-foreground">
                    Captured from <code className="font-mono">{auth.tokenPath}</code>
                  </p>
                )}
                {auth.capturedFromEndpoint && (
                  <p className="text-[10.5px] text-muted-foreground">
                    Source endpoint: <code className="font-mono">{auth.capturedFromEndpoint}</code>
                  </p>
                )}
              </div>
            ) : (
              <Empty label="No token" />
            )}
          </Section>

          <Section title="Paste token manually">
            <div className="flex gap-2">
              <Input
                value={manualToken}
                onChange={(e) => setManualToken(e.target.value)}
                placeholder="Bearer token"
                className="h-7 text-[12px] font-mono"
              />
              <Button
                size="sm"
                disabled={!manualToken.trim() || pasting || !activeProjectId}
                onClick={handlePaste}
                className="h-7 text-[11px]"
              >
                Save
              </Button>
            </div>
          </Section>

          {auth && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleClear}
              className="w-full h-7 text-[11px] text-destructive hover:text-destructive"
            >
              <Trash2 className="w-3 h-3" />
              Clear authentication
            </Button>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  )
}

interface SectionProps {
  title: string
  children: React.ReactNode
}

function Section({ title, children }: SectionProps) {
  return (
    <div className="space-y-2">
      <h4 className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        {title}
      </h4>
      {children}
    </div>
  )
}

function Empty({ label }: { label: string }) {
  return <p className="text-[11px] italic text-muted-foreground/70">{label}</p>
}

function UserRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
}) {
  return (
    <div className={cn('flex items-center gap-2 text-[12px]')}>
      <Icon className="w-3.5 h-3.5 text-muted-foreground" />
      <span className="text-muted-foreground/80 w-16">{label}</span>
      <span className="font-medium text-foreground/90 truncate">{value}</span>
    </div>
  )
}
