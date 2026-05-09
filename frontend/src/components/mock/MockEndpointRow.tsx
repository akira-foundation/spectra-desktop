import { useState } from 'react'
import { ChevronDown, ChevronRight, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useHttpMethod } from '@/hooks/useHttpMethod'
import type { MockOverride, MockSource } from '@/services/mockService'
import type { ScannedEndpoint } from '@/services/scannerService'
import { cn } from '@/lib/utils'

interface Props {
  endpoint: ScannedEndpoint
  override: MockOverride | undefined
  onSave: (input: SavePayload) => Promise<void>
}

interface SavePayload {
  endpointId: string
  enabled: boolean
  status: number
  latencyMs: number
  source: MockSource
  body: string
}

export function MockEndpointRow({ endpoint, override, onSave }: Props) {
  const { getMethodColor } = useHttpMethod()
  const [open, setOpen] = useState(false)
  const [enabled, setEnabled] = useState(override?.enabled ?? true)
  const [status, setStatus] = useState(String(override?.status || 200))
  const [latency, setLatency] = useState(String(override?.latencyMs || 0))
  const [source, setSource] = useState<MockSource>((override?.source as MockSource) || 'auto')
  const [body, setBody] = useState(override?.body || '')
  const [saving, setSaving] = useState(false)

  const dirty =
    enabled !== (override?.enabled ?? true) ||
    parseInt(status, 10) !== (override?.status || 200) ||
    parseInt(latency, 10) !== (override?.latencyMs || 0) ||
    source !== ((override?.source as MockSource) || 'auto') ||
    body !== (override?.body || '')

  const save = async () => {
    setSaving(true)
    try {
      await onSave({
        endpointId: endpoint.id,
        enabled,
        status: parseInt(status, 10) || 200,
        latencyMs: parseInt(latency, 10) || 0,
        source,
        body,
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <li className="border-b border-border/40 last:border-b-0">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="w-full px-3.5 py-2 flex items-center gap-2.5 hover:bg-accent/20 text-left"
      >
        {open ? (
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        )}
        <span
          className={cn(
            'inline-flex shrink-0 justify-center text-[9px] font-bold tracking-wider rounded px-1.5 py-px w-12',
            getMethodColor(endpoint.method),
          )}
        >
          {endpoint.method}
        </span>
        <span className="font-mono text-[11.5px] truncate flex-1 min-w-0">
          {endpoint.path}
        </span>
        {!enabled && (
          <span className="text-[10px] text-amber-500 uppercase tracking-wider shrink-0">
            disabled
          </span>
        )}
        {override && (
          <span className="text-[10px] text-purple-500 uppercase tracking-wider shrink-0">
            override
          </span>
        )}
      </button>
      {open && (
        <div className="px-10 py-3 border-t border-border/40 bg-muted/20 space-y-3">
          <div className="grid grid-cols-4 gap-2">
            <Field label="Enabled">
              <button
                type="button"
                onClick={() => setEnabled((v) => !v)}
                className={cn(
                  'h-7 px-2 rounded-md border text-[11px] font-medium transition-colors',
                  enabled
                    ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                    : 'border-border/50 bg-muted/40 text-muted-foreground',
                )}
              >
                {enabled ? 'On' : 'Off'}
              </button>
            </Field>
            <Field label="Status">
              <Input
                value={status}
                onChange={(e) => setStatus(e.target.value.replace(/[^0-9]/g, ''))}
                className="h-7 text-[12px] font-mono"
              />
            </Field>
            <Field label="Latency (ms)">
              <Input
                value={latency}
                onChange={(e) => setLatency(e.target.value.replace(/[^0-9]/g, ''))}
                className="h-7 text-[12px] font-mono"
              />
            </Field>
            <Field label="Source">
              <Select value={source} onValueChange={(v) => setSource(v as MockSource)}>
                <SelectTrigger className="h-7 text-[12px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="auto">Auto (history → generated)</SelectItem>
                  <SelectItem value="history">History only</SelectItem>
                  <SelectItem value="generated">Generated only</SelectItem>
                  <SelectItem value="custom">Custom body</SelectItem>
                </SelectContent>
              </Select>
            </Field>
          </div>
          {source === 'custom' && (
            <Field label="Response body">
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder='{"data": {...}}'
                className="w-full h-32 rounded-md border border-border/60 bg-input/60 px-2.5 py-1.5 text-[12px] font-mono resize-y outline-none"
              />
            </Field>
          )}
          <div className="flex justify-end">
            <Button size="sm" onClick={save} disabled={saving || !dirty}>
              <Save className="h-3.5 w-3.5" />
              {saving ? 'Saving…' : 'Save'}
            </Button>
          </div>
        </div>
      )}
    </li>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1">
      <label className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}
