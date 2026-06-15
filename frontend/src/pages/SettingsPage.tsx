import { useEffect, useState } from 'react'
import {
  Box,
  TextField,
  Grid,
  Alert,
  CircularProgress,
  Tabs,
  Tab,
  Button,
  IconButton,
  Chip,
  List,
  ListItem,
  ListItemText,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Typography,
} from '@mui/material'
import { Delete as DeleteIcon } from '@mui/icons-material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '../contexts/AuthContext'
import { useProject } from '../contexts/ProjectContext'
import { authService } from '../services/authService'
import { projectService, ProjectMapping, ProjectMember } from '../services/projectService'
import ModuleTitle from '../components/ModuleTitle'
import ModuleButton from '../components/ModuleButton'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle } from '../components/ModuleText'
import { kuraColors } from '../theme'

export default function SettingsPage() {
  const { user, refreshUser } = useAuth()
  const { currentProject, projects } = useProject()
  const queryClient = useQueryClient()

  const [name, setName] = useState(user?.name || '')
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const [selectedProjectId, setSelectedProjectId] = useState<string>(currentProject?.id || '')
  const [tab, setTab] = useState(0)
  const [newRepo, setNewRepo] = useState('')
  const [mappingError, setMappingError] = useState<string | null>(null)
  const [gitopsRepoInputs, setGitopsRepoInputs] = useState<Record<string, string>>({})
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberRole, setNewMemberRole] = useState<'admin' | 'member'>('member')
  const [memberError, setMemberError] = useState<string | null>(null)

  useEffect(() => {
    if (!selectedProjectId && currentProject?.id) {
      setSelectedProjectId(currentProject.id)
    }
  }, [currentProject?.id, selectedProjectId])

  const handleUpdateProfile = async () => {
    setLoading(true)
    setMessage(null)

    try {
      await authService.updateUser({ name })
      await refreshUser()
      setMessage({ type: 'success', text: 'Profil mis à jour avec succès' })
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Erreur lors de la mise à jour' })
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async () => {
    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: 'Les mots de passe ne correspondent pas' })
      return
    }

    if (newPassword.length < 6) {
      setMessage({ type: 'error', text: 'Le mot de passe doit contenir au moins 6 caractères' })
      return
    }

    setLoading(true)
    setMessage(null)

    try {
      await authService.changePassword(currentPassword, newPassword)
      setMessage({ type: 'success', text: 'Mot de passe modifié avec succès' })
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Erreur lors du changement de mot de passe' })
    } finally {
      setLoading(false)
    }
  }

  const { data: mappingsData, isLoading: mappingsLoading } = useQuery({
    queryKey: ['project-mappings', selectedProjectId],
    queryFn: () => projectService.listMappings(selectedProjectId),
    enabled: !!selectedProjectId,
  })

  const createMappingMutation = useMutation({
    mutationFn: (repo: string) => projectService.createMapping(selectedProjectId, { forgejo_repository: repo }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', selectedProjectId] })
      setNewRepo('')
      setMappingError(null)
    },
    onError: (err: any) => {
      setMappingError(err.response?.data?.error || err.message || "Erreur lors de l'ajout du dépôt")
    },
  })

  const deleteMappingMutation = useMutation({
    mutationFn: (mappingId: string) => projectService.deleteMapping(selectedProjectId, mappingId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', selectedProjectId] })
    },
  })

  const setGitOpsRepositoryMutation = useMutation({
    mutationFn: ({ mappingId, repo }: { mappingId: string; repo: string }) =>
      projectService.setMappingGitOpsRepository(selectedProjectId, mappingId, repo),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-mappings', selectedProjectId] })
    },
  })

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['project-members', selectedProjectId],
    queryFn: () => projectService.getProjectMembers(selectedProjectId),
    enabled: !!selectedProjectId,
  })

  const addMemberMutation = useMutation({
    mutationFn: (data: { email: string; role: 'admin' | 'member' }) =>
      projectService.addProjectMember(selectedProjectId, { email: data.email, role: data.role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', selectedProjectId] })
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
      projectService.updateProjectMember(selectedProjectId, userId, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', selectedProjectId] })
    },
    onError: (err: any) => {
      setMemberError(err.response?.data?.error || err.message || 'Erreur lors de la mise à jour du membre')
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => projectService.removeProjectMember(selectedProjectId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-members', selectedProjectId] })
    },
    onError: (err: any) => {
      setMemberError(err.response?.data?.error || err.message || 'Erreur lors de la suppression du membre')
    },
  })

  const handleAddRepo = () => {
    const repo = newRepo.trim()
    if (!repo) return
    if (!/^[\w.-]+\/[\w.-]+$/.test(repo)) {
      setMappingError('Format attendu : owner/repo')
      return
    }
    createMappingMutation.mutate(repo)
  }

  const handleInviteMember = () => {
    const email = newMemberEmail.trim()
    if (!email) {
      setMemberError("L'email est requis")
      return
    }
    addMemberMutation.mutate({ email, role: newMemberRole })
  }

  return (
    <Box>
      <ModuleTitle>Paramètres</ModuleTitle>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <ModuleCard>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Profil
            </ModuleSubtitle>
              {message && (
                <Alert severity={message.type} sx={{ mb: 2 }}>
                  {message.text}
                </Alert>
              )}
              <TextField
                fullWidth
                label="Email"
                value={user?.email || ''}
                disabled
                sx={{ mb: 2 }}
              />
              <TextField
                fullWidth
                label="Nom"
                value={name}
                onChange={(e) => setName(e.target.value)}
                sx={{ mb: 2 }}
              />
              <ModuleButton
                onClick={handleUpdateProfile}
                disabled={loading}
              >
                {loading ? <CircularProgress size={24} /> : 'Mettre à jour le profil'}
              </ModuleButton>
          </ModuleCard>
        </Grid>

        <Grid item xs={12} md={6}>
          <ModuleCard>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Changer le mot de passe
            </ModuleSubtitle>
            <TextField
              fullWidth
              label="Mot de passe actuel"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Nouveau mot de passe"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Confirmer le nouveau mot de passe"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <ModuleButton
              onClick={handleChangePassword}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : 'Changer le mot de passe'}
            </ModuleButton>
          </ModuleCard>
        </Grid>

        <Grid item xs={12}>
          <ModuleCard>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2, mb: 1 }}>
              <ModuleSubtitle sx={{ mb: 0 }}>
                Configuration du projet
              </ModuleSubtitle>
              <FormControl size="small" sx={{ minWidth: 220 }}>
                <InputLabel id="settings-project-label">Projet</InputLabel>
                <Select
                  labelId="settings-project-label"
                  label="Projet"
                  value={selectedProjectId}
                  onChange={(e) => setSelectedProjectId(e.target.value)}
                >
                  {projects.map((p) => (
                    <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Box>

            {!selectedProjectId ? (
              <Typography sx={{ color: '#707070', fontSize: '0.875rem', mt: 2 }}>
                Sélectionnez un projet pour gérer ses dépôts et ses membres.
              </Typography>
            ) : (
              <>
                <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3, borderBottom: `1px solid ${kuraColors.border1}` }}>
                  <Tab label="Dépôts & GitOps" />
                  <Tab label="Membres" />
                </Tabs>

                {tab === 0 && (
                  <Box>
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
                  </Box>
                )}

                {tab === 1 && (
                  <Box>
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
                                  <IconButton
                                    size="small"
                                    edge="end"
                                    disabled={isOwner}
                                    onClick={() => removeMemberMutation.mutate(m.user_id)}
                                    sx={{ color: kuraColors.error }}
                                  >
                                    <DeleteIcon fontSize="small" />
                                  </IconButton>
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
                  </Box>
                )}
              </>
            )}
          </ModuleCard>
        </Grid>
      </Grid>
    </Box>
  )
}
