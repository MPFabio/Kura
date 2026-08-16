export interface RegistryImageLayer {
  instruction: string
  size_bytes: number
  created_at?: string
  // Étape n'ayant produit aucune couche (ENV, LABEL, EXPOSE…).
  empty: boolean
}

// Ce qu'un registre OCI peut restituer d'une image : sa configuration
// d'exécution et l'historique de sa construction. Les sources, dont le
// Dockerfile, n'y figurent pas et ne peuvent pas en être extraites.
export interface RegistryImageDetail {
  repository: string
  tag: string
  digest: string
  architecture?: string
  os?: string
  created_at?: string
  entrypoint?: string[]
  cmd?: string[]
  working_dir?: string
  exposed_ports?: string[]
  env?: string[]
  labels?: Record<string, string>
  layers?: RegistryImageLayer[]
  total_bytes: number
}

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

  getImage: async (repository: string, tag: string): Promise<RegistryImageDetail> => {
    try {
      const response = await apiClient.get<RegistryImageDetail>('/v1/k8s/registry/image', {
        params: { repository, tag },
      })
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de l'image ${repository}:${tag}:`, error)
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
