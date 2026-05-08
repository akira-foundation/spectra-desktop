import { useCallback, useState } from 'react'
import { runnerService, type RunnerInput, type RunnerResponse } from '@/services/runnerService'

export interface RunnerError {
  code?: string
  message: string
}

export interface RunnerState {
  loading: boolean
  response: RunnerResponse | null
  error: RunnerError | null
  execute: (input: RunnerInput) => Promise<void>
  reset: () => void
}

export function useRequestRunner(): RunnerState {
  const [loading, setLoading] = useState(false)
  const [response, setResponse] = useState<RunnerResponse | null>(null)
  const [error, setError] = useState<RunnerError | null>(null)

  const execute = useCallback(async (input: RunnerInput) => {
    setLoading(true)
    setError(null)
    try {
      const result = await runnerService.execute(input)
      setResponse(result)
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      setError({ code: classify(message), message })
      setResponse(null)
    } finally {
      setLoading(false)
    }
  }, [])

  const reset = useCallback(() => {
    setLoading(false)
    setResponse(null)
    setError(null)
  }, [])

  return { loading, response, error, execute, reset }
}

function classify(message: string): string | undefined {
  const lower = message.toLowerCase()
  if (lower.includes('connection refused')) return 'connection_refused'
  if (lower.includes('timed out') || lower.includes('timeout')) return 'timeout'
  if (lower.includes('dns') || lower.includes('no such host')) return 'dns'
  if (lower.includes('invalid url')) return 'invalid_url'
  if (lower.includes('tls')) return 'tls'
  if (lower.includes('missing base url')) return 'missing_base_url'
  return undefined
}
