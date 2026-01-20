import { apiClient, LoginRequest, LoginResponse, RegisterRequest, User } from './api'

export const authService = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', credentials)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<{ message: string; user: User }> => {
    const response = await apiClient.post('/api/v1/auth/register', data)
    return response.data
  },

  logout: async (): Promise<void> => {
    const refreshToken = localStorage.getItem('refreshToken')
    if (refreshToken) {
      try {
        await apiClient.post('/api/v1/auth/logout', { refresh_token: refreshToken })
      } catch (error) {
        console.error('Erreur lors de la déconnexion:', error)
      }
    }
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  },

  refreshToken: async (refreshToken: string): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/api/v1/auth/refresh', {
      refresh_token: refreshToken,
    })
    return response.data
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get<User>('/api/v1/auth/me')
    return response.data
  },

  updateUser: async (data: Partial<User>): Promise<User> => {
    const response = await apiClient.put<User>('/api/v1/auth/me', data)
    return response.data
  },

  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    await apiClient.put('/api/v1/auth/password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
  },
}
