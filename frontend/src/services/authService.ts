import axios, { AxiosInstance } from 'axios'
import { apiClient, LoginRequest, LoginResponse, RegisterRequest, User } from './api'

// Contournement si Kong retourne 502 : appeler auth-service directement (VITE_AUTH_URL=http://localhost:8080)
const authBaseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_AUTH_URL
    ? String(import.meta.env.VITE_AUTH_URL).replace(/\/$/, '')
    : null

const authClient = authBaseURL
  ? axios.create({
      baseURL: authBaseURL,
      headers: { 'Content-Type': 'application/json' },
    })
  : apiClient

if (authBaseURL) {
  authClient.interceptors.request.use((config) => {
    const token = localStorage.getItem('token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })
}

const getAuthClient = () => authClient

// Fallback direct vers auth-service sur 502 (Kong indisponible ou auth-service pas encore vu par Kong)
const FALLBACK_AUTH_URL = 'http://localhost:8080'
let directAuthClient: AxiosInstance | null = null
function getDirectAuthClient(): AxiosInstance {
  if (!directAuthClient) {
    directAuthClient = axios.create({
      baseURL: FALLBACK_AUTH_URL,
      headers: { 'Content-Type': 'application/json' },
    })
    directAuthClient.interceptors.request.use((config) => {
      const token = localStorage.getItem('token')
      if (token) config.headers.Authorization = `Bearer ${token}`
      return config
    })
  }
  return directAuthClient
}

function is502OrBadGateway(error: unknown): boolean {
  const status = (error as { response?: { status?: number } })?.response?.status
  return status === 502 || status === 503 || status === 504
}

async function with502Fallback<T>(request: (client: AxiosInstance) => Promise<T>): Promise<T> {
  try {
    return await request(getAuthClient())
  } catch (e) {
    if (is502OrBadGateway(e) && !authBaseURL) {
      return await request(getDirectAuthClient())
    }
    throw e
  }
}

export const authService = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await with502Fallback((client) =>
      client.post<LoginResponse>('/v1/auth/login', credentials)
    )
    return response.data
  },

  register: async (data: RegisterRequest): Promise<{ message: string; user: User }> => {
    const response = await with502Fallback((client) =>
      client.post('/v1/auth/register', data)
    )
    return response.data
  },

  logout: async (): Promise<void> => {
    const refreshToken = localStorage.getItem('refreshToken')
    if (refreshToken) {
      try {
        await with502Fallback((client) =>
          client.post('/v1/auth/logout', { refresh_token: refreshToken })
        )
      } catch (error) {
        console.error('Erreur lors de la déconnexion:', error)
      }
    }
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  },

  refreshToken: async (refreshToken: string): Promise<LoginResponse> => {
    const response = await with502Fallback((client) =>
      client.post<LoginResponse>('/v1/auth/refresh', { refresh_token: refreshToken })
    )
    return response.data
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await with502Fallback((client) => client.get<User>('/v1/auth/me'))
    return response.data
  },

  updateUser: async (data: Partial<User>): Promise<User> => {
    const response = await with502Fallback((client) =>
      client.put<User>('/v1/auth/me', data)
    )
    return response.data
  },

  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    await with502Fallback((client) =>
      client.put('/v1/auth/password', {
        current_password: currentPassword,
        new_password: newPassword,
      })
    )
  },
}
