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
  getStates: async (): Promise<TerraformStateResponse> => {
    try {
      const response = await apiClient.get<TerraformStateResponse>('/api/v1/terraform/states')
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

  uploadState: async (name: string, stateFile: File): Promise<TerraformState> => {
    try {
      const formData = new FormData()
      formData.append('file', stateFile)
      formData.append('name', name)
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
