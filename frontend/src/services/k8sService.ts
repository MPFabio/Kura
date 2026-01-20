import {
  apiClient,
  K8sNamespacesResponse,
  K8sPodsResponse,
  K8sDeploymentsResponse,
  K8sServicesResponse,
  K8sConfigMapsResponse,
  K8sSecretsResponse,
  K8sNodesResponse,
  Event,
} from './api'

export const k8sService = {
  getNamespaces: async (): Promise<K8sNamespacesResponse> => {
    try {
      const response = await apiClient.get<K8sNamespacesResponse>('/api/v1/k8s/namespaces')
      // S'assurer que la réponse a le format attendu
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des namespaces:', error)
      throw error
    }
  },

  getPods: async (namespace: string): Promise<K8sPodsResponse> => {
    try {
      if (!namespace) {
        throw new Error('Namespace requis')
      }
      const response = await apiClient.get<K8sPodsResponse>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods`
      )
      // S'assurer que la réponse a le format attendu
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des pods pour le namespace ${namespace}:`, error)
      throw error
    }
  },

  getDeployments: async (namespace: string): Promise<K8sDeploymentsResponse> => {
    try {
      if (!namespace) {
        throw new Error('Namespace requis')
      }
      const response = await apiClient.get<K8sDeploymentsResponse>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments`
      )
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des deployments pour le namespace ${namespace}:`, error)
      throw error
    }
  },

  getServices: async (namespace: string): Promise<K8sServicesResponse> => {
    try {
      if (!namespace) {
        throw new Error('Namespace requis')
      }
      const response = await apiClient.get<K8sServicesResponse>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/services`
      )
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des services pour le namespace ${namespace}:`, error)
      throw error
    }
  },

  getConfigMaps: async (namespace: string): Promise<K8sConfigMapsResponse> => {
    try {
      if (!namespace) {
        throw new Error('Namespace requis')
      }
      const response = await apiClient.get<K8sConfigMapsResponse>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/configmaps`
      )
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des ConfigMaps pour le namespace ${namespace}:`, error)
      throw error
    }
  },

  getSecrets: async (namespace: string): Promise<K8sSecretsResponse> => {
    try {
      if (!namespace) {
        throw new Error('Namespace requis')
      }
      const response = await apiClient.get<K8sSecretsResponse>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/secrets`
      )
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des secrets pour le namespace ${namespace}:`, error)
      throw error
    }
  },

  getNodes: async (): Promise<K8sNodesResponse> => {
    try {
      const response = await apiClient.get<K8sNodesResponse>('/api/v1/k8s/nodes')
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des nodes:', error)
      throw error
    }
  },

  getPodLogs: async (namespace: string, name: string, container?: string, tailLines?: number): Promise<string> => {
    try {
      const params = new URLSearchParams()
      if (container) params.append('container', container)
      if (tailLines) params.append('tail', tailLines.toString())
      const query = params.toString()
      const url = `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(name)}/logs${query ? `?${query}` : ''}`
      const response = await apiClient.get<{ logs: string }>(url)
      return response.data.logs || ''
    } catch (error) {
      console.error(`Erreur lors de la récupération des logs pour pod ${namespace}/${name}:`, error)
      throw error
    }
  },

  getPodYAML: async (namespace: string, name: string): Promise<string> => {
    try {
      const response = await apiClient.get(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(name)}/yaml`,
        { responseType: 'text' }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du YAML pour pod ${namespace}/${name}:`, error)
      throw error
    }
  },

  getDeploymentYAML: async (namespace: string, name: string): Promise<string> => {
    try {
      const response = await apiClient.get(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments/${encodeURIComponent(name)}/yaml`,
        { responseType: 'text' }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du YAML pour deployment ${namespace}/${name}:`, error)
      throw error
    }
  },

  getServiceYAML: async (namespace: string, name: string): Promise<string> => {
    try {
      const response = await apiClient.get(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(name)}/yaml`,
        { responseType: 'text' }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du YAML pour service ${namespace}/${name}:`, error)
      throw error
    }
  },

  scaleDeployment: async (namespace: string, name: string, replicas: number): Promise<void> => {
    try {
      await apiClient.put(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments/${encodeURIComponent(name)}/scale`,
        { replicas }
      )
    } catch (error) {
      console.error(`Erreur lors du scale du deployment ${namespace}/${name}:`, error)
      throw error
    }
  },

  deletePod: async (namespace: string, name: string): Promise<void> => {
    try {
      await apiClient.delete(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(name)}`
      )
    } catch (error) {
      console.error(`Erreur lors de la suppression du pod ${namespace}/${name}:`, error)
      throw error
    }
  },

  deleteDeployment: async (namespace: string, name: string): Promise<void> => {
    try {
      await apiClient.delete(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments/${encodeURIComponent(name)}`
      )
    } catch (error) {
      console.error(`Erreur lors de la suppression du deployment ${namespace}/${name}:`, error)
      throw error
    }
  },

  deleteService: async (namespace: string, name: string): Promise<void> => {
    try {
      await apiClient.delete(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(name)}`
      )
    } catch (error) {
      console.error(`Erreur lors de la suppression du service ${namespace}/${name}:`, error)
      throw error
    }
  },

  getEvents: async (namespace: string): Promise<{ items: Event[] }> => {
    try {
      const response = await apiClient.get<{ items: Event[] }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/events`
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des événements pour namespace ${namespace}:`, error)
      throw error
    }
  },

  // Actions en masse (Bulk Actions)
  bulkDeletePods: async (namespace: string, names: string[]): Promise<{ success: string[]; failed: Record<string, string>; total: number }> => {
    try {
      const response = await apiClient.post<{ success: string[]; failed: Record<string, string>; total: number }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/bulk/delete`,
        { names }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la suppression en masse des pods:`, error)
      throw error
    }
  },

  bulkRestartPods: async (namespace: string, names: string[]): Promise<{ success: string[]; failed: Record<string, string>; total: number }> => {
    try {
      const response = await apiClient.post<{ success: string[]; failed: Record<string, string>; total: number }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/bulk/restart`,
        { names }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors du redémarrage en masse des pods:`, error)
      throw error
    }
  },

  bulkDeleteDeployments: async (namespace: string, names: string[]): Promise<{ success: string[]; failed: Record<string, string>; total: number }> => {
    try {
      const response = await apiClient.post<{ success: string[]; failed: Record<string, string>; total: number }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments/bulk/delete`,
        { names }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la suppression en masse des deployments:`, error)
      throw error
    }
  },

  bulkScaleDeployments: async (namespace: string, names: string[], replicas: number): Promise<{ success: string[]; failed: Record<string, string>; total: number }> => {
    try {
      const response = await apiClient.post<{ success: string[]; failed: Record<string, string>; total: number }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/deployments/bulk/scale`,
        { names, replicas }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors du scale en masse des deployments:`, error)
      throw error
    }
  },

  bulkDeleteServices: async (namespace: string, names: string[]): Promise<{ success: string[]; failed: Record<string, string>; total: number }> => {
    try {
      const response = await apiClient.post<{ success: string[]; failed: Record<string, string>; total: number }>(
        `/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/services/bulk/delete`,
        { names }
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la suppression en masse des services:`, error)
      throw error
    }
  },
}
