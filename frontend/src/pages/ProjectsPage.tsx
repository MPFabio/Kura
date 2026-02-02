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
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await projectService.getProjects()
      return response.items || []
    },
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
    <>
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
          background: 'rgba(0, 0, 0, 0.3)',
          backdropFilter: 'blur(30px) saturate(180%)',
          borderBottom: '1px solid rgba(176, 190, 197, 0.1)',
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
            py: 2.5,
            minHeight: '80px',
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
      <Box sx={{ minHeight: '100vh', position: 'relative', width: '100vw', p: 4, pt: '130px', zIndex: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 5 }}>
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
            p: 4,
            textAlign: 'center',
            background: 'linear-gradient(135deg, rgba(236, 64, 122, 0.05) 0%, rgba(178, 235, 242, 0.9) 50%, rgba(224, 247, 250, 0.95) 100%)',
            backdropFilter: 'blur(30px) saturate(180%)',
            borderRadius: '32px',
            border: '1px solid rgba(171, 71, 188, 0.55)',
            boxShadow: '0 8px 32px rgba(236, 64, 122, 0.2), 0 4px 12px rgba(171, 71, 188, 0.15)',
          }}
        >
          <FolderIcon sx={{ fontSize: 64, color: '#EC407A', mb: 2, filter: 'drop-shadow(0 0 12px rgba(236, 64, 122, 0.5))' }} />
          <ModuleSubtitle sx={{ mb: 2, color: '#f0f0f0' }}>
            Aucun projet
          </ModuleSubtitle>
          <ModuleSecondaryText sx={{ mb: 3 }}>
            Utilisez le bouton "Créer un projet" en haut à droite pour commencer
          </ModuleSecondaryText>
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
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                    <Box
                      sx={{
                        position: 'relative',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 60,
                        height: 60,
                        mb: 1,
                      }}
                    >
                      {/* Cercles orbitaux pour effet moderne */}
                      <Box
                        sx={{
                          position: 'absolute',
                          width: 60,
                          height: 60,
                          border: '1px solid rgba(236, 64, 122, 0.35)',
                          borderRadius: '50%',
                          animation: 'constructAnimation 8s linear infinite',
                        }}
                      />
                      <Box
                        sx={{
                          position: 'absolute',
                          width: 70,
                          height: 70,
                          border: '1px solid transparent',
                          borderTop: '1px solid rgba(0, 229, 255, 0.5)',
                          borderRight: '1px solid rgba(179, 136, 255, 0.4)',
                          borderRadius: '50%',
                          animation: 'constructAnimation 6s linear infinite reverse',
                        }}
                      />
                      {/* Icône dossier avec effet de lueur */}
                      <FolderIcon 
                        sx={{ 
                          fontSize: 40, 
                          color: '#00E5FF', 
                          position: 'relative',
                          zIndex: 1,
                          filter: 'drop-shadow(0 0 20px rgba(0, 229, 255, 0.8)) drop-shadow(0 0 40px rgba(236, 64, 122, 0.5))',
                          animation: 'breathingGlow 3s ease-in-out infinite',
                        }} 
                      />
                    </Box>
                    {isActive && (
                      <Chip
                        icon={<CheckCircleIcon />}
                        label="Actif"
                        size="small"
                        color="success"
                        sx={{
                          backgroundColor: 'rgba(102, 187, 106, 0.2)',
                          color: '#81C784',
                          border: '1px solid rgba(102, 187, 106, 0.4)',
                          fontWeight: 500,
                        }}
                      />
                    )}
                  </Box>
                  <ModuleSubtitle
                    sx={{
                      mb: 2.5,
                      fontSize: '1.5rem',
                      background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
                      backgroundClip: 'text',
                      WebkitBackgroundClip: 'text',
                      WebkitTextFillColor: 'transparent',
                    }}
                  >
                    {project.name}
                  </ModuleSubtitle>
                  {project.description && (
                    <ModuleSecondaryText sx={{ mb: 3 }}>
                      {project.description}
                    </ModuleSecondaryText>
                  )}
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 'auto', pt: 2 }}>
                    <ModuleCaption sx={{ color: '#b8b8b8', fontWeight: 500 }}>
                      {new Date(project.created_at).toLocaleDateString('fr-FR')}
                    </ModuleCaption>
                    <IconButton
                      size="small"
                      onClick={(e) => handleDeleteProject(project.id, e)}
                      sx={{
                        color: '#F48FB1',
                        '&:hover': {
                          color: '#EC407A',
                          background: 'rgba(236, 64, 122, 0.2)',
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
            background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
            backgroundClip: 'text',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            fontWeight: 600,
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
              background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
              '&:hover': {
                background: 'linear-gradient(135deg, #00B8D4, #9C6ADE)',
              },
            }}
          >
            {createProjectMutation.isPending ? <CircularProgress size={20} /> : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>
      </Box>
    </>
  )
}
