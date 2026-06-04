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

export const metricsService = {
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
}
