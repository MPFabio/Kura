import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User } from '../services/api'
import { authService } from '../services/authService'

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name?: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('token')
      if (token) {
        try {
          const currentUser = await authService.getCurrentUser()
          setUser(currentUser)
        } catch (error) {
          console.error('Erreur lors de la récupération de l\'utilisateur:', error)
          localStorage.removeItem('token')
          localStorage.removeItem('refreshToken')
          setUser(null)
        }
      } else {
        setUser(null)
      }
      setLoading(false)
    }

    const onSessionExpired = () => {
      setUser(null)
    }

    initAuth()
    window.addEventListener('auth:session-expired', onSessionExpired)
    return () => window.removeEventListener('auth:session-expired', onSessionExpired)
  }, [])

  const login = async (email: string, password: string) => {
    const response = await authService.login({ email, password })
    localStorage.setItem('token', response.token)
    localStorage.setItem('refreshToken', response.refresh_token)
    setUser(response.user)
  }

  const register = async (email: string, password: string, name?: string) => {
    // Le backend attend un username et éventuellement first_name / last_name.
    // On dérive un username simple à partir de la partie locale de l'email.
    const localPart = email.split('@')[0] || 'user'
    const username =
      localPart.length >= 3
        ? localPart
        : `user_${localPart}`

    await authService.register({
      email,
      username,
      password,
      first_name: name,
    })
  }

  const logout = async () => {
    await authService.logout()
    setUser(null)
  }

  const refreshUser = async () => {
    try {
      const currentUser = await authService.getCurrentUser()
      setUser(currentUser)
    } catch (error) {
      console.error('Erreur lors du rafraîchissement de l\'utilisateur:', error)
    }
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}
