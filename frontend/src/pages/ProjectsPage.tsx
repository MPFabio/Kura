import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Box,
  Card,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  CircularProgress,
  IconButton,
  Chip,
  Avatar,
  Menu,
  MenuItem,
  Typography,
} from '@mui/material'
import {
  Add as AddIcon,
  Folder as FolderIcon,
  Delete as DeleteIcon,
  Logout,
  CheckCircle as CheckCircleIcon,
} from '@mui/icons-material'
import { useProject } from '../contexts/ProjectContext'
import { useAuth } from '../contexts/AuthContext'
import { projectService, CreateProjectRequest } from '../services/projectService'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import Logo from '../components/Logo'
import AnimatedBackground from '../components/AnimatedBackground'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import ModuleButton from '../components/ModuleButton'
import { ModuleSubtitle, ModuleSecondaryText, ModuleCaption } from '../components/ModuleText'

export default function ProjectsPage() {
  const navigate = useNavigate()
  const { currentProject, setCurrentProject, projects, refreshProjects } = useProject()
  const { user, logout } = useAuth()
  const [openDialog, setOpenDialog] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [projectDescription, setProjectDescription] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const queryClient = useQueryClient()

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleMenuClose = () => {
    setAnchorEl(null)
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
    handleMenuClose()
  }

  const { data: projectsData, isLoading } = useQuery({
    queryKey: ['projects', user?.id],
    queryFn: async () => {
      const response = await projectService.getProjects()
      return response.items || []
    },
    enabled: !!user,
  })

  const createProjectMutation = useMutation({
    mutationFn: (data: CreateProjectRequest) => projectService.createProject(data),
    onSuccess: async (newProject) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      // Rafraîchir la liste des projets dans le contexte
      await refreshProjects()
      setOpenDialog(false)
      setProjectName('')
      setProjectDescription('')
      setError(null)
      // Sélectionner automatiquement le projet créé et rediriger
      if (newProject) {
        setCurrentProject(newProject)
        navigate('/modules')
      }
    },
    onError: (err: any) => {
      console.error('Erreur création projet:', err)
      const errorMessage = err.response?.data?.error || err.message || 'Erreur lors de la création du projet'
      setError(errorMessage)
    },
  })

  const deleteProjectMutation = useMutation({
    mutationFn: (id: string) => projectService.deleteProject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      if (currentProject && projectsData?.some(p => p.id === currentProject.id)) {
        setCurrentProject(null)
      }
    },
  })

  const handleCreateProject = () => {
    if (!projectName.trim()) {
      setError('Le nom du projet est requis')
      return
    }

    createProjectMutation.mutate({
      name: projectName.trim(),
      description: projectDescription.trim() || undefined,
    })
  }

  const handleSelectProject = (project: any) => {
    setCurrentProject(project)
    navigate('/modules')
  }

  const handleDeleteProject = async (projectId: string, event: React.MouseEvent) => {
    event.stopPropagation()
    if (window.confirm('Êtes-vous sûr de vouloir supprimer ce projet ? Cette action est irréversible.')) {
      deleteProjectMutation.mutate(projectId)
    }
  }

  const displayedProjects = projectsData || projects

  return (
    <Box
      component="div"
      data-page="projects"
      sx={{ position: 'relative', zIndex: 1, minHeight: '100vh', background: '#32364a' }}
      style={{ minHeight: '100vh', background: '#32364a', position: 'relative', zIndex: 1 }}
    >
      <AnimatedBackground />
      {/* Header simple en haut avec logo et avatar - pleine largeur */}
      <Box
        component="header"
        sx={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          width: '100vw',
          background: '#2c2f3f',
          borderBottom: '2px solid rgba(0, 229, 255, 0.2)',
          borderRadius: 0,
          zIndex: 1000,
        }}
      >
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            px: { xs: 3, sm: 4 },
            py: 3,
            minHeight: '100px',
          }}
        >
          <Logo variant="full" size="small" />
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <IconButton
              size="small"
              edge="end"
              aria-label="account menu"
              aria-controls="account-menu"
              aria-haspopup="true"
              onClick={handleMenuClick}
              sx={{ p: 0.5 }}
            >
              <Avatar sx={{ width: 32, height: 32, fontSize: '0.875rem' }}>
                {user?.email?.charAt(0).toUpperCase() || 'U'}
              </Avatar>
            </IconButton>
            <Menu
              id="account-menu"
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
            >
              <MenuItem onClick={handleLogout}>
                <Logout sx={{ mr: 1 }} />
                Déconnexion
              </MenuItem>
            </Menu>
          </Box>
        </Box>
      </Box>
      {/* Contenu principal - EXACTEMENT même style que ModulesPage */}
      <Box sx={{ minHeight: '100vh', position: 'relative', width: '100vw', p: 4, pt: '170px', pb: 6, zIndex: 1, background: '#32364a', color: '#f0f0f0' }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 5, mt: 1 }}>
          <ModuleTitle sx={{ mb: 0 }}>Projets</ModuleTitle>
          <ModuleButton
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
          >
            Créer un projet
          </ModuleButton>
        </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
          <CircularProgress />
        </Box>
      ) : displayedProjects.length === 0 ? (
        <Card
          sx={{
            p: 6,
            textAlign: 'center',
            borderLeft: '4px solid #AB47BC',
          }}
        >
          <FolderIcon sx={{ fontSize: 64, color: '#AB47BC', mb: 3, opacity: 0.6 }} />
          <Typography sx={{ mb: 2, color: '#f0f0f0', fontSize: '1.25rem', fontWeight: 700 }}>
            Aucun projet
          </Typography>
          <Typography sx={{ color: '#a0a0a0', fontSize: '0.9375rem' }}>
            Utilisez le bouton "Créer un projet" pour commencer
          </Typography>
        </Card>
      ) : (
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)' },
            gap: 3,
          }}
        >
          {displayedProjects.map((project) => {
            const isActive = currentProject?.id === project.id
            return (
              <ModuleCard
                key={project.id}
                active={isActive}
                inactive={!isActive}
                onClick={() => handleSelectProject(project)}
                sx={{
                  minHeight: '300px',
                }}
              >
                <Box sx={{ position: 'relative', width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
                    <FolderIcon sx={{ fontSize: 48, color: '#00E5FF' }} />
                    {isActive && (
                      <Box
                        sx={{
                          px: 1.5,
                          py: 0.5,
                          border: '1px solid #81C784',
                          color: '#81C784',
                          fontSize: '0.6875rem',
                          fontWeight: 700,
                          letterSpacing: '0.05em',
                        }}
                      >
                        ACTIF
                      </Box>
                    )}
                  </Box>
                  <Typography
                    sx={{
                      mb: 2,
                      fontSize: '1.5rem',
                      fontWeight: 700,
                      color: '#f0f0f0',
                    }}
                  >
                    {project.name}
                  </Typography>
                  {project.description && (
                    <Typography sx={{ mb: 3, color: '#a0a0a0', fontSize: '0.9375rem', lineHeight: 1.6 }}>
                      {project.description}
                    </Typography>
                  )}
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 'auto', pt: 3, borderTop: '1px solid rgba(255, 255, 255, 0.08)' }}>
                    <Typography sx={{ color: '#707070', fontSize: '0.8125rem', fontFamily: '"JetBrains Mono", monospace' }}>
                      {new Date(project.created_at).toLocaleDateString('fr-FR')}
                    </Typography>
                    <IconButton
                      size="small"
                      onClick={(e) => handleDeleteProject(project.id, e)}
                      sx={{
                        color: '#EC407A',
                        '&:hover': {
                          background: 'rgba(236, 64, 122, 0.15)',
                        },
                      }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Box>
              </ModuleCard>
            )
          })}
        </Box>
      )}

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle
          sx={{
            color: '#f0f0f0',
            fontWeight: 700,
          }}
        >
          Créer un nouveau projet
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Nom du projet"
            fullWidth
            variant="outlined"
            value={projectName}
            onChange={(e) => setProjectName(e.target.value)}
            sx={{ mb: 2, mt: 2 }}
          />
          <TextField
            margin="dense"
            label="Description (optionnel)"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={projectDescription}
            onChange={(e) => setProjectDescription(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Annuler</Button>
          <Button
            onClick={handleCreateProject}
            variant="contained"
            disabled={createProjectMutation.isPending || !projectName.trim()}
            sx={{
              background: '#00E5FF',
              color: '#0d0e12',
              fontWeight: 700,
              '&:hover': {
                background: '#26C6DA',
              },
            }}
          >
            {createProjectMutation.isPending ? <CircularProgress size={20} sx={{ color: '#0d0e12' }} /> : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>
      </Box>
    </Box>
  )
}
