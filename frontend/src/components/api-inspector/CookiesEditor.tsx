import { Cookie as CookieIcon } from 'lucide-react'

export interface CookieRow {
  name: string
  value: string
  domain?: string
  path?: string
  enabled: boolean
}

interface Props {
  cookies?: CookieRow[]
}

export function CookiesEditor({ cookies = [] }: Props) {
  if (cookies.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-10 gap-2">
        <CookieIcon className="w-5 h-5 text-muted-foreground/50" />
        <p className="text-[11.5px] italic text-muted-foreground/70">No cookies for this request.</p>
      </div>
    )
  }
  return (
    <ul className="m-0 p-0 list-none divide-y divide-border/20">
      {cookies.map((c, i) => (
        <li key={i} className="grid grid-cols-[140px_1fr] gap-3 px-3 py-1.5 hover:bg-accent/20">
          <code className="text-[10.5px] font-mono text-foreground/85 truncate" title={c.name}>
            {c.name}
          </code>
          <code className="text-[10.5px] font-mono text-muted-foreground break-all">{c.value}</code>
        </li>
      ))}
    </ul>
  )
}
