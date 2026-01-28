import axios from 'axios'

const baseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_BASE_URL != null
    ? String(import.meta.env.VITE_API_BASE_URL).replace(/\/$/, '')
    : ''

export const apiClient = axios.create({
  baseURL,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      const refreshToken = localStorage.getItem('refreshToken')
      if (refreshToken) {
        try {
          const { data } = await axios.post(baseURL + '/api/v1/auth/refresh', {
            refresh_token: refreshToken,
          })
          localStorage.setItem('token', data.token)
          if (data.refresh_token) localStorage.setItem('refreshToken', data.refresh_token)
          originalRequest.headers.Authorization = `Bearer ${data.token}`
          return apiClient(originalRequest)
        } catch (_) {
          localStorage.removeItem('token')
          localStorage.removeItem('refreshToken')
        }
      }
    }
    return Promise.reject(error)
  }
)

// Auth types
export interface User {
  id: string
  email: string
  username: string
  roles?: string[]
  first_name?: string
  last_name?: string
  active?: boolean
  created_at?: string
  updated_at?: string
  last_login?: string | null
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  refresh_token: string
  user: User
  expires_at?: string
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
  first_name?: string
  last_name?: string
}

// K8s types
export interface K8sNamespacesResponse {
  items: { name: string; [key: string]: unknown }[]
}

export interface K8sPodsResponse {
  items: { name: string; namespace?: string; [key: string]: unknown }[]
}

export interface Deployment {
  name: string
  namespace?: string
  replicas?: number
  [key: string]: unknown
}

export interface K8sDeploymentsResponse {
  items: Deployment[]
}

export interface K8sServicesResponse {
  items: { name: string; namespace?: string; [key: string]: unknown }[]
}

export interface K8sConfigMapsResponse {
  items: { name: string; namespace?: string; [key: string]: unknown }[]
}

export interface K8sSecretsResponse {
  items: { name: string; namespace?: string; [key: string]: unknown }[]
}

export interface K8sNodesResponse {
  items: { name: string; [key: string]: unknown }[]
}

export interface Event {
  type?: string
  reason?: string
  message?: string
  count?: number
  firstTimestamp?: string
  lastTimestamp?: string
  involvedObject?: { kind?: string; name?: string; namespace?: string }
  [key: string]: unknown
}
