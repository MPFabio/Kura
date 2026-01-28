import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

// Utiliser par défaut le même origin que le frontend,
// et laisser nginx proxyfier /api vers Kong (voir nginx.conf).
// On peut toujours surcharger via VITE_API_BASE_URL si besoin.
const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL && import.meta.env.VITE_API_BASE_URL.trim() !== ''
    ? import.meta.env.VITE_API_BASE_URL
    : '/api'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Intercepteur pour ajouter le token d'authentification
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('token')
        if (token && config.headers) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Intercepteur pour gérer les erreurs 401 (non autorisé)
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Token expiré, déconnexion
          localStorage.removeItem('token')
          localStorage.removeItem('refreshToken')
          // Ne pas rediriger si on est déjà sur la page de login
          // pour éviter les boucles de redirection
          if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
            window.location.href = '/login'
          }
        }
        return Promise.reject(error)
      }
    )
  }

  get instance() {
    return this.client
  }
}

export const apiClient = new ApiClient().instance

// Types pour les réponses API
export interface User {
  id: string
  email: string
  name?: string
  roles: string[]
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  refresh_token: string
  user: User
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
  first_name?: string
  last_name?: string
}

export interface Namespace {
  name: string
  status?: string
  created_at?: string
}

export interface Pod {
  name: string
  namespace: string
  status: string
  node?: string
  created_at?: string
}

export interface K8sNamespacesResponse {
  items: Namespace[]
}

export interface K8sPodsResponse {
  items: Pod[]
}

export interface Deployment {
  name: string
  namespace: string
  replicas: number
  readyReplicas: number
  availableReplicas: number
  creationTimestamp: string
  labels?: Record<string, string>
}

export interface Service {
  name: string
  namespace: string
  type: string
  clusterIP: string
  ports?: ServicePort[]
  creationTimestamp: string
  labels?: Record<string, string>
}

export interface ServicePort {
  name?: string
  port: number
  protocol: string
  targetPort?: string
}

export interface ConfigMap {
  name: string
  namespace: string
  dataKeys: string[]
  creationTimestamp: string
  labels?: Record<string, string>
}

export interface Secret {
  name: string
  namespace: string
  type: string
  dataKeys: string[]
  creationTimestamp: string
  labels?: Record<string, string>
}

export interface Node {
  name: string
  status: string
  kubeletVersion: string
  osImage: string
  architecture: string
  cpu: string
  memory: string
  pods: string
  creationTimestamp: string
  labels?: Record<string, string>
}

export interface K8sDeploymentsResponse {
  items: Deployment[]
}

export interface K8sServicesResponse {
  items: Service[]
}

export interface K8sConfigMapsResponse {
  items: ConfigMap[]
}

export interface K8sSecretsResponse {
  items: Secret[]
}

export interface K8sNodesResponse {
  items: Node[]
}

export interface Event {
  name: string
  namespace: string
  type: string
  reason: string
  message: string
  firstTimestamp: string
  lastTimestamp: string
  count: number
  involvedObject: string
  involvedObjectKind: string
}
