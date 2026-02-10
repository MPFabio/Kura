import { apiClient } from './api'

export interface TerraformState {
  id: string
  name: string
  state: {
    version: number
    terraform_version?: string
    serial: number
    lineage?: string
    outputs?: Record<string, any>
    resources?: TerraformResource[]
  }
  uploaded_at: string
  last_checked?: string
}

export interface TerraformResource {
  module?: string
  mode: string
  type: string
  name: string
  provider: string
  instances?: TerraformResourceInstance[]
}

export interface TerraformResourceInstance {
  schema_version?: number
  attributes: Record<string, any>
  sensitive_attributes?: any[]
  dependencies?: any[]
}

export interface TerraformStateSummary {
  resource_count: number
  output_count: number
  last_modified?: string
  drift_count: number
}

export interface TerraformStateResponse {
  items: TerraformState[]
}

export interface TerraformDriftResult {
  resource_address: string
  resource_type: string
  status: string // "in_sync", "drifted", "missing", "unknown"
  differences?: Array<{
    attribute: string
    expected: any
    actual: any
    change_type: string
  }>
  detected_at: string
  message?: string
}

export interface TerraformDriftResponse {
  items: TerraformDriftResult[]
}

export const terraformService = {
  getStates: async (projectId: string): Promise<TerraformStateResponse> => {
    try {
      const response = await apiClient.get<TerraformStateResponse>(`/api/v1/terraform/states?project_id=${projectId}`)
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des états Terraform:', error)
      throw error
    }
  },

  getState: async (id: string): Promise<TerraformState> => {
    try {
      const response = await apiClient.get<TerraformState>(`/api/v1/terraform/states/${id}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de l'état Terraform ${id}:`, error)
      throw error
    }
  },

  getStateSummary: async (id: string): Promise<TerraformStateSummary> => {
    try {
      const response = await apiClient.get<TerraformStateSummary>(`/api/v1/terraform/states/${id}/summary`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du résumé de l'état ${id}:`, error)
      throw error
    }
  },

  getResources: async (id: string): Promise<{ items: TerraformResource[] }> => {
    try {
      const response = await apiClient.get<{ items: TerraformResource[] }>(`/api/v1/terraform/states/${id}/resources`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des ressources pour l'état ${id}:`, error)
      throw error
    }
  },

  getOutputs: async (id: string): Promise<Record<string, any>> => {
    try {
      const response = await apiClient.get<Record<string, any>>(`/api/v1/terraform/states/${id}/outputs`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des sorties pour l'état ${id}:`, error)
      throw error
    }
  },

  detectDrift: async (id: string): Promise<TerraformDriftResponse> => {
    try {
      const response = await apiClient.post<TerraformDriftResponse>(`/api/v1/terraform/states/${id}/drift`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la détection de drift pour l'état ${id}:`, error)
      throw error
    }
  },

  uploadState: async (name: string, stateFile: File, projectId?: string): Promise<TerraformState> => {
    try {
      const formData = new FormData()
      formData.append('file', stateFile)
      formData.append('name', name)
      if (projectId) {
        formData.append('project_id', projectId)
      }
      const response = await apiClient.post<TerraformState>('/api/v1/terraform/states/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      return response.data
    } catch (error) {
      console.error('Erreur lors de l\'upload de l\'état Terraform:', error)
      throw error
    }
  },

  uploadStateJSON: async (name: string, state: any): Promise<TerraformState> => {
    try {
      const response = await apiClient.post<TerraformState>('/api/v1/terraform/states', {
        name,
        state,
      })
      return response.data
    } catch (error) {
      console.error('Erreur lors de l\'upload de l\'état Terraform (JSON):', error)
      throw error
    }
  },

  deleteState: async (id: string): Promise<void> => {
    try {
      await apiClient.delete(`/api/v1/terraform/states/${id}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression de l'état ${id}:`, error)
      throw error
    }
  },
}

export interface TerraformSource {
  id: string
  state_file_id: string
  type: 's3' | 'azure' | 'gcp' | 'local' | 'terraform_cloud'
  config: {
    // S3
    s3_bucket?: string
    s3_key?: string
    s3_region?: string
    s3_endpoint?: string
    aws_access_key_id?: string
    aws_secret_access_key?: string
    // Azure
    azure_account_name?: string
    azure_account_key?: string
    azure_connection_string?: string
    azure_container?: string
    azure_blob_name?: string
    // GCP
    gcp_bucket?: string
    gcp_object_name?: string
    gcp_credentials_json?: string
    // Terraform Cloud
    terraform_cloud_org?: string
    terraform_cloud_workspace?: string
    terraform_cloud_token?: string
    // Synchronisation
    sync_interval?: string
    auto_sync: boolean
  }
  enabled: boolean
  last_sync?: string
  next_sync?: string
  created_at: string
  updated_at: string
}

export interface TerraformSourceResponse {
  items: TerraformSource[]
}

export interface TerraformSyncJob {
  id: string
  state_file_id: string
  source_id: string
  status: 'pending' | 'running' | 'success' | 'failed'
  started_at?: string
  completed_at?: string
  error?: string
  message?: string
}

export const terraformSourceService = {
  getSources: async (): Promise<TerraformSourceResponse> => {
    try {
      const response = await apiClient.get<TerraformSourceResponse>('/api/v1/terraform/sources')
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des sources:', error)
      throw error
    }
  },

  getSource: async (id: string): Promise<TerraformSource> => {
    try {
      const response = await apiClient.get<TerraformSource>(`/api/v1/terraform/sources/${id}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de la source ${id}:`, error)
      throw error
    }
  },

  addSource: async (source: Partial<TerraformSource>): Promise<TerraformSource> => {
    try {
      const response = await apiClient.post<TerraformSource>('/api/v1/terraform/sources', source)
      return response.data
    } catch (error) {
      console.error('Erreur lors de l\'ajout de la source:', error)
      throw error
    }
  },

  syncSource: async (sourceId: string): Promise<TerraformSyncJob> => {
    try {
      const response = await apiClient.post<TerraformSyncJob>(`/api/v1/terraform/sources/${sourceId}/sync`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la synchronisation de la source ${sourceId}:`, error)
      throw error
    }
  },

  updateSource: async (sourceId: string, source: Partial<TerraformSource>): Promise<TerraformSource> => {
    try {
      const response = await apiClient.put<TerraformSource>(`/api/v1/terraform/sources/${sourceId}`, source)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la mise à jour de la source ${sourceId}:`, error)
      throw error
    }
  },

  deleteSource: async (sourceId: string): Promise<void> => {
    try {
      await apiClient.delete(`/api/v1/terraform/sources/${sourceId}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression de la source ${sourceId}:`, error)
      throw error
    }
  },

  testS3Connection: async (config: {
    bucket: string
    region: string
    endpoint?: string
    aws_access_key_id?: string
    aws_secret_access_key?: string
  }): Promise<{ message: string }> => {
    try {
      const response = await apiClient.post<{ message: string }>('/api/v1/terraform/sources/test-s3', config)
      return response.data
    } catch (error) {
      console.error('Erreur lors du test de connexion S3:', error)
      throw error
    }
  },
}
