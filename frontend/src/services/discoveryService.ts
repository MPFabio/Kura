import { apiClient } from './api'

export interface DiscoveredApplication {
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

export interface DiscoveredComponent {
  name: string
  found: boolean
  namespace?: string
  pod_name?: string
}

export interface DiscoveryReport {
  applications: DiscoveredApplication[]
  observability: DiscoveredComponent[]
}

// discoveryService interroge l'auto-découverte du cluster client : les
// Applications ArgoCD déployées et les composants d'observabilité
// (Prometheus/VictoriaMetrics, Grafana, Loki, Tempo) reconnus automatiquement
// par labels, sans configuration côté client.
export const discoveryService = {
  getReport: async (): Promise<DiscoveryReport> => {
    const { data } = await apiClient.get<DiscoveryReport>('/v1/k8s/discovery')
    return data
  },
}
