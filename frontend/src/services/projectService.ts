import axios from 'axios'
import { apiClient } from './api'

// Client pour auth-service (auth + projects) - utilise VITE_AUTH_URL si Kong retourne 502
const authBaseURL =
  typeof import.meta !== 'undefined' && import.meta.env?.VITE_AUTH_URL
    ? String(import.meta.env.VITE_AUTH_URL).replace(/\/$/, '')
    : null

const projectClient = authBaseURL
  ? (() => {
      const client = axios.create({
        baseURL: authBaseURL,
        headers: { 'Content-Type': 'application/json' },
      })
      client.interceptors.request.use((config) => {
        const token = localStorage.getItem('token')
        if (token) config.headers.Authorization = `Bearer ${token}`
        return config
      })
      return client
    })()
  : apiClient

const getProjectClient = () => projectClient

export interface Project {
  id: string
  name: string
  description?: string
  owner_id: string
  created_at: string
  updated_at: string
  members?: ProjectMember[]
}

export interface ProjectMember {
  id: string
  project_id: string
  user_id: string
  role: 'owner' | 'admin' | 'member'
  joined_at: string
  user?: {
    id: string
    email: string
    username: string
    first_name?: string
    last_name?: string
  }
}

export interface ProjectResponse {
  items: Project[]
}

export interface CreateProjectRequest {
  name: string
  description?: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
}

export interface AddProjectMemberRequest {
  user_id: string
  role: 'admin' | 'member'
}

export interface UpdateProjectMemberRequest {
  role: 'admin' | 'member'
}

export const projectService = {
  getProjects: async (): Promise<ProjectResponse> => {
    try {
      const response = await getProjectClient().get<ProjectResponse>('/v1/projects')
      if (!response.data || !response.data.items) {
        return { items: [] }
      }
      return response.data
    } catch (error) {
      console.error('Erreur lors de la récupération des projets:', error)
      throw error
    }
  },

  getProject: async (id: string): Promise<Project> => {
    try {
      const response = await getProjectClient().get<Project>(`/v1/projects/${id}`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération du projet ${id}:`, error)
      throw error
    }
  },

  createProject: async (project: CreateProjectRequest): Promise<Project> => {
    try {
      const response = await getProjectClient().post<Project>('/v1/projects', project)
      return response.data
    } catch (error) {
      console.error('Erreur lors de la création du projet:', error)
      throw error
    }
  },

  updateProject: async (id: string, project: UpdateProjectRequest): Promise<Project> => {
    try {
      const response = await getProjectClient().put<Project>(`/v1/projects/${id}`, project)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la mise à jour du projet ${id}:`, error)
      throw error
    }
  },

  deleteProject: async (id: string): Promise<void> => {
    try {
      await getProjectClient().delete(`/v1/projects/${id}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression du projet ${id}:`, error)
      throw error
    }
  },

  getProjectMembers: async (projectId: string): Promise<{ items: ProjectMember[] }> => {
    try {
      const response = await getProjectClient().get<{ items: ProjectMember[] }>(`/v1/projects/${projectId}/members`)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de la récupération des membres du projet ${projectId}:`, error)
      throw error
    }
  },

  addProjectMember: async (projectId: string, member: AddProjectMemberRequest): Promise<ProjectMember> => {
    try {
      const response = await getProjectClient().post<ProjectMember>(`/v1/projects/${projectId}/members`, member)
      return response.data
    } catch (error) {
      console.error(`Erreur lors de l'ajout du membre au projet ${projectId}:`, error)
      throw error
    }
  },

  updateProjectMember: async (projectId: string, userId: string, member: UpdateProjectMemberRequest): Promise<void> => {
    try {
      await getProjectClient().put(`/v1/projects/${projectId}/members/${userId}`, member)
    } catch (error) {
      console.error(`Erreur lors de la mise à jour du membre ${userId} du projet ${projectId}:`, error)
      throw error
    }
  },

  removeProjectMember: async (projectId: string, userId: string): Promise<void> => {
    try {
      await getProjectClient().delete(`/v1/projects/${projectId}/members/${userId}`)
    } catch (error) {
      console.error(`Erreur lors de la suppression du membre ${userId} du projet ${projectId}:`, error)
      throw error
    }
  },
}
