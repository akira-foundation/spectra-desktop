export function RawView({ raw }: { raw: string }) {
  return (
    <pre className="h-full w-full m-0 p-3 text-[11px] font-mono whitespace-pre-wrap break-all text-foreground/85 bg-muted/20 rounded-md border border-border/40 overflow-auto">
      {raw}
    </pre>
  )
}
