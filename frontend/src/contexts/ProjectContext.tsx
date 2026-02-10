import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { projectService, Project } from '../services/projectService'
import { useAuth } from './AuthContext'

interface ProjectContextType {
  currentProject: Project | null
  loading: boolean
  setCurrentProject: (project: Project | null) => void
  refreshProjects: () => Promise<void>
  projects: Project[]
}

const ProjectContext = createContext<ProjectContextType | undefined>(undefined)

export const useProject = () => {
  const context = useContext(ProjectContext)
  if (!context) {
    throw new Error('useProject must be used within a ProjectProvider')
  }
  return context
}

export const ProjectProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [currentProject, setCurrentProjectState] = useState<Project | null>(null)
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const { user, loading: authLoading } = useAuth()

  const setCurrentProject = (project: Project | null) => {
    setCurrentProjectState(project)
    if (project) {
      localStorage.setItem('currentProjectId', project.id)
    } else {
      localStorage.removeItem('currentProjectId')
    }
  }

  const refreshProjects = async () => {
    try {
      const response = await projectService.getProjects()
      setProjects(response.items || [])
      
      // Si un projet était sélectionné, le restaurer
      const savedProjectId = localStorage.getItem('currentProjectId')
      if (savedProjectId) {
        const savedProject = response.items?.find(p => p.id === savedProjectId)
        if (savedProject) {
          setCurrentProject(savedProject)
        } else {
          // Le projet sauvegardé n'existe plus ou l'utilisateur n'y a plus accès
          localStorage.removeItem('currentProjectId')
          setCurrentProject(null)
        }
      }
    } catch (error) {
      console.error('Erreur lors du chargement des projets:', error)
      setProjects([])
      setCurrentProject(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const initProject = async () => {
      // Ne charger les projets que si l'utilisateur est authentifié
      if (!authLoading && user) {
        await refreshProjects()
      } else if (!authLoading && !user) {
        // Si pas d'utilisateur, ne pas charger mais arrêter le loading
        setLoading(false)
      }
    }

    initProject()
  }, [authLoading, user])

  // Rediriger vers /projects si utilisateur connecté, aucun projet sélectionné, et pas déjà sur /projects
  useEffect(() => {
    if (!user) return
    if (!loading && !currentProject) {
      const currentPath = window.location.pathname
      const willRedirect = currentPath !== '/projects' && !currentPath.startsWith('/login') && !currentPath.startsWith('/register')
      if (willRedirect) {
        navigate('/projects', { replace: true })
      }
    }
  }, [user, loading, currentProject, navigate])

  return (
    <ProjectContext.Provider
      value={{
        currentProject,
        loading,
        setCurrentProject,
        refreshProjects,
        projects,
      }}
    >
      {children}
    </ProjectContext.Provider>
  )
}
