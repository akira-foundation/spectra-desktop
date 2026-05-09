import { useEffect, useMemo, useState } from 'react'
import { Check, ChevronsUpDown, KeyRound, Plus, ShieldCheck, Star } from 'lucide-react'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useAccountsStore } from '@/store/accountsStore'
import { useUIStore } from '@/store/uiStore'
import { AccountKindBadge } from './AccountKindBadge'
import type { ProjectAccount } from '@/services/accountsService'
import { cn } from '@/lib/utils'

const EMPTY_ACCOUNTS: ProjectAccount[] = []

function hasUsableCreds(acc: ProjectAccount): boolean {
  if (acc.kind === 'apikey') return acc.hasApiKey
  return acc.hasToken || acc.hasPassword
}

interface Props {
  projectId: string
  tabId?: string | null
}

export function AccountSelector({ projectId, tabId }: Props) {
  const accounts = useAccountsStore((s) => s.byProject[projectId] ?? EMPTY_ACCOUNTS)
  const list = useAccountsStore((s) => s.list)
  const setActive = useAccountsStore((s) => s.setActive)
  const setActiveForTab = useAccountsStore((s) => s.setActiveForTab)
  const activeProjectId = useAccountsStore((s) => s.activeByProject[projectId])
  const activeTab = useAccountsStore((s) =>
    tabId ? s.activeByTab[tabId] : undefined,
  )
  const setCurrentPage = useUIStore((s) => s.setCurrentPage)
  const compact = useUIStore((s) => s.compactToolbar)

  const [open, setOpen] = useState(false)

  useEffect(() => {
    if (projectId) void list(projectId)
  }, [projectId, list])

  const activeId = activeTab || activeProjectId
  const active = useMemo(
    () => accounts.find((a) => a.id === activeId) ?? null,
    [accounts, activeId],
  )
  const overriding = !!tabId && !!activeTab && activeTab !== activeProjectId

  function pick(id: string | null, opts: { tabOnly?: boolean } = {}) {
    if (tabId && opts.tabOnly) {
      setActiveForTab(tabId, id)
    } else {
      setActive(projectId, id)
      if (tabId) setActiveForTab(tabId, null)
    }
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            'shrink-0 h-7 inline-flex items-center gap-1.5 px-2 rounded-md text-[11px] border transition-colors',
            overriding
              ? 'border-amber-500/60 bg-amber-500/10 text-amber-600 dark:text-amber-400'
              : 'border-border/50 hover:bg-accent/40',
            !active && 'text-muted-foreground',
          )}
          title={
            active
              ? `Active: ${active.label}${overriding ? ' (tab override)' : ''}`
              : 'No account active'
          }
        >
          <KeyRound className="w-3 h-3" />
          {active ? (
            <>
              {!compact && (
                <span className="truncate max-w-[120px]">{active.label}</span>
              )}
              {active.hasTotp && <ShieldCheck className="w-3 h-3 text-emerald-500" />}
              {!hasUsableCreds(active) && (
                <span
                  className="w-1.5 h-1.5 rounded-full bg-amber-500"
                  title="No token saved · request will be unauthenticated"
                />
              )}
            </>
          ) : (
            !compact && <span>No account</span>
          )}
          {!compact && <ChevronsUpDown className="w-3 h-3 opacity-60" />}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[320px] p-0">
        <Command>
          <CommandInput placeholder="Search accounts…" className="h-8 text-[12px]" />
          <CommandList className="max-h-72">
            <CommandEmpty className="py-3 text-center text-[11.5px] text-muted-foreground">
              No accounts
            </CommandEmpty>
            <CommandGroup heading={tabId ? 'For this tab' : 'For all tabs'}>
              {accounts.map((acc) => (
                <CommandItem
                  key={acc.id}
                  value={`${acc.label} ${acc.kind}`}
                  onSelect={() => pick(acc.id)}
                  className="gap-2 text-[11.5px]"
                >
                  <span
                    className="text-muted-foreground"
                    title={acc.isDefault ? 'Default' : ''}
                  >
                    <Star
                      className="w-3 h-3"
                      fill={acc.isDefault ? 'currentColor' : 'none'}
                    />
                  </span>
                  <span className="flex-1 truncate font-medium">{acc.label}</span>
                  <AccountKindBadge kind={acc.kind} />
                  {activeId === acc.id && <Check className="w-3 h-3 text-primary" />}
                </CommandItem>
              ))}
            </CommandGroup>
            {tabId && (
              <CommandGroup heading="Tab override">
                {accounts.map((acc) => (
                  <CommandItem
                    key={`tab-${acc.id}`}
                    value={`tab ${acc.label} ${acc.kind}`}
                    onSelect={() => pick(acc.id, { tabOnly: true })}
                    className="gap-2 text-[11.5px]"
                  >
                    <span className="w-3" />
                    <span className="flex-1 truncate">Use {acc.label} only here</span>
                    {activeTab === acc.id && <Check className="w-3 h-3 text-primary" />}
                  </CommandItem>
                ))}
                {activeTab && (
                  <CommandItem
                    value="clear-tab-override"
                    onSelect={() => pick(null, { tabOnly: true })}
                    className="text-[11.5px] italic text-muted-foreground"
                  >
                    Clear tab override
                  </CommandItem>
                )}
              </CommandGroup>
            )}
            <CommandGroup>
              <CommandItem
                value="manage-accounts"
                onSelect={() => {
                  setOpen(false)
                  setCurrentPage('accounts')
                }}
                className="gap-2 text-[11.5px] text-primary"
              >
                <Plus className="w-3 h-3" />
                Manage accounts…
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
