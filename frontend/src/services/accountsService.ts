import {
  ListProjectAccounts,
  SaveProjectAccount,
  DeleteProjectAccount,
  SetDefaultProjectAccount,
  GenerateAccountTOTP,
  RefreshAccountToken,
  GetAccountSecrets,
} from '../../wailsjs/go/app/App'
import type { app } from '../../wailsjs/go/models'

export type AccountKind = 'bearer' | 'basic' | 'apikey' | 'oauth2' | 'login'
export type APIKeyLocation = 'header' | 'query'

export interface OAuthConfig {
  grantType: string
  tokenUrl: string
  clientId: string
  hasSecret: boolean
  scopes?: string[]
  audience?: string
  username?: string
}

export interface ProjectAccount {
  id: string
  projectID: string
  label: string
  kind: AccountKind
  scheme: string
  username: string
  hasPassword: boolean
  hasApiKey: boolean
  apiKeyHeader: string
  apiKeyIn: APIKeyLocation
  hasToken: boolean
  tokenPreview?: string
  hasRefreshToken: boolean
  expiresAt?: string
  oauth?: OAuthConfig
  hasTotp: boolean
  totpParam: string
  loginEndpointId: string
  loginBodyTemplate: string
  tokenPath: string
  user?: Record<string, unknown>
  hasCookies: boolean
  extraHeaders?: Record<string, string>
  isDefault: boolean
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface SaveOAuthInput {
  grantType: string
  tokenUrl: string
  clientId: string
  clientSecret?: string | null
  scopes?: string[]
  audience?: string
  username?: string
}

export interface SaveAccountInput {
  id?: string
  projectID: string
  label: string
  kind: AccountKind
  scheme?: string
  username?: string
  password?: string | null
  apiKey?: string | null
  apiKeyHeader?: string
  apiKeyIn?: APIKeyLocation
  token?: string | null
  refreshToken?: string | null
  oauth?: SaveOAuthInput
  totpSecret?: string | null
  totpParam?: string
  loginEndpointId?: string
  loginBodyTemplate?: string
  tokenPath?: string
  extraHeaders?: Record<string, string>
  isDefault?: boolean
  sortOrder?: number
  expiresAt?: string | null
  user?: Record<string, unknown>
}

function decode(dto: app.ProjectAccountDTO): ProjectAccount {
  return dto as unknown as ProjectAccount
}

export const accountsService = {
  async list(projectId: string): Promise<ProjectAccount[]> {
    const rows = await ListProjectAccounts(projectId)
    return (rows ?? []).map(decode)
  },
  async save(input: SaveAccountInput): Promise<ProjectAccount> {
    const saved = await SaveProjectAccount(input as unknown as app.SaveAccountInput)
    return decode(saved)
  },
  async remove(accountId: string): Promise<void> {
    await DeleteProjectAccount(accountId)
  },
  async setDefault(projectId: string, accountId: string): Promise<void> {
    await SetDefaultProjectAccount(projectId, accountId)
  },
  async totp(accountId: string): Promise<string> {
    return await GenerateAccountTOTP(accountId)
  },
  async refresh(accountId: string): Promise<ProjectAccount> {
    const refreshed = await RefreshAccountToken(accountId)
    return decode(refreshed)
  },
  async secrets(accountId: string): Promise<{
    username: string
    password?: string
    token?: string
    apiKey?: string
  } | null> {
    const result = await GetAccountSecrets(accountId)
    return (result ?? null) as
      | { username: string; password?: string; token?: string; apiKey?: string }
      | null
  },
}
