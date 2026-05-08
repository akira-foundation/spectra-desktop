export type SchemaSource = 'form_request' | 'inline_validation' | ''
export type ConfidenceLevel = 'high' | 'medium' | 'low'

export interface InferredField {
  name: string
  type: string
  required: boolean
  rules?: string[]
  example?: unknown
}

export interface RequestSchema {
  source: SchemaSource
  confidence: ConfidenceLevel
  fields: InferredField[]
}

export function parseRequestSchema(raw?: string | null): RequestSchema | null {
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as RequestSchema
    if (!parsed.fields || !Array.isArray(parsed.fields)) return null
    return parsed
  } catch {
    return null
  }
}

export function buildExampleBody(fields: InferredField[]): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  for (const field of fields) {
    out[field.name] = field.example ?? defaultForType(field.type)
  }
  return out
}

function defaultForType(type: string): unknown {
  switch (type) {
    case 'integer':
      return 0
    case 'numeric':
      return 0
    case 'boolean':
      return false
    case 'array':
      return []
    case 'object':
      return {}
    case 'email':
      return 'user@example.com'
    case 'date':
      return '2025-01-01'
    case 'uuid':
      return '00000000-0000-4000-8000-000000000000'
    case 'url':
      return 'https://example.com'
    default:
      return ''
  }
}

export function sourceLabel(source: SchemaSource): string {
  switch (source) {
    case 'form_request':
      return 'FormRequest'
    case 'inline_validation':
      return 'Inline Validation'
    default:
      return 'Unknown'
  }
}
