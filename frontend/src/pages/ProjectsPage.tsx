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
  List,
  ListItem,
  ListItemText,
  Select,
  FormControl,
  InputLabel,
  Tooltip,
} from '@mui/material'
import {
  Add as AddIcon,
  Folder as FolderIcon,
  Delete as DeleteIcon,
  Logout,
  CheckCircle as CheckCircleIcon,
  // GitHub as GitHubIcon, // conservé mais désactivé en prod (remplacé par ForgejoIcon)
  People as PeopleIcon,
} from '@mui/icons-material'
import ForgejoIcon from '../components/icons/ForgejoIcon'
import { useProject } from '../contexts/ProjectContext'
import { useAuth } from '../contexts/AuthContext'
import { projectService, CreateProjectRequest, ProjectMapping, ProjectMember } from '../services/projectService'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import Logo from '../components/Logo'
import AnimatedBackground from '../components/AnimatedBackground'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import ModuleButton from '../components/ModuleButton'
import { ModuleSubtitle, ModuleSecondaryText, ModuleCaption } from '../components/ModuleText'
import { kuraColors } from '../theme'

export default function ProjectsPage() {
  const navigate = useNavigate()
  const { currentProject, setCurrentProject, projects, refreshProjects } = useProject()
  const { user, logout } = useAuth()
  const [openDialog, setOpenDialog] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [projectDescription, setProjectDescription] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const [mappingsProject, setMappingsProject] = useState<{ id: string; name: string } | null>(null)
  const [newRepo, setNewRepo] = useState('')
  const [mappingError, setMappingError] = useState<string | null>(null)
  const [gitopsRepoInputs, setGitopsRepoInputs] = useState<Record<string, string>>({})
  const [membersProject, setMembersProject] = useState<{ id: string; name: string } | null>(null)
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberRole, setNewMemberRole] = useState<'admin' | 'member'>('member')
  const [memberError, setMemberError] = useState<string | null>(null)
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

  const { data: mappingsData, isLoading: mappingsLoading } = useQuery({
    queryKey: ['project-mappings', mappingsProject?.id],
    queryFn: () => projectService.listMappings(mappingsProject!.id),
    enabled: !!mappingsProject,
  })

  const createMappingMutation = useMutation({
    // mutationFn: (repo: string) => projectService.createMapping(mappingsProject!.id, { github_repository: repo }), // conservé mais désactivé en prod
    mutationFn: (repo: string) => projectService.createMapping(mappingsProject!.id, { forgejo_repository: repo }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', mappingsProject?.id] })
      setNewRepo('')
      setMappingError(null)
    },
    onError: (err: any) => {
      setMappingError(err.response?.data?.error || err.message || 'Erreur lors de l\'ajout du dépôt')
    },
  })

  const deleteMappingMutation = useMutation({
    mutationFn: (mappingId: string) => projectService.deleteMapping(mappingsProject!.id, mappingId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', mappingsProject?.id] })
    },
  })

  const setGitOpsRepositoryMutation = useMutation({
    mutationFn: ({ mappingId, repo }: { mappingId: string; repo: string }) =>
      projectService.setMappingGitOpsRepository(mappingsProject!.id, mappingId, repo),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', mappingsProject?.id] })
    },
  })

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['project-members', membersProject?.id],
    queryFn: () => projectService.getProjectMembers(membersProject!.id),
    enabled: !!membersProject,
  })

  const addMemberMutation = useMutation({
    mutationFn: (data: { email: string; role: 'admin' | 'member' }) =>
      projectService.addProjectMember(membersProject!.id, { email: data.email, role: data.role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', membersProject?.id] })
      setNewMemberEmail('')
      setNewMemberRole('member')
      setMemberError(null)
    },
    onError: (err: any) => {
      setMemberError(err.response?.data?.error || err.message || "Erreur lors de l'ajout du membre")
    },
  })

  const updateMemberMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: 'admin' | 'member' }) =>
      projectService.updateProjectMember(membersProject!.id, userId, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', membersProject?.id] })
    },
    onError: (err: any) => {
      setMemberError(err.response?.data?.error || err.message || "Erreur lors de la mise à jour du membre")
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => projectService.removeProjectMember(membersProject!.id, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', membersProject?.id] })
    },
    onError: (err: any) => {
      setMemberError(err.response?.data?.error || err.message || "Erreur lors de la suppression du membre")
    },
  })

  const handleInviteMember = () => {
    const email = newMemberEmail.trim()
    if (!email) {
      setMemberError("L'email est requis")
      return
    }
    addMemberMutation.mutate({ email, role: newMemberRole })
  }

  const handleAddRepo = () => {
    const repo = newRepo.trim()
    if (!repo) return
    if (!/^[\w.-]+\/[\w.-]+$/.test(repo)) {
      setMappingError('Format attendu : owner/repo')
      return
    }
    createMappingMutation.mutate(repo)
  }

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
      sx={{ position: 'relative', zIndex: 1, minHeight: '100vh' }}
      style={{ minHeight: '100vh', position: 'relative', zIndex: 1 }}
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
          background: kuraColors.bg1,
          borderBottom: `1px solid ${kuraColors.border1}`,
          borderRadius: 0,
          zIndex: 1000,
        }}
      >
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            px: 3,
            height: 72,
          }}
        >
          <Logo variant="icon" size="small" />
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
      <Box sx={{ minHeight: '100vh', position: 'relative', width: '100vw', px: 4, pt: '90px', pb: 6, zIndex: 1, color: '#f0f0f0' }}>
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
                  minHeight: 'unset',
                }}
              >
                <Box sx={{ position: 'relative', width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
                    <FolderIcon sx={{ fontSize: 48, color: kuraColors.accent }} />
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
                    <Box sx={{ display: 'flex', gap: 0.5 }}>
                      <Tooltip title="Membres">
                        <IconButton
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation()
                            setMembersProject({ id: project.id, name: project.name })
                            setNewMemberEmail('')
                            setNewMemberRole('member')
                            setMemberError(null)
                          }}
                          sx={{
                            color: kuraColors.text2,
                            '&:hover': {
                              color: kuraColors.text0,
                            },
                          }}
                        >
                          <PeopleIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation()
                          setMappingsProject({ id: project.id, name: project.name })
                          setNewRepo('')
                          setMappingError(null)
                        }}
                        sx={{
                          color: kuraColors.text2,
                          '&:hover': {
                            color: kuraColors.text0,
                          },
                        }}
                      >
                        {/* <GitHubIcon fontSize="small" /> conservé mais désactivé en prod */}
                        <ForgejoIcon active />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={(e) => handleDeleteProject(project.id, e)}
                        sx={{
                          color: kuraColors.error,
                          '&:hover': {
                            background: kuraColors.errorBg,
                          },
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Box>
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
            }}
          >
            {createProjectMutation.isPending ? <CircularProgress size={20} sx={{ color: '#0d0e12' }} /> : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={!!mappingsProject} onClose={() => setMappingsProject(null)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ color: '#f0f0f0', fontWeight: 700 }}>
          Dépôts Forgejo/Codeberg liés — {mappingsProject?.name}
        </DialogTitle>
        <DialogContent>
          <Typography sx={{ color: '#a0a0a0', fontSize: '0.875rem', mb: 2 }}>
            Liez un ou plusieurs dépôts Forgejo/Codeberg (format <code>owner/repo</code>) pour les rendre disponibles dans le module Repository.
          </Typography>

          {mappingError && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setMappingError(null)}>
              {mappingError}
            </Alert>
          )}

          <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
            <TextField
              size="small"
              fullWidth
              placeholder="owner/repo"
              value={newRepo}
              onChange={(e) => setNewRepo(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleAddRepo() }}
            />
            <Button
              variant="contained"
              onClick={handleAddRepo}
              disabled={createMappingMutation.isPending || !newRepo.trim()}
            >
              {createMappingMutation.isPending ? <CircularProgress size={18} sx={{ color: '#0d0e12' }} /> : 'Ajouter'}
            </Button>
          </Box>

          {mappingsLoading ? (
            <CircularProgress size={20} />
          ) : (
            <List dense>
              {(mappingsData?.items ?? [])
                // .filter((m) => !!m.github_repository) // conservé mais désactivé en prod
                .filter((m) => !!m.forgejo_repository)
                .map((m: ProjectMapping) => (
                  <ListItem
                    key={m.id}
                    alignItems="flex-start"
                    secondaryAction={
                      <IconButton
                        size="small"
                        edge="end"
                        onClick={() => deleteMappingMutation.mutate(m.id)}
                        sx={{ color: kuraColors.error }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    }
                  >
                    <Box sx={{ width: '100%', pr: 4 }}>
                      <ListItemText
                        primary={m.forgejo_repository}
                        primaryTypographyProps={{ sx: { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.875rem', color: '#f0f0f0' } }}
                      />
                      <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                        <TextField
                          size="small"
                          fullWidth
                          label="Dépôt GitOps (ArgoCD)"
                          placeholder={`${m.forgejo_repository}-gitops (par défaut)`}
                          helperText="owner/repo — laisser vide pour créer automatiquement un dépôt dédié"
                          value={gitopsRepoInputs[m.id] ?? m.forgejo_gitops_repository ?? ''}
                          onChange={(e) => setGitopsRepoInputs({ ...gitopsRepoInputs, [m.id]: e.target.value })}
                        />
                        <Button
                          variant="outlined"
                          size="small"
                          disabled={
                            setGitOpsRepositoryMutation.isPending ||
                            (gitopsRepoInputs[m.id] ?? m.forgejo_gitops_repository ?? '') === (m.forgejo_gitops_repository ?? '')
                          }
                          onClick={() => setGitOpsRepositoryMutation.mutate({ mappingId: m.id, repo: gitopsRepoInputs[m.id] ?? '' })}
                        >
                          {setGitOpsRepositoryMutation.isPending ? <CircularProgress size={16} /> : 'Enregistrer'}
                        </Button>
                      </Box>
                    </Box>
                  </ListItem>
                ))}
              {(mappingsData?.items ?? []).filter((m) => !!m.forgejo_repository).length === 0 && (
                <Typography sx={{ color: '#707070', fontSize: '0.875rem' }}>
                  Aucun dépôt lié pour le moment.
                </Typography>
              )}
            </List>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMappingsProject(null)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={!!membersProject} onClose={() => setMembersProject(null)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ color: '#f0f0f0', fontWeight: 700 }}>
          Membres — {membersProject?.name}
        </DialogTitle>
        <DialogContent>
          <Typography sx={{ color: '#a0a0a0', fontSize: '0.875rem', mb: 2 }}>
            Ajoutez des collaborateurs par email. Les administrateurs peuvent modifier les ressources du projet et gérer les membres, les membres ont un accès en lecture et peuvent agir sur les ressources sans gérer les membres.
          </Typography>

          {memberError && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setMemberError(null)}>
              {memberError}
            </Alert>
          )}

          <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
            <TextField
              size="small"
              fullWidth
              type="email"
              placeholder="email@exemple.com"
              value={newMemberEmail}
              onChange={(e) => setNewMemberEmail(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleInviteMember() }}
            />
            <FormControl size="small" sx={{ minWidth: 130 }}>
              <InputLabel id="new-member-role-label">Rôle</InputLabel>
              <Select
                labelId="new-member-role-label"
                label="Rôle"
                value={newMemberRole}
                onChange={(e) => setNewMemberRole(e.target.value as 'admin' | 'member')}
              >
                <MenuItem value="member">Member</MenuItem>
                <MenuItem value="admin">Admin</MenuItem>
              </Select>
            </FormControl>
            <Button
              variant="contained"
              onClick={handleInviteMember}
              disabled={addMemberMutation.isPending || !newMemberEmail.trim()}
            >
              {addMemberMutation.isPending ? <CircularProgress size={18} sx={{ color: '#0d0e12' }} /> : 'Inviter'}
            </Button>
          </Box>

          {membersLoading ? (
            <CircularProgress size={20} />
          ) : (
            <List dense>
              {(membersData?.items ?? []).map((m: ProjectMember) => {
                const isOwner = m.role === 'owner'
                const label = m.user?.email || m.user?.username || m.user_id
                return (
                  <ListItem
                    key={m.id}
                    secondaryAction={
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {isOwner ? (
                          <Chip label="Owner" size="small" sx={{ bgcolor: kuraColors.accent, color: '#0d0e12' }} />
                        ) : (
                          <Select
                            size="small"
                            value={m.role}
                            onChange={(e) =>
                              updateMemberMutation.mutate({ userId: m.user_id, role: e.target.value as 'admin' | 'member' })
                            }
                            sx={{ minWidth: 110 }}
                          >
                            <MenuItem value="member">Member</MenuItem>
                            <MenuItem value="admin">Admin</MenuItem>
                          </Select>
                        )}
                        <Tooltip title={isOwner ? "Le propriétaire ne peut pas être retiré" : 'Retirer du projet'}>
                          <span>
                            <IconButton
                              size="small"
                              edge="end"
                              disabled={isOwner}
                              onClick={() => removeMemberMutation.mutate(m.user_id)}
                              sx={{ color: kuraColors.error }}
                            >
                              <DeleteIcon fontSize="small" />
                            </IconButton>
                          </span>
                        </Tooltip>
                      </Box>
                    }
                  >
                    <ListItemText
                      primary={label}
                      primaryTypographyProps={{ sx: { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.875rem', color: '#f0f0f0' } }}
                    />
                  </ListItem>
                )
              })}
              {(membersData?.items ?? []).length === 0 && (
                <Typography sx={{ color: '#707070', fontSize: '0.875rem' }}>
                  Aucun membre pour le moment.
                </Typography>
              )}
            </List>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMembersProject(null)}>Fermer</Button>
        </DialogActions>
      </Dialog>
      </Box>
    </Box>
  )
}
