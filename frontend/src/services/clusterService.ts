import { apiClient } from './api'

export interface KubernetesCluster {
  id: string
  name: string
  description?: string
  endpoint?: string
  kubeconfig: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ClusterStatus {
  cluster_id: string
  connected: boolean
  version?: string
  nodes_count?: number
  last_checked?: string
  error?: string
}

export interface ClusterResponse {
  items: KubernetesCluster[]
}

export const clusterService = {
  getClusters: async (): Promise<ClusterResponse> => {
    try {
      const response = await apiClient.get<ClusterResponse>('/api/v1/k8s/clusters')
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des clusters:', error)
      throw error
    }
  },

  getCluster: async (id: string): Promise<KubernetesCluster> => {
    try {
      const response = await apiClient.get<KubernetesCluster>(`/api/v1/k8s/clusters/${id}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du cluster ${id}:`, error)
      throw error
    }
  },

  getActiveCluster: async (): Promise<KubernetesCluster | null> => {
    try {
      const response = await apiClient.get<KubernetesCluster>('/api/v1/k8s/clusters/active')
      return response.data
    } catch (error: any) {
      if (error.response?.status === 404) {
        return null
      }
      console.error('Erreur lors de la récupération du cluster actif:', error)
      throw error
    }
  },

  createCluster: async (cluster: {
    name: string
    description?: string
    endpoint?: string
    kubeconfig: string
    is_active?: boolean
  }): Promise<KubernetesCluster> => {
    try {
      const response = await apiClient.post<KubernetesCluster>('/api/v1/k8s/clusters', cluster)
      return response.data
    } catch (error) {
      console.error('Erreur lors de la création du cluster:', error)
      throw error
    }
  },

  updateCluster: async (
    id: string,
    cluster: {
      name?: string
      description?: string
      endpoint?: string
      kubeconfig?: string
      is_active?: boolean
    }
  ): Promise<KubernetesCluster> => {
    try {
      const response = await apiClient.put<KubernetesCluster>(`/api/v1/k8s/clusters/${id}`, cluster)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la mise à jour du cluster ${id}:`, error)
      throw error
    }
  },

  deleteCluster: async (id: string): Promise<void> => {
    try {
      await apiClient.delete(`/api/v1/k8s/clusters/${id}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression du cluster ${id}:`, error)
      throw error
    }
  },

  setActiveCluster: async (id: string): Promise<void> => {
    try {
      await apiClient.post(`/api/v1/k8s/clusters/${id}/activate`)
    } catch (error) {
      console.error(`Erreur lors de l'activation du cluster ${id}:`, error)
      throw error
    }
  },

  testClusterConnection: async (id: string): Promise<ClusterStatus> => {
    try {
      const response = await apiClient.get<ClusterStatus>(`/api/v1/k8s/clusters/${id}/test`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors du test de connexion du cluster ${id}:`, error)
      throw error
    }
  },
}
