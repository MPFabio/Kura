import axios from 'axios'
import { apiClient } from './api'

// Contournement : si Kong ne route pas le pipeline, utiliser l'URL directe (VITE_PIPELINE_URL=http://localhost:8084)
const pipelineBaseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_PIPELINE_URL
    ? String(import.meta.env.VITE_PIPELINE_URL).replace(/\/$/, '')
    : null

const pipelineClient = pipelineBaseURL
  ? axios.create({
      baseURL: pipelineBaseURL,
      headers: { 'Content-Type': 'application/json' },
    })
  : apiClient

const getClient = () => pipelineClient

export type PipelineProvider = 'github' | 'gitlab' | 'jenkins'

export type PipelineRunStatus =
  | 'pending'
  | 'running'
  | 'success'
  | 'failure'
  | 'cancelled'
  | 'skipped'

export interface PipelineRun {
  id: string
  provider: PipelineProvider
  repository: string
  branch: string
  commit_sha?: string
  commit_msg?: string
  author?: string
  status: PipelineRunStatus
  external_id?: string
  external_url?: string
  workflow_name?: string
  duration_ms?: number
  started_at?: string
  finished_at?: string
  created_at: string
}

export interface PipelineRunsResponse {
  runs: PipelineRun[]
  count: number
}

export interface AggregatedStatus {
  repository: string
  branch: string
  provider: PipelineProvider
  last_status: PipelineRunStatus
  last_run_id: string
  last_run_at?: string
  success_count: number
  failure_count: number
  total_runs: number
}

export interface PipelineProviderInfo {
  id: string
  name: string
}

export interface PipelineProvidersResponse {
  providers: PipelineProviderInfo[]
}

export interface PipelineConfig {
  github_repos: string[]
  linked: boolean
}

// Intercepteur auth pour pipelineClient (copie le token si on utilise l'URL directe)
if (pipelineBaseURL) {
  pipelineClient.interceptors.request.use((config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })
}

export const pipelineService = {
  getRuns: async (params?: {
    provider?: string
    repository?: string
    branch?: string
    limit?: number
  }): Promise<PipelineRunsResponse> => {
    const searchParams = new URLSearchParams()
    if (params?.provider) searchParams.set('provider', params.provider)
    if (params?.repository) searchParams.set('repository', params.repository)
    if (params?.branch) searchParams.set('branch', params.branch)
    if (params?.limit) searchParams.set('limit', String(params.limit))
    const qs = searchParams.toString()
    const url = `/api/v1/pipeline/runs${qs ? `?${qs}` : ''}`
    const response = await getClient().get<PipelineRunsResponse>(url)
    return response.data
  },

  getRun: async (id: string): Promise<PipelineRun> => {
    const response = await getClient().get<PipelineRun>(`/api/v1/pipeline/runs/${id}`)
    return response.data
  },

  getAggregatedStatus: async (
    provider: string,
    repository: string,
    branch?: string
  ): Promise<AggregatedStatus | { message: string }> => {
    const params = new URLSearchParams({ provider, repository })
    if (branch) params.set('branch', branch)
    const response = await getClient().get<AggregatedStatus | { message: string }>(
      `/api/v1/pipeline/aggregated?${params}`
    )
    return response.data
  },

  getProviders: async (): Promise<PipelineProvidersResponse> => {
    const response = await getClient().get<PipelineProvidersResponse>(
      '/api/v1/pipeline/providers'
    )
    return response.data
  },

  getConfig: async (): Promise<PipelineConfig> => {
    const response = await getClient().get<PipelineConfig>('/api/v1/pipeline/config')
    return response.data
  },

  setConfig: async (data: {
    github_token?: string
    github_repos?: string[]
  }): Promise<{ message: string; config: PipelineConfig }> => {
    const response = await getClient().post<{ message: string; config: PipelineConfig }>(
      '/api/v1/pipeline/config',
      data
    )
    return response.data
  },

  /** Déclenche une sync manuelle depuis l'API GitHub */
  sync: async (): Promise<{ message: string; runs: number }> => {
    const response = await getClient().post<{ message: string; runs: number }>(
      '/api/v1/pipeline/sync'
    )
    return response.data
  },
}
