import { apiClient } from './api'

export interface RegistryRepository {
  name: string
  tag_count: number
}

export interface RegistryTag {
  name: string
  digest: string
  media_type: string
  size_bytes: number
  signed: boolean
  type: 'image' | 'helm-chart' | string
}

export interface RegistryRepositoryDetail {
  name: string
  tags: RegistryTag[]
}

export const registryService = {
  listRepositories: async (): Promise<RegistryRepository[]> => {
    try {
      const response = await apiClient.get<{ items: RegistryRepository[] }>('/v1/k8s/registry/repositories')
      return response.data?.items ?? []
    } catch (error) {
      console.error('Erreur lors de la récupération des dépôts du registre:', error)
      throw error
    }
  },

  getRepository: async (name: string): Promise<RegistryRepositoryDetail> => {
    try {
      const response = await apiClient.get<RegistryRepositoryDetail>(
        `/v1/k8s/registry/repositories/${name}`
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du dépôt ${name}:`, error)
      throw error
    }
  },
}
