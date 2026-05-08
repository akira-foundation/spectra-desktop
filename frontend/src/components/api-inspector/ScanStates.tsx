import { Loader2, AlertTriangle, Inbox, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/common/EmptyState'
import type { ScanError } from '@/store/endpointsStore'

export function ScanLoadingState() {
  return (
    <EmptyState
      size="sm"
      icon={Loader2}
      title="Scanning routes..."
      description="Running php artisan route:list against your project."
    />
  )
}

interface ScanErrorStateProps {
  error: ScanError
  onRetry?: () => void
}

export function ScanErrorState({ error, onRetry }: ScanErrorStateProps) {
  const copy = errorCopy(error)
  return (
    <EmptyState
      icon={AlertTriangle}
      title={copy.title}
      description={copy.description}
      hint={
        <code className="text-[10.5px] font-mono text-muted-foreground/80">{error.message}</code>
      }
      action={
        onRetry && (
          <Button size="sm" variant="outline" onClick={onRetry}>
            <RefreshCw className="w-3.5 h-3.5" />
            Retry
          </Button>
        )
      }
    />
  )
}

export function NoRoutesState({ onRetry }: { onRetry?: () => void }) {
  return (
    <EmptyState
      size="sm"
      icon={Inbox}
      title="No routes found"
      description="Define routes in routes/api.php or routes/web.php and rescan."
      action={
        onRetry && (
          <Button size="sm" variant="outline" onClick={onRetry}>
            <RefreshCw className="w-3.5 h-3.5" />
            Rescan
          </Button>
        )
      }
    />
  )
}

interface ErrorCopy {
  title: string
  description: string
}

function errorCopy(error: ScanError): ErrorCopy {
  switch (error.code) {
    case 'php_not_found':
      return {
        title: 'PHP not found',
        description: 'Install PHP and ensure it is available in your PATH to scan Laravel routes.',
      }
    case 'artisan_missing':
      return {
        title: 'artisan file missing',
        description: 'This folder does not contain an artisan binary. Pick a Laravel project root.',
      }
    case 'artisan_failed':
      return {
        title: 'php artisan route:list failed',
        description:
          'Spectra detected Laravel but artisan exited with an error. Run composer install and try again.',
      }
    case 'invalid_json':
      return {
        title: 'Unexpected artisan output',
        description: 'route:list returned invalid JSON. Try a clean Laravel install or check for buffered output.',
      }
    case 'no_routes':
      return {
        title: 'No routes returned',
        description: 'artisan completed but returned no routes. Define routes in routes/api.php.',
      }
    case 'not_laravel':
    case 'no_driver':
      return {
        title: 'Not a Laravel project',
        description: 'No supported framework detected in this folder.',
      }
    default:
      return {
        title: 'Scan failed',
        description: 'Spectra could not scan routes for this project.',
      }
  }
}
