import { ListScratchRequests, SaveScratchRequest, DeleteScratchRequest } from '../../wailsjs/go/app/App'
import { app } from '../../wailsjs/go/models'
import type { HeaderRow } from '@/components/api-inspector/HeadersEditor'
import type { TimelineData } from '@/components/api-inspector/response'

export interface ScratchResponse {
  status: number
  body: string
  headers: Record<string, string[]>
  durationMs: number
  timeline?: TimelineData | null
}

export interface ScratchRequest {
  id: string
  projectID: string
  name: string
  method: string
  url: string
  headers: HeaderRow[]
  body: string
  response?: ScratchResponse | null
  sortOrder: number
}

function decode(dto: app.ScratchRequestDTO): ScratchRequest {
  let headers: HeaderRow[] = []
  try {
    const parsed = dto.headersJson ? JSON.parse(dto.headersJson) : []
    if (Array.isArray(parsed)) headers = parsed
  } catch {}
  let response: ScratchResponse | null = null
  if (dto.responseJson) {
    try {
      response = JSON.parse(dto.responseJson)
    } catch {}
  }
  return {
    id: dto.id,
    projectID: dto.projectID,
    name: dto.name,
    method: dto.method,
    url: dto.url,
    headers,
    body: dto.body,
    response,
    sortOrder: dto.sortOrder ?? 0,
  }
}

function encode(req: ScratchRequest): app.ScratchRequestDTO {
  return app.ScratchRequestDTO.createFrom({
    id: req.id,
    projectID: req.projectID,
    name: req.name,
    method: req.method,
    url: req.url,
    headersJson: JSON.stringify(req.headers ?? []),
    body: req.body ?? '',
    responseJson: req.response ? JSON.stringify(req.response) : '',
    sortOrder: req.sortOrder ?? 0,
  })
}

export const scratchService = {
  async list(projectID: string): Promise<ScratchRequest[]> {
    const rows = await ListScratchRequests(projectID)
    return rows.map(decode)
  },
  async save(req: ScratchRequest): Promise<ScratchRequest> {
    const saved = await SaveScratchRequest(encode(req))
    return decode(saved)
  },
  async remove(id: string): Promise<void> {
    await DeleteScratchRequest(id)
  },
}
