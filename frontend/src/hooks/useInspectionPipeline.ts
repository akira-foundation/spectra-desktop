import { useCallback, useEffect, useRef, useState } from 'react'

export type StepStatus = 'pending' | 'active' | 'done' | 'error'

export interface InspectionStep {
  id: string
  label: string
  run: () => Promise<void>
}

export interface PipelineStepState extends InspectionStep {
  status: StepStatus
  error?: string
}

interface UsePipelineOptions {
  onComplete?: () => void
  onError?: (err: unknown) => void
}

export function useInspectionPipeline(
  steps: InspectionStep[],
  options: UsePipelineOptions = {},
) {
  const [state, setState] = useState<PipelineStepState[]>(() =>
    steps.map((s) => ({ ...s, status: 'pending' })),
  )
  const [running, setRunning] = useState(false)
  const cancelled = useRef(false)
  const optionsRef = useRef(options)
  optionsRef.current = options

  const reset = useCallback(() => {
    cancelled.current = false
    setRunning(false)
    setState(steps.map((s) => ({ ...s, status: 'pending' })))
  }, [steps])

  const run = useCallback(async () => {
    cancelled.current = false
    setRunning(true)
    setState(steps.map((s) => ({ ...s, status: 'pending' })))

    for (let i = 0; i < steps.length; i++) {
      if (cancelled.current) break
      const step = steps[i]
      setState((prev) => prev.map((s, idx) => (idx === i ? { ...s, status: 'active' } : s)))
      try {
        await step.run()
        if (cancelled.current) break
        setState((prev) => prev.map((s, idx) => (idx === i ? { ...s, status: 'done' } : s)))
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setState((prev) =>
          prev.map((s, idx) => (idx === i ? { ...s, status: 'error', error: message } : s)),
        )
        setRunning(false)
        optionsRef.current.onError?.(err)
        return
      }
    }
    setRunning(false)
    if (!cancelled.current) {
      optionsRef.current.onComplete?.()
    }
  }, [steps])

  useEffect(() => {
    return () => {
      cancelled.current = true
    }
  }, [])

  return { state, run, reset, running }
}

export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
