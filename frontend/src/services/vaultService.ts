import { apiClient } from './api'

export interface VaultStatus {
  initialized: boolean
  sealed: boolean
  version: string
  cluster_name?: string
}

export interface VaultConfig {
  vault_addr: string
  vault_mount_path: string
  vault_token: string
  linked: string
}

export interface SecretMetadata {
  path: string
  version?: number
  created_at?: string
  updated_at?: string
  destroyed?: boolean
}

export interface Secret {
  path: string
  data: Record<string, any>
  metadata?: SecretMetadata
}

export interface VaultSecretsListResponse {
  keys: string[]
  path: string
}

export const vaultService = {
  getStatus: async (): Promise<VaultStatus> => {
    const { data } = await apiClient.get('/v1/vault/status')
    return data
  },

  getConfig: async (): Promise<VaultConfig> => {
    const { data } = await apiClient.get('/v1/vault/config')
    return data
  },

  setConfig: async (config: { vault_addr?: string; vault_token?: string; vault_mount_path?: string }): Promise<void> => {
    await apiClient.post('/v1/vault/config', config)
  },

  listSecrets: async (path: string = ''): Promise<VaultSecretsListResponse> => {
    const { data } = await apiClient.get('/v1/vault/secrets', { params: { path } })
    return data
  },

  getSecret: async (path: string): Promise<Secret> => {
    const { data } = await apiClient.get(`/v1/vault/secrets/${path}`)
    return data
  },

  writeSecret: async (path: string, secretData: Record<string, any>): Promise<Secret> => {
    const { data } = await apiClient.post(`/v1/vault/secrets/${path}`, { data: secretData })
    return data
  },

  deleteSecret: async (path: string): Promise<void> => {
    await apiClient.delete(`/v1/vault/secrets/${path}`)
  },
}
