import { apiClient } from './api'

export interface ArgoApplication {
  name: string
  namespace: string
  project: string
  sync_status: string
  health_status: string
  repo_url: string
  path: string
  target_revision: string
  dest_namespace: string
  dest_server: string
}

export interface ArgoHistoryEntry {
  id: number
  revision_id: string
  deployed_at: string
  source: string
}

export interface ArgoApplicationDetail extends ArgoApplication {
  history: ArgoHistoryEntry[]
}

export interface CreateApplicationRequest {
  name: string
  project?: string
  source_type?: 'git' | 'helm'
  repo_url: string
  path: string
  chart?: string
  helm_values?: string
  target_revision?: string
  dest_namespace: string
  dest_server?: string
  sync_policy_automated: boolean
  prune: boolean
  self_heal: boolean
  branch: string
  create_branch_from?: string
}

export interface GitOpsInfo {
  clone_url: string
  repository: string
  branches: string[]
}

export interface ArgoCDStatus {
  installed: boolean
  server_ready: boolean
  self_managed: boolean
  version?: string
}

export interface HelmChartSummary {
  package_id: string
  name: string
  display_name: string
  description: string
  version: string
  logo_url: string
  repo_url: string
  repo_name: string
  official: boolean
  cncf: boolean
  stars: number
  home_url: string
}

export const argocdService = {
  installArgoCD: async (branch: string, createBranchFrom?: string): Promise<void> => {
    try {
      await apiClient.post('/v1/k8s/argocd/install', {
        branch,
        create_branch_from: createBranchFrom || undefined,
      })
    } catch (error) {
      console.error("Erreur lors de l'installation d'ArgoCD:", error)
      throw error
    }
  },

  getGitOpsInfo: async (): Promise<GitOpsInfo> => {
    try {
      const response = await apiClient.get<GitOpsInfo>('/v1/k8s/argocd/gitops/branches')
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des branches GitOps:', error)
      throw error
    }
  },

  getStatus: async (): Promise<ArgoCDStatus> => {
    try {
      const response = await apiClient.get<ArgoCDStatus>('/v1/k8s/argocd/status')
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération du statut ArgoCD:', error)
      throw error
    }
  },

  listApplications: async (): Promise<ArgoApplication[]> => {
    try {
      const response = await apiClient.get<{ items: ArgoApplication[] }>('/v1/k8s/argocd/applications')
      return response.data?.items ?? []
    } catch (error) {
      console.error('Erreur lors de la récupération des Applications ArgoCD:', error)
      throw error
    }
  },

  getApplication: async (name: string): Promise<ArgoApplicationDetail> => {
    try {
      const response = await apiClient.get<ArgoApplicationDetail>(
        `/v1/k8s/argocd/applications/${encodeURIComponent(name)}`
      )
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération de l'Application ${name}:`, error)
      throw error
    }
  },

  createApplication: async (req: CreateApplicationRequest): Promise<ArgoApplication> => {
    try {
      const response = await apiClient.post<ArgoApplication>('/v1/k8s/argocd/applications', req)
      return response.data
    } catch (error) {
      console.error("Erreur lors de la création de l'Application ArgoCD:", error)
      throw error
    }
  },

  syncApplication: async (name: string, prune = false): Promise<void> => {
    try {
      await apiClient.post(`/v1/k8s/argocd/applications/${encodeURIComponent(name)}/sync`, null, {
        params: prune ? { prune: true } : undefined,
      })
    } catch (error) {
      console.error(`Erreur lors de la synchronisation de l'Application ${name}:`, error)
      throw error
    }
  },

  refreshApplication: async (name: string): Promise<void> => {
    try {
      await apiClient.post(`/v1/k8s/argocd/applications/${encodeURIComponent(name)}/refresh`)
    } catch (error) {
      console.error(`Erreur lors du rafraîchissement de l'Application ${name}:`, error)
      throw error
    }
  },

  rollbackApplication: async (name: string, id: number): Promise<void> => {
    try {
      await apiClient.post(`/v1/k8s/argocd/applications/${encodeURIComponent(name)}/rollback`, { id })
    } catch (error) {
      console.error(`Erreur lors du rollback de l'Application ${name}:`, error)
      throw error
    }
  },

  deleteApplication: async (name: string): Promise<void> => {
    try {
      await apiClient.delete(`/v1/k8s/argocd/applications/${encodeURIComponent(name)}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression de l'Application ${name}:`, error)
      throw error
    }
  },

  searchHelmCatalog: async (query = '', page = 1): Promise<HelmChartSummary[]> => {
    try {
      const response = await apiClient.get<{ items: HelmChartSummary[] }>('/v1/k8s/argocd/helm-catalog', {
        params: { q: query || undefined, page },
      })
      return response.data?.items ?? []
    } catch (error) {
      console.error('Erreur lors de la récupération du catalogue Helm:', error)
      throw error
    }
  },
}
