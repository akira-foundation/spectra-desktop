import { KeyRound } from 'lucide-react'
import { Card } from './Card'

interface AuthState {
  user?: { name?: string; username?: string; email?: string; role?: string } | null
  hasToken?: boolean
  tokenPreview?: string
}

export function AuthCard({ auth }: { auth: AuthState | null }) {
  return (
    <Card title="Authentication" icon={KeyRound}>
      {auth?.user ? (
        <div className="space-y-1">
          <p className="text-[13px] font-medium truncate">
            {auth.user.name || auth.user.username || auth.user.email || 'User'}
          </p>
          {auth.user.email && <p className="text-[11px] text-muted-foreground truncate">{auth.user.email}</p>}
          {auth.user.role && <p className="text-[10.5px] text-muted-foreground/80">{auth.user.role}</p>}
          {auth.hasToken && (
            <div className="flex items-center gap-1.5 pt-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              <span className="text-[10.5px] text-muted-foreground font-mono truncate">{auth.tokenPreview}</span>
            </div>
          )}
        </div>
      ) : auth?.hasToken ? (
        <div className="space-y-1">
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span className="text-[12px] font-medium">Token active</span>
          </div>
          <p className="text-[10.5px] text-muted-foreground font-mono truncate">{auth.tokenPreview}</p>
        </div>
      ) : (
        <p className="text-[11.5px] italic text-muted-foreground">No active session.</p>
      )}
    </Card>
  )
}
