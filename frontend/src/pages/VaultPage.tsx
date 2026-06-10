import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box,
  Alert,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Typography,
  Button,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Snackbar,
  Breadcrumbs,
  Link,
} from '@mui/material'
import {
  Refresh as RefreshIcon,
  Folder as FolderIcon,
  Description as SecretIcon,
  Visibility as ViewIcon,
  Delete as DeleteIcon,
  Add as AddIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  Link as LinkIcon,
  ContentCopy as CopyIcon,
  VisibilityOff as VisibilityOffIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle, ModuleSecondaryText } from '../components/ModuleText'
import { vaultService, Secret } from '../services/vaultService'
import { kuraColors } from '../theme'

export default function VaultPage() {
  const queryClient = useQueryClient()
  const [configExpanded, setConfigExpanded] = useState(false)
  const [vaultAddr, setVaultAddr] = useState('')
  const [vaultToken, setVaultToken] = useState('')
  const [vaultMountPath, setVaultMountPath] = useState('')
  const [currentPath, setCurrentPath] = useState('')
  const [selectedSecret, setSelectedSecret] = useState<Secret | null>(null)
  const [secretDialogOpen, setSecretDialogOpen] = useState(false)
  const [revealedKeys, setRevealedKeys] = useState<Set<string>>(new Set())
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [newSecretPath, setNewSecretPath] = useState('')
  const [newSecretFields, setNewSecretFields] = useState([{ key: '', value: '' }])
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  const { data: status } = useQuery({
    queryKey: ['vault-status'],
    queryFn: () => vaultService.getStatus(),
    retry: false,
  })

  const { data: configData, refetch: refetchConfig } = useQuery({
    queryKey: ['vault-config'],
    queryFn: () => vaultService.getConfig(),
    retry: false,
  })

  const { data: secretsData, isLoading: secretsLoading, refetch: refetchSecrets } = useQuery({
    queryKey: ['vault-secrets', currentPath],
    queryFn: () => vaultService.listSecrets(currentPath),
    retry: false,
  })

  const saveConfigMutation = useMutation({
    mutationFn: (data: { vault_addr?: string; vault_token?: string; vault_mount_path?: string }) =>
      vaultService.setConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vault-config'] })
      queryClient.invalidateQueries({ queryKey: ['vault-status'] })
      queryClient.invalidateQueries({ queryKey: ['vault-secrets'] })
      refetchConfig()
      setVaultToken('')
      setConfigExpanded(false)
    },
  })

  const writeSecretMutation = useMutation({
    mutationFn: ({ path, data }: { path: string; data: Record<string, any> }) =>
      vaultService.writeSecret(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vault-secrets'] })
      setCreateDialogOpen(false)
      setNewSecretPath('')
      setNewSecretFields([{ key: '', value: '' }])
      setSnackbar({ open: true, message: 'Secret enregistré avec succès', severity: 'success' })
    },
    onError: () => {
      setSnackbar({ open: true, message: "Erreur lors de l'enregistrement du secret", severity: 'error' })
    },
  })

  const deleteSecretMutation = useMutation({
    mutationFn: (path: string) => vaultService.deleteSecret(path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vault-secrets'] })
      setDeleteTarget(null)
      setSnackbar({ open: true, message: 'Secret supprimé', severity: 'success' })
    },
    onError: () => {
      setSnackbar({ open: true, message: 'Erreur lors de la suppression du secret', severity: 'error' })
    },
  })

  const handleSaveConfig = () => {
    saveConfigMutation.mutate({
      ...(vaultAddr && { vault_addr: vaultAddr }),
      ...(vaultToken && { vault_token: vaultToken }),
      ...(vaultMountPath && { vault_mount_path: vaultMountPath }),
    })
  }

  const handleEntryClick = async (key: string) => {
    if (key.endsWith('/')) {
      setCurrentPath(currentPath ? `${currentPath}${key}` : key)
      return
    }
    const fullPath = currentPath ? `${currentPath}${key}` : key
    try {
      const secret = await vaultService.getSecret(fullPath)
      setSelectedSecret(secret)
      setRevealedKeys(new Set())
      setSecretDialogOpen(true)
    } catch {
      setSnackbar({ open: true, message: 'Impossible de lire ce secret', severity: 'error' })
    }
  }

  const breadcrumbParts = currentPath.split('/').filter(Boolean)

  const handleBreadcrumbClick = (index: number) => {
    if (index === -1) {
      setCurrentPath('')
      return
    }
    setCurrentPath(breadcrumbParts.slice(0, index + 1).join('/') + '/')
  }

  const toggleReveal = (key: string) => {
    setRevealedKeys((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const handleCopy = (value: string) => {
    navigator.clipboard.writeText(value)
    setSnackbar({ open: true, message: 'Copié dans le presse-papier', severity: 'success' })
  }

  const handleAddField = () => {
    setNewSecretFields([...newSecretFields, { key: '', value: '' }])
  }

  const handleFieldChange = (index: number, field: 'key' | 'value', value: string) => {
    const updated = [...newSecretFields]
    updated[index][field] = value
    setNewSecretFields(updated)
  }

  const handleRemoveField = (index: number) => {
    setNewSecretFields(newSecretFields.filter((_, i) => i !== index))
  }

  const handleCreateSecret = () => {
    if (!newSecretPath.trim()) {
      setSnackbar({ open: true, message: 'Le chemin du secret est requis', severity: 'error' })
      return
    }
    const data: Record<string, any> = {}
    for (const field of newSecretFields) {
      if (field.key.trim()) data[field.key.trim()] = field.value
    }
    if (Object.keys(data).length === 0) {
      setSnackbar({ open: true, message: 'Au moins une clé/valeur est requise', severity: 'error' })
      return
    }
    const fullPath = currentPath ? `${currentPath}${newSecretPath.trim()}` : newSecretPath.trim()
    writeSecretMutation.mutate({ path: fullPath, data })
  }

  const sealed = status?.sealed
  const configured = configData?.linked === 'true'

  return (
    <Box>
      <ModuleTitle>Vault</ModuleTitle>

      {/* Panneau de configuration Vault */}
      <ModuleCard sx={{ mb: 2 }}>
        <Box
          sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
          onClick={() => setConfigExpanded(!configExpanded)}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <LinkIcon sx={{ color: 'primary.main' }} />
            <Typography sx={{ fontWeight: 600 }}>Connexion Vault</Typography>
            {configured ? (
              <Chip label="Connecté" size="small" sx={{ color: '#00FF88', borderColor: '#00FF88' }} variant="outlined" />
            ) : (
              <Chip label="Non configuré" size="small" color="warning" variant="outlined" />
            )}
            {status && (
              <Chip
                label={sealed ? 'Scellé' : 'Déscellé'}
                size="small"
                sx={{
                  color: sealed ? kuraColors.error : kuraColors.success,
                  borderColor: sealed ? kuraColors.error : kuraColors.success,
                }}
                variant="outlined"
              />
            )}
          </Box>
          {configExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Box>

        {configExpanded && (
          <Box sx={{ px: 2, pb: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            {configData?.vault_addr && (
              <Alert severity="info" sx={{ mb: 1 }}>
                Adresse actuelle : <strong>{configData.vault_addr}</strong> — Mount path : <strong>{configData.vault_mount_path}</strong>
                {configured && ' — Token configuré ✓'}
              </Alert>
            )}
            <TextField
              label="Adresse Vault"
              placeholder="http://vault:8200"
              value={vaultAddr}
              onChange={(e) => setVaultAddr(e.target.value)}
              size="small"
              fullWidth
            />
            <TextField
              label="Token Vault"
              placeholder="Laisser vide pour conserver l'actuel"
              value={vaultToken}
              onChange={(e) => setVaultToken(e.target.value)}
              type="password"
              size="small"
              fullWidth
            />
            <TextField
              label="Mount path"
              placeholder="secret"
              value={vaultMountPath}
              onChange={(e) => setVaultMountPath(e.target.value)}
              size="small"
              sx={{ width: 200 }}
            />
            <Box>
              <Button
                variant="contained"
                onClick={handleSaveConfig}
                disabled={saveConfigMutation.isPending}
                sx={{ mr: 1 }}
              >
                {saveConfigMutation.isPending ? 'Connexion...' : 'Connecter'}
              </Button>
              {saveConfigMutation.isError && (
                <Typography color="error" variant="caption">Erreur lors de la connexion</Typography>
              )}
            </Box>
          </Box>
        )}
      </ModuleCard>

      <ModuleCard>
        <Box sx={{ p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <ModuleSubtitle>Secrets</ModuleSubtitle>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                variant="contained"
                size="small"
                startIcon={<AddIcon />}
                onClick={() => setCreateDialogOpen(true)}
              >
                Nouveau secret
              </Button>
              <IconButton onClick={() => refetchSecrets()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
          </Box>

          <Breadcrumbs sx={{ mb: 2 }}>
            <Link
              component="button"
              underline="hover"
              onClick={() => handleBreadcrumbClick(-1)}
              sx={{ color: currentPath ? kuraColors.text2 : kuraColors.text0, fontWeight: currentPath ? 400 : 600 }}
            >
              {configData?.vault_mount_path || 'secret'}
            </Link>
            {breadcrumbParts.map((part, idx) => (
              <Link
                key={idx}
                component="button"
                underline="hover"
                onClick={() => handleBreadcrumbClick(idx)}
                sx={{
                  color: idx === breadcrumbParts.length - 1 ? kuraColors.text0 : kuraColors.text2,
                  fontWeight: idx === breadcrumbParts.length - 1 ? 600 : 400,
                }}
              >
                {part}
              </Link>
            ))}
          </Breadcrumbs>

          {secretsLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (secretsData?.keys?.length ?? 0) > 0 ? (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Nom</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {secretsData!.keys.map((key) => {
                    const isFolder = key.endsWith('/')
                    return (
                      <TableRow
                        key={key}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() => handleEntryClick(key)}
                      >
                        <TableCell>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            {isFolder ? <FolderIcon sx={{ color: kuraColors.info, fontSize: 20 }} /> : <SecretIcon sx={{ color: kuraColors.text2, fontSize: 20 }} />}
                            {key}
                          </Box>
                        </TableCell>
                        <TableCell>
                          <Chip
                            label={isFolder ? 'Dossier' : 'Secret'}
                            size="small"
                            sx={{
                              fontSize: '0.6875rem',
                              bgcolor: isFolder ? `${kuraColors.info}22` : `${kuraColors.success}22`,
                              border: `1px solid ${isFolder ? kuraColors.info : kuraColors.success}`,
                              color: isFolder ? kuraColors.info : kuraColors.success,
                            }}
                          />
                        </TableCell>
                        <TableCell align="right">
                          {!isFolder && (
                            <>
                              <Tooltip title="Voir">
                                <IconButton size="small" onClick={(e) => { e.stopPropagation(); handleEntryClick(key) }}>
                                  <ViewIcon fontSize="small" />
                                </IconButton>
                              </Tooltip>
                              <Tooltip title="Supprimer">
                                <IconButton
                                  size="small"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    setDeleteTarget(currentPath ? `${currentPath}${key}` : key)
                                  }}
                                >
                                  <DeleteIcon fontSize="small" sx={{ color: kuraColors.error }} />
                                </IconButton>
                              </Tooltip>
                            </>
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <Alert severity="info">
              Aucun secret trouvé{currentPath ? ` dans ${currentPath}` : ''}.
            </Alert>
          )}
        </Box>
      </ModuleCard>

      {/* Dialog de visualisation d'un secret */}
      <Dialog open={secretDialogOpen} onClose={() => setSecretDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Secret : {selectedSecret?.path}</DialogTitle>
        <DialogContent>
          {selectedSecret && Object.entries(selectedSecret.data).map(([key, value]) => (
            <Box key={key} sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
              <TextField
                label={key}
                value={revealedKeys.has(key) ? String(value) : '••••••••'}
                size="small"
                fullWidth
                InputProps={{ readOnly: true, sx: { fontFamily: '"JetBrains Mono", monospace' } }}
              />
              <Tooltip title={revealedKeys.has(key) ? 'Masquer' : 'Révéler'}>
                <IconButton onClick={() => toggleReveal(key)}>
                  {revealedKeys.has(key) ? <VisibilityOffIcon fontSize="small" /> : <ViewIcon fontSize="small" />}
                </IconButton>
              </Tooltip>
              <Tooltip title="Copier">
                <IconButton onClick={() => handleCopy(String(value))}>
                  <CopyIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            </Box>
          ))}
          {selectedSecret?.metadata?.version != null && (
            <ModuleSecondaryText>Version : {selectedSecret.metadata.version}</ModuleSecondaryText>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSecretDialogOpen(false)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Dialog de création d'un secret */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Nouveau secret</DialogTitle>
        <DialogContent>
          <TextField
            label="Chemin du secret"
            placeholder="ex: api-keys/stripe"
            value={newSecretPath}
            onChange={(e) => setNewSecretPath(e.target.value)}
            size="small"
            fullWidth
            sx={{ mt: 1, mb: 2 }}
            helperText={currentPath ? `Sera créé sous ${currentPath}` : undefined}
          />
          {newSecretFields.map((field, idx) => (
            <Box key={idx} sx={{ display: 'flex', gap: 1, mb: 1.5 }}>
              <TextField
                label="Clé"
                value={field.key}
                onChange={(e) => handleFieldChange(idx, 'key', e.target.value)}
                size="small"
                sx={{ flex: 1 }}
              />
              <TextField
                label="Valeur"
                value={field.value}
                onChange={(e) => handleFieldChange(idx, 'value', e.target.value)}
                size="small"
                sx={{ flex: 1 }}
              />
              <IconButton onClick={() => handleRemoveField(idx)} disabled={newSecretFields.length === 1}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Box>
          ))}
          <Button size="small" onClick={handleAddField}>+ Ajouter une clé/valeur</Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Annuler</Button>
          <Button variant="contained" onClick={handleCreateSecret} disabled={writeSecretMutation.isPending}>
            {writeSecretMutation.isPending ? 'Enregistrement...' : 'Enregistrer'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Dialog de confirmation de suppression */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>Supprimer le secret ?</DialogTitle>
        <DialogContent>
          <Typography>
            Voulez-vous vraiment supprimer le secret <strong>{deleteTarget}</strong> ? Cette action est irréversible.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)}>Annuler</Button>
          <Button color="error" variant="contained" onClick={() => deleteTarget && deleteSecretMutation.mutate(deleteTarget)}>
            Supprimer
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        message={snackbar.message}
      />
    </Box>
  )
}
