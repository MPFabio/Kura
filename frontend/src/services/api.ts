import axios from 'axios'

const rawBaseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_BASE_URL != null
    ? String(import.meta.env.VITE_API_BASE_URL).replace(/\/$/, '')
    : ''

// Garde-fou sur la valeur injectée à la compilation.
//
// Une image construite hors de docker compose peut recevoir une valeur
// aberrante sans que rien ne le signale : un `--build-arg VITE_API_BASE_URL=/api`
// lancé depuis Git Bash sous Windows arrive ainsi sous la forme
// « C:\Program Files\Git\api », MSYS convertissant les chemins absolus. Toutes
// les requêtes échouent alors avant même d'être émises, avec un « Unsupported
// protocol » qu'aucun écran ne montre.
//
// On préfère repartir sur une base relative, qui fonctionne derrière le proxy,
// plutôt que de servir une application inutilisable.
function sanitizeBaseURL(value: string): string {
  const looksLikeFilesystemPath = /^[a-zA-Z]:[\\/]/.test(value) || value.startsWith('\\\\')
  if (looksLikeFilesystemPath) {
    console.error(
      `[Kura] VITE_API_BASE_URL invalide (« ${value} ») : cette image a été construite avec un argument corrompu. ` +
        'Reconstruire avec `docker compose build frontend`. Repli sur une base relative.'
    )
    return ''
  }
  return value
}

const baseURL = sanitizeBaseURL(rawBaseURL)

export const apiClient = axios.create({
  baseURL,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // Projet actif : permet aux services de vérifier les permissions par module
  const projectId = localStorage.getItem('currentProjectId')
  if (projectId) {
    config.headers['X-Project-ID'] = projectId
  }
  return config
})

const authBaseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_AUTH_URL
    ? String(import.meta.env.VITE_AUTH_URL).replace(/\/$/, '')
    : baseURL

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      const refreshToken = localStorage.getItem('refreshToken')
      if (refreshToken) {
        try {
          const { data } = await axios.post(authBaseURL + '/v1/auth/refresh', {
            refresh_token: refreshToken,
          })
          localStorage.setItem('token', data.token)
          if (data.refresh_token) localStorage.setItem('refreshToken', data.refresh_token)
          originalRequest.headers.Authorization = `Bearer ${data.token}`
          return apiClient(originalRequest)
        } catch (_) {
          localStorage.removeItem('token')
          localStorage.removeItem('refreshToken')
          window.dispatchEvent(new CustomEvent('auth:session-expired'))
        }
      } else {
        localStorage.removeItem('token')
        localStorage.removeItem('refreshToken')
        window.dispatchEvent(new CustomEvent('auth:session-expired'))
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

export interface K8sPodDetail {
  metadata?: { name?: string; namespace?: string }
  spec?: { containers?: { name: string }[] }
  status?: { phase?: string; containerStatuses?: { name: string; ready: boolean }[] }
  [key: string]: unknown
}

export interface K8sDeploymentDetail {
  metadata?: { name?: string; namespace?: string }
  spec?: { replicas?: number; selector?: { matchLabels?: Record<string, string> } }
  status?: { replicas?: number; readyReplicas?: number; availableReplicas?: number; updatedReplicas?: number; conditions?: { type: string; status: string; message?: string }[] }
  [key: string]: unknown
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
