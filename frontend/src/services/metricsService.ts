import { apiClient } from './api'

export interface ServiceHealth {
  name: string
  job: string
  up: boolean
  goroutines: number
  memory_mb: number
}

export interface ServiceMetric {
  name: string
  job: string
  up: boolean
  goroutines: number
  cpu_rate: number
  memory_mb: number
}

export interface Overview {
  total_services: number
  services_up: number
  services_down: number
  total_goroutines: number
  total_memory_mb: number
}

export interface LogEntry {
  timestamp: string
  line: string
  labels: Record<string, string>
}

export interface TraceSummary {
  trace_id: string
  root_service_name: string
  root_trace_name: string
  start_time_unix_nano: string
  duration_ms: number
}

export interface TraceSpan {
  spanID: string
  name: string
  startTimeUnixNano: string
  durationNanos?: string
  attributes?: { key: string; value: Record<string, unknown> }[]
}

export interface TraceScopeSpans {
  scope?: { name: string }
  spans: TraceSpan[]
}

export interface TraceResourceSpans {
  resource?: { attributes?: { key: string; value: Record<string, unknown> }[] }
  scopeSpans: TraceScopeSpans[]
}

export interface TraceDetail {
  batches?: TraceResourceSpans[]
  resourceSpans?: TraceResourceSpans[]
}

export interface PlatformConfig {
  internal_observability_enabled: boolean
}

// projectObservabilityService interroge la stack d'observabilité déployée
// dans le cluster du client (Prometheus/Loki/Tempo via le catalogue Helm
// ArgoCD), à travers le k8s-service (port-forward vers le cluster actif).
// Les formes de réponses sont adaptées pour réutiliser les mêmes composants
// d'affichage que l'observabilité interne de Kura.
export const projectObservabilityService = {
  getOverview: async (): Promise<Overview> => {
    const { data } = await apiClient.get<{
      total_services: number
      total_goroutines: number
      total_memory_mb: number
    }>('/v1/k8s/observability/overview')
    return {
      total_services: data.total_services ?? 0,
      services_up: data.total_services ?? 0,
      services_down: 0,
      total_goroutines: data.total_goroutines ?? 0,
      total_memory_mb: data.total_memory_mb ?? 0,
    }
  },

  getServices: async (): Promise<ServiceMetric[]> => {
    const { data } = await apiClient.get<Record<string, { goroutines?: number; cpu_rate?: number; memory_mb?: number }>>('/v1/k8s/observability/services')
    return Object.entries(data ?? {}).map(([job, m]) => ({
      name: job,
      job,
      up: true,
      goroutines: m.goroutines ?? 0,
      cpu_rate: m.cpu_rate ?? 0,
      memory_mb: m.memory_mb ?? 0,
    }))
  },

  getHealth: async (): Promise<ServiceHealth[]> => {
    const services = await projectObservabilityService.getServices()
    return services.map((s) => ({ name: s.name, job: s.job, up: s.up, goroutines: s.goroutines, memory_mb: s.memory_mb }))
  },

  getLogs: async (params: { service?: string; search?: string; limit?: number }): Promise<LogEntry[]> => {
    const { data } = await apiClient.get<{ items: LogEntry[] }>('/v1/k8s/observability/logs', { params })
    return data.items ?? []
  },

  getLogServices: async (): Promise<string[]> => {
    const { data } = await apiClient.get<{ items: string[] }>('/v1/k8s/observability/logs/services')
    return data.items ?? []
  },

  searchTraces: async (params: { service?: string; min_duration_ms?: number; limit?: number }): Promise<TraceSummary[]> => {
    const { data } = await apiClient.get<{ items: TraceSummary[] }>('/v1/k8s/observability/traces', { params })
    return data.items ?? []
  },

  getTrace: async (traceId: string): Promise<TraceDetail> => {
    const { data } = await apiClient.get<TraceDetail>(`/v1/k8s/observability/traces/${traceId}`)
    return data
  },
}

export const metricsService = {
  getPlatformConfig: async (): Promise<PlatformConfig> => {
    const { data } = await apiClient.get<PlatformConfig>('/v1/metrics/platform-config')
    return data
  },

  getHealth: async (): Promise<ServiceHealth[]> => {
    const { data } = await apiClient.get<ServiceHealth[]>('/v1/metrics/health')
    return data
  },

  getServices: async (): Promise<ServiceMetric[]> => {
    const { data } = await apiClient.get<ServiceMetric[]>('/v1/metrics/services')
    return data
  },

  getOverview: async (): Promise<Overview> => {
    const { data } = await apiClient.get<Overview>('/v1/metrics/overview')
    return data
  },

  getLogs: async (params: { service?: string; search?: string; limit?: number }): Promise<LogEntry[]> => {
    const { data } = await apiClient.get<{ items: LogEntry[] }>('/v1/metrics/logs', { params })
    return data.items ?? []
  },

  getLogServices: async (): Promise<string[]> => {
    const { data } = await apiClient.get<{ items: string[] }>('/v1/metrics/logs/services')
    return data.items ?? []
  },

  searchTraces: async (params: { service?: string; min_duration_ms?: number; limit?: number }): Promise<TraceSummary[]> => {
    const { data } = await apiClient.get<{ items: TraceSummary[] }>('/v1/metrics/traces', { params })
    return data.items ?? []
  },

  getTrace: async (traceId: string): Promise<TraceDetail> => {
    const { data } = await apiClient.get<TraceDetail>(`/v1/metrics/traces/${traceId}`)
    return data
  },
}
