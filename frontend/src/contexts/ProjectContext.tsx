import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { projectService, Project } from '../services/projectService'

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
      await refreshProjects()
    }

    initProject()
  }, [])

  // Rediriger vers /projects si aucun projet n'est sélectionné et qu'on n'est pas déjà sur /projects
  useEffect(() => {
    if (!loading && !currentProject) {
      const currentPath = window.location.pathname
      if (currentPath !== '/projects' && !currentPath.startsWith('/login') && !currentPath.startsWith('/register')) {
        navigate('/projects', { replace: true })
      }
    }
  }, [loading, currentProject, navigate])

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
