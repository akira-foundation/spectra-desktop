import { CheckCircle, AlertCircle, Loader2 } from 'lucide-react'

interface StatusBarProps {
  sdkStatus?: 'connected' | 'disconnected' | 'syncing' | 'error'
  lastSyncTime?: Date | null
  coreStatus?: 'ready' | 'initializing' | 'error'
}

export function StatusBar({
  sdkStatus = 'connected',
  lastSyncTime,
  coreStatus = 'ready',
}: StatusBarProps) {
  const formatTime = (date: Date | null | undefined) => {
    if (!date) return 'Never'
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 1) return 'Just now'
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }

  const getSdkIcon = () => {
    switch (sdkStatus) {
      case 'syncing':
        return <Loader2 className="w-3 h-3 animate-spin text-info" />
      case 'connected':
        return <CheckCircle className="w-3 h-3 text-success" />
      case 'error':
        return <AlertCircle className="w-3 h-3 text-error" />
      default:
        return <AlertCircle className="w-3 h-3 text-warning" />
    }
  }

  const getSdkLabel = () => {
    switch (sdkStatus) {
      case 'syncing':
        return 'Syncing...'
      case 'connected':
        return 'Connected'
      case 'error':
        return 'Error'
      default:
        return 'Disconnected'
    }
  }

  return (
    <div className="h-10 border-t border-border bg-card flex items-center justify-between px-6 text-xs text-foreground/50">
      {/* Left: Status Indicators */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          {getSdkIcon()}
          <span>SDK: {getSdkLabel()}</span>
        </div>
        <div className="flex items-center gap-2">
          <CheckCircle className="w-3 h-3 text-foreground/30" />
          <span>Last Sync: {formatTime(lastSyncTime)}</span>
        </div>
        <div className="flex items-center gap-2">
          {coreStatus === 'ready' ? (
            <CheckCircle className="w-3 h-3 text-foreground/30" />
          ) : (
            <Loader2 className="w-3 h-3 animate-spin text-foreground/30" />
          )}
          <span>Core: {coreStatus === 'ready' ? 'Ready' : 'Initializing'}</span>
        </div>
      </div>

      {/* Right: App Version */}
      <div className="text-foreground/40">v0.1.0</div>
    </div>
  )
}
