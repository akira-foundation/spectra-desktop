import { Github, Mail, Globe } from 'lucide-react'

interface Props {
  provider: string
  className?: string
}

const ICON_SIZE = 'h-4 w-4'

export function ProviderIcon({ provider, className }: Props) {
  const cls = className ?? `${ICON_SIZE} text-foreground/80`
  const key = provider.toLowerCase()
  switch (key) {
    case 'github':
      return <Github className={cls} />
    case 'email':
      return <Mail className={cls} />
    case 'apple':
      return <AppleSVG className={cls} />
    case 'google':
      return <GoogleSVG className={cls} />
    case 'microsoft':
      return <MicrosoftSVG className={cls} />
    case 'gitlab':
      return <GitlabSVG className={cls} />
    case 'bitbucket':
      return <BitbucketSVG className={cls} />
    default:
      return <Globe className={cls} />
  }
}

function AppleSVG({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden="true">
      <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.74.79 0 2.25-.91 3.79-.78.64.03 2.47.26 3.64 1.94-3.13 1.87-2.55 5.97 1.36 7.4-.66 1.81-1.66 3.6-2.87 4.67Zm-5-15.27c.13-2.31 1.99-3.99 4.16-4.01.27 2.4-1.69 4.45-4.16 4.01Z" />
    </svg>
  )
}

function GoogleSVG({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      />
    </svg>
  )
}

function MicrosoftSVG({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <path fill="#F25022" d="M1 1h10v10H1z" />
      <path fill="#7FBA00" d="M13 1h10v10H13z" />
      <path fill="#00A4EF" d="M1 13h10v10H1z" />
      <path fill="#FFB900" d="M13 13h10v10H13z" />
    </svg>
  )
}

function GitlabSVG({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <path
        fill="#FC6D26"
        d="M23.6 9.54l-.03-.09L20.3.74a.85.85 0 0 0-.34-.41.88.88 0 0 0-1 .05.88.88 0 0 0-.29.41L16.46 7.35H7.55L5.34.77a.85.85 0 0 0-.29-.41.88.88 0 0 0-1-.05.86.86 0 0 0-.34.41L.43 9.45l-.03.09a6.06 6.06 0 0 0 2.01 7l.01.01.03.02 4.98 3.73 2.46 1.86 1.5 1.13a1 1 0 0 0 1.22 0l1.5-1.13 2.46-1.86 5.01-3.75.01-.01a6.06 6.06 0 0 0 2.01-7z"
      />
    </svg>
  )
}

function BitbucketSVG({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="#2684FF" className={className} aria-hidden="true">
      <path d="M.78 1.21a.77.77 0 0 0-.77.9l3.26 19.8c.09.5.51.87 1.02.87H19.95a.77.77 0 0 0 .77-.65l3.27-20.02a.77.77 0 0 0-.77-.9zM14.52 15.53H9.52l-1.35-7.07h7.56z" />
    </svg>
  )
}
