import { useEffect, useState } from 'react'
import { Code2, FileText, Shield, KeyRound, FileCheck, Sparkles, LogIn } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { sourceLabel, type RequestSchema } from '@/lib/request-schema'
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import { cn } from '@/lib/utils'
import { SetEndpointAuth } from '../../../wailsjs/go/app/App'
import { app } from '../../../wailsjs/go/models'

interface EndpointInfoSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  endpointId?: string
  method?: string
  path?: string
  controller?: string
  sourceFile?: string
  sourceLine?: number
  middleware?: string[]
  authRequired?: boolean
  schema?: RequestSchema | null
  authRole?: string
  authHint?: string
  authRoleOverride?: string
  tokenPathOverride?: string
  onAuthOverrideSaved?: () => void
}

export function EndpointInfoSheet({
  open,
  onOpenChange,
  endpointId,
  method,
  path,
  controller,
  sourceFile,
  sourceLine,
  middleware,
  authRequired,
  schema,
  authRole,
  authHint,
  authRoleOverride,
  tokenPathOverride,
  onAuthOverrideSaved,
}: EndpointInfoSheetProps) {
  const { getMethodColor } = useHttpMethod()
  const [roleDraft, setRoleDraft] = useState(authRoleOverride ?? authRole ?? '')
  const [pathDraft, setPathDraft] = useState(tokenPathOverride ?? '')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setRoleDraft(authRoleOverride ?? authRole ?? '')
    setPathDraft(tokenPathOverride ?? '')
  }, [endpointId, authRole, authRoleOverride, tokenPathOverride])

  const handleSaveAuth = async () => {
    if (!endpointId) return
    setSaving(true)
    try {
      await SetEndpointAuth(
        app.SetEndpointAuthInput.createFrom({
          endpointID: endpointId,
          role: roleDraft,
          tokenPath: pathDraft.trim(),
        }),
      )
      onAuthOverrideSaved?.()
    } finally {
      setSaving(false)
    }
  }

  const effectiveRole = authRoleOverride && authRoleOverride !== '' ? authRoleOverride : authRole

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent className="data-[vaul-drawer-direction=right]:sm:max-w-lg !bg-sidebar">
        <DrawerHeader className="border-b border-border/60 space-y-2">
          <DrawerTitle className="text-[13px] font-semibold tracking-tight">
            Endpoint info
          </DrawerTitle>
          {method && path && (
            <div className="flex items-center gap-2 min-w-0">
              <span
                className={cn(
                  'inline-flex w-12 shrink-0 justify-center text-[10px] font-bold tracking-wider rounded px-1 py-0.5',
                  getMethodColor(method),
                )}
              >
                {method}
              </span>
              <code className="font-mono text-[12px] text-foreground/90 truncate">{path}</code>
            </div>
          )}
        </DrawerHeader>

        <div className="flex-1 overflow-auto">
          <dl className="divide-y divide-border/60">
            {controller && (
              <Row icon={Code2} label="Controller">
                <code
                  title={controller}
                  className="font-mono text-[12px] break-all text-foreground/90 leading-relaxed"
                >
                  {shortClass(controller)}
                </code>
              </Row>
            )}

            {sourceFile && (
              <Row icon={FileText} label="Source">
                <code className="font-mono text-[11.5px] break-all text-foreground/90 leading-relaxed">
                  {sourceLine !== undefined ? `${sourceFile}:${sourceLine}` : sourceFile}
                </code>
              </Row>
            )}

            {middleware && middleware.length > 0 && (
              <Row icon={Shield} label="Middleware" count={middleware.length}>
                <ul className="space-y-1">
                  {middleware.map((m) => (
                    <li
                      key={m}
                      title={m}
                      className="font-mono text-[12px] text-foreground/85 break-all leading-relaxed"
                    >
                      <span className="text-muted-foreground/60 mr-1.5">›</span>
                      {shortClass(m)}
                    </li>
                  ))}
                </ul>
              </Row>
            )}

            {authRequired !== undefined && (
              <Row icon={KeyRound} label="Auth">
                <span
                  className={cn(
                    'inline-flex items-center gap-1.5 text-[12px] font-medium',
                    authRequired ? 'text-foreground' : 'text-muted-foreground',
                  )}
                >
                  <span
                    className={cn(
                      'w-1.5 h-1.5 rounded-full',
                      authRequired ? 'bg-emerald-500' : 'bg-muted-foreground',
                    )}
                  />
                  {authRequired ? 'Required' : 'Public'}
                </span>
              </Row>
            )}

            {endpointId && (
              <Row icon={LogIn} label="Auth role">
                <div className="space-y-2">
                  {effectiveRole && (
                    <div className="flex items-center gap-2 text-[12px]">
                      <span
                        className={cn(
                          'inline-flex items-center text-[10px] font-medium px-1.5 py-0.5 rounded border capitalize',
                          authRoleOverride
                            ? 'border-primary/30 bg-primary/10 text-primary'
                            : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-500',
                        )}
                      >
                        {effectiveRole}
                      </span>
                      {authRoleOverride && (
                        <span className="text-[10.5px] text-muted-foreground">manual override</span>
                      )}
                    </div>
                  )}
                  {authHint && (
                    <p className="text-[10.5px] text-muted-foreground leading-relaxed">{authHint}</p>
                  )}
                  <div className="space-y-1.5">
                    <label className="text-[10.5px] font-medium text-muted-foreground">
                      Mark as
                    </label>
                    <select
                      value={roleDraft}
                      onChange={(e) => setRoleDraft(e.target.value)}
                      className="w-full h-7 text-[11.5px] bg-muted/40 border border-border/40 rounded-md px-2"
                    >
                      <option value="">Auto ({authRole || 'none'})</option>
                      <option value="login">Login</option>
                      <option value="logout">Logout</option>
                      <option value="refresh">Refresh</option>
                      <option value="csrf">CSRF</option>
                      <option value="none">Not auth</option>
                    </select>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10.5px] font-medium text-muted-foreground">
                      Token JSONPath (override)
                    </label>
                    <Input
                      value={pathDraft}
                      onChange={(e) => setPathDraft(e.target.value)}
                      placeholder="data.token"
                      className="h-7 text-[11.5px] font-mono"
                    />
                  </div>
                  <Button
                    size="sm"
                    onClick={handleSaveAuth}
                    disabled={saving}
                    className="w-full h-7 text-[11px]"
                  >
                    {saving ? 'Saving...' : 'Save auth config'}
                  </Button>
                </div>
              </Row>
            )}

            {schema && schema.fields.length > 0 && (
              <Row
                icon={schema.confidence === 'high' ? FileCheck : Sparkles}
                label="Validation"
                count={schema.fields.length}
              >
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <span
                      className={cn(
                        'inline-flex items-center text-[10px] font-medium px-1.5 py-0.5 rounded border',
                        schema.confidence === 'high'
                          ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-500'
                          : 'border-amber-500/30 bg-amber-500/10 text-amber-500',
                      )}
                    >
                      {sourceLabel(schema.source)}
                    </span>
                    <span className="text-[10.5px] text-muted-foreground capitalize">
                      {schema.confidence} confidence
                    </span>
                  </div>
                  <ul className="space-y-1">
                    {schema.fields.map((f) => (
                      <li
                        key={f.name}
                        className="flex items-center gap-2 text-[11.5px] font-mono"
                      >
                        <span className="text-foreground/85 truncate">{f.name}</span>
                        {f.required && (
                          <span className="text-destructive text-[10px]">*</span>
                        )}
                        <span className="ml-auto text-[10px] text-muted-foreground">
                          {f.type}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              </Row>
            )}
          </dl>
        </div>
      </DrawerContent>
    </Drawer>
  )
}

interface RowProps {
  icon: LucideIcon
  label: string
  count?: number
  children: React.ReactNode
}

function shortClass(value: string): string {
  if (!value) return value
  const [head, tail] = value.includes('@') ? value.split('@', 2) : [value, '']
  const segments = head.split('\\').filter(Boolean)
  const last = segments[segments.length - 1] ?? head
  return tail ? `${last}@${tail}` : last
}

function Row({ icon: Icon, label, count, children }: RowProps) {
  return (
    <div className="px-4 py-3 space-y-1.5">
      <dt className="flex items-center gap-1.5">
        <Icon className="w-3.5 h-3.5 text-muted-foreground" />
        <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
        {count !== undefined && (
          <span className="text-[10px] font-mono text-muted-foreground/70">{count}</span>
        )}
      </dt>
      <dd className="pl-5">{children}</dd>
    </div>
  )
}
