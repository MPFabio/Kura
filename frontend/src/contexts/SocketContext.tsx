import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { io, Socket } from 'socket.io-client'
import { useAuth } from './AuthContext'

interface SocketContextType {
  socket: Socket | null
  connected: boolean
  events: any[]
}

const SocketContext = createContext<SocketContextType | undefined>(undefined)

export const useSocket = () => {
  const context = useContext(SocketContext)
  if (!context) {
    throw new Error('useSocket must be used within a SocketProvider')
  }
  return context
}

const SOCKET_URL = import.meta.env.VITE_SOCKET_URL || 'http://localhost:8000'

export const SocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [socket, setSocket] = useState<Socket | null>(null)
  const [connected, setConnected] = useState(false)
  const [events, setEvents] = useState<any[]>([])
  const { user } = useAuth()

  useEffect(() => {
    if (!user) {
      return
    }

    // Créer la connexion WebSocket avec gestion d'erreur silencieuse
    // Note: Socket.io n'est pas encore implémenté dans le backend, donc on ignore les erreurs
    let newSocket: Socket | null = null
    
    try {
      newSocket = io(SOCKET_URL, {
        transports: ['websocket', 'polling'], // Essayer polling en fallback
        autoConnect: true,
        reconnection: true,
        reconnectionAttempts: 3,
        reconnectionDelay: 1000,
        timeout: 5000,
        auth: {
          token: localStorage.getItem('token'),
        },
      })

      newSocket.on('connect', () => {
        console.log('[Socket] WebSocket connecté')
        setConnected(true)
      })

      newSocket.on('disconnect', (reason) => {
        console.log('[Socket] WebSocket déconnecté:', reason)
        setConnected(false)
      })

      newSocket.on('connect_error', (error) => {
        // Erreur silencieuse - Socket.io n'est pas encore implémenté
        console.log('[Socket] Erreur de connexion (ignorée - service non implémenté):', error.message)
        setConnected(false)
      })

      newSocket.on('k8s:event', (data: any) => {
        console.log('[Socket] Événement K8s reçu:', data)
        setEvents((prev) => [...prev.slice(-99), { type: 'k8s', data, timestamp: new Date() }])
      })

      newSocket.on('terraform:event', (data: any) => {
        console.log('[Socket] Événement Terraform reçu:', data)
        setEvents((prev) => [...prev.slice(-99), { type: 'terraform', data, timestamp: new Date() }])
      })

      newSocket.on('pipeline:event', (data: any) => {
        console.log('[Socket] Événement Pipeline reçu:', data)
        setEvents((prev) => [...prev.slice(-99), { type: 'pipeline', data, timestamp: new Date() }])
      })

      newSocket.on('alert', (data: any) => {
        console.log('[Socket] Alerte reçue:', data)
        setEvents((prev) => [...prev.slice(-99), { type: 'alert', data, timestamp: new Date() }])
      })

      setSocket(newSocket)
    } catch (error) {
      // Erreur silencieuse - Socket.io n'est pas encore implémenté
      console.log('[Socket] Impossible de créer la connexion (service non implémenté)')
      setConnected(false)
    }

    return () => {
      if (newSocket) {
        newSocket.close()
      }
    }
  }, [user])

  return (
    <SocketContext.Provider value={{ socket, connected, events }}>
      {children}
    </SocketContext.Provider>
  )
}
