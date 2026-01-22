import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  CircularProgress,
  Alert,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Tooltip,
  Snackbar,
  List,
  ListItem,
  ListItemText,
  Divider,
  Tabs,
  Tab,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  FormControlLabel,
  Switch,
} from '@mui/material'
import {
  CloudUpload as CloudUploadIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Warning as WarningIcon,
  Sync as SyncIcon,
  CloudQueue as CloudQueueIcon,
  Edit as EditIcon,
} from '@mui/icons-material'
import { terraformService, TerraformState, terraformSourceService, TerraformDriftResult } from '../services/terraformService'

export default function TerraformPage() {
  const [activeTab, setActiveTab] = useState(0)
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [stateName, setStateName] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [selectedState, setSelectedState] = useState<TerraformState | null>(null)
  const [detailDialogOpen, setDetailDialogOpen] = useState(false)
  const [driftResults, setDriftResults] = useState<any[]>([])
  const [driftDialogOpen, setDriftDialogOpen] = useState(false)
  const [detectingDrift, setDetectingDrift] = useState(false)
  const [selectedResource, setSelectedResource] = useState<any>(null)
  const [resourceDialogOpen, setResourceDialogOpen] = useState(false)
  const [sourceDialogOpen, setSourceDialogOpen] = useState(false)
  const [editingSource, setEditingSource] = useState<any>(null)
  const [pendingEditSource, setPendingEditSource] = useState<any>(null)
  const [sourceType, setSourceType] = useState<'s3' | 'azure' | 'gcp'>('gcp')
  const [gcpConfig, setGcpConfig] = useState({
    bucket: '',
    object_name: '',
    credentials_json: '',
    sync_interval: '15m',
    auto_sync: true,
  })
  const [s3Config, setS3Config] = useState({
    bucket: '',
    key: '',
    region: 'us-east-1',
    endpoint: '',
    aws_access_key_id: '',
    aws_secret_access_key: '',
    sync_interval: '15m',
    auto_sync: true,
  })
  const [azureConfig, setAzureConfig] = useState({
    account_name: '',
    container: '',
    blob_name: '',
    account_key: '',
    connection_string: '',
    sync_interval: '15m',
    auto_sync: true,
  })
  const [selectedStateForSource, setSelectedStateForSource] = useState<string>('')
  const [createStateFromSource, setCreateStateFromSource] = useState(false)
  const [newStateName, setNewStateName] = useState('')
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })
  const queryClient = useQueryClient()

  const { data: states, isLoading, error, refetch } = useQuery({
    queryKey: ['terraform-states'],
    queryFn: async () => {
      const statesData = await terraformService.getStates()
      // Pour chaque état, charger le résumé si les ressources ne sont pas disponibles
      const statesWithCounts = await Promise.all(
        statesData.items.map(async (state) => {
          // Si les ressources ne sont pas chargées, utiliser le résumé pour obtenir le nombre
          if (!state.state?.resources || state.state.resources.length === 0) {
            try {
              const summary = await terraformService.getStateSummary(state.id)
              // Retourner l'état avec le nombre de ressources depuis le résumé
              return {
                ...state,
                _resourceCount: summary.resource_count, // Stocker le nombre pour l'affichage
              }
            } catch (err) {
              // Si le résumé n'est pas disponible, utiliser 0
              return {
                ...state,
                _resourceCount: 0,
              }
            }
          }
          return state
        })
      )
      return { items: statesWithCounts }
    },
  })

  const { data: sources, isLoading: sourcesLoading, refetch: refetchSources } = useQuery({
    queryKey: ['terraform-sources'],
    queryFn: () => terraformSourceService.getSources(),
  })

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFile) throw new Error('Aucun fichier sélectionné')
      const name = stateName || selectedFile.name
      return await terraformService.uploadState(name, selectedFile)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['terraform-states'] })
      setUploadDialogOpen(false)
      setStateName('')
      setSelectedFile(null)
      setSnackbar({ open: true, message: 'État Terraform uploadé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || 'Erreur lors de l\'upload de l\'état Terraform',
        severity: 'error',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => terraformService.deleteState(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['terraform-states'] })
      setSnackbar({ open: true, message: 'État Terraform supprimé avec succès', severity: 'success' })
    },
    onError: () => {
      setSnackbar({ open: true, message: 'Erreur lors de la suppression de l\'état', severity: 'error' })
    },
  })

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedFile(event.target.files[0])
      if (!stateName) {
        setStateName(event.target.files[0].name.replace('.tfstate', ''))
      }
    }
  }

  const handleUpload = () => {
    uploadMutation.mutate()
  }

  const handleDelete = (id: string) => {
    if (window.confirm('Êtes-vous sûr de vouloir supprimer cet état Terraform ?')) {
      deleteMutation.mutate(id)
    }
  }

  const handleViewDetails = async (state: TerraformState) => {
    try {
      const fullState = await terraformService.getState(state.id)
      setSelectedState(fullState)
      setDetailDialogOpen(true)
    } catch (error) {
      setSnackbar({ open: true, message: 'Erreur lors du chargement des détails', severity: 'error' })
    }
  }

  const handleAddSource = () => {
    setEditingSource(null)
    setSourceDialogOpen(true)
  }

  const handleEditSource = (source: any) => {
    // Stocker la source à éditer pour que useEffect la traite
    setPendingEditSource(source)
  }

  // Fonction pour fermer le dialog et réinitialiser tous les états
  const handleCloseDialog = () => {
    setSourceDialogOpen(false)
    setEditingSource(null)
    setPendingEditSource(null)
    // Réinitialiser les configurations
    setGcpConfig({
      bucket: '',
      object_name: '',
      credentials_json: '',
      sync_interval: '15m',
      auto_sync: true,
    })
    setS3Config({
      bucket: '',
      key: '',
      region: 'us-east-1',
      endpoint: '',
      aws_access_key_id: '',
      aws_secret_access_key: '',
      sync_interval: '15m',
      auto_sync: true,
    })
    setAzureConfig({
      account_name: '',
      container: '',
      blob_name: '',
      account_key: '',
      connection_string: '',
      sync_interval: '15m',
      auto_sync: true,
    })
    setCreateStateFromSource(false)
    setSelectedStateForSource('')
    setNewStateName('')
    setSourceType('gcp')
  }

  // useEffect pour gérer l'ouverture du dialog après la mise à jour des états
  useEffect(() => {
    if (pendingEditSource) {
      const source = pendingEditSource
      // Configurer les états selon le type de source
      if (source.type === 'gcp') {
        setGcpConfig({
          bucket: source.config.gcp_bucket || '',
          object_name: source.config.gcp_object_name || '',
          credentials_json: '', // Ne pas pré-remplir pour sécurité
          sync_interval: source.config.sync_interval || '15m',
          auto_sync: source.config.auto_sync !== false,
        })
      } else if (source.type === 's3') {
        setS3Config({
          bucket: source.config.s3_bucket || '',
          key: source.config.s3_key || '',
          region: source.config.s3_region || 'us-east-1',
          endpoint: source.config.s3_endpoint || '',
          aws_access_key_id: '', // Ne pas pré-remplir
          aws_secret_access_key: '', // Ne pas pré-remplir
          sync_interval: source.config.sync_interval || '15m',
          auto_sync: source.config.auto_sync !== false,
        })
      } else if (source.type === 'azure') {
        setAzureConfig({
          account_name: source.config.azure_account_name || '',
          container: source.config.azure_container || '',
          blob_name: source.config.azure_blob_name || '',
          account_key: '', // Ne pas pré-remplir
          connection_string: '', // Ne pas pré-remplir
          sync_interval: source.config.sync_interval || '15m',
          auto_sync: source.config.auto_sync !== false,
        })
      }
      setCreateStateFromSource(false)
      setSelectedStateForSource(source.state_file_id || '')
      setSourceType(source.type)
      setEditingSource(source)
      setSourceDialogOpen(true)
      setPendingEditSource(null) // Réinitialiser
    }
  }, [pendingEditSource])

  const addSourceMutation = useMutation({
    mutationFn: async () => {
      let stateFileID = selectedStateForSource
      if (createStateFromSource) {
        if (!newStateName) {
          throw new Error('Veuillez saisir un nom pour le nouvel état')
        }
        stateFileID = `temp-${newStateName}-${Date.now()}`
      } else {
        if (!selectedStateForSource) {
          throw new Error('Veuillez sélectionner un état Terraform ou créer un nouvel état')
        }
      }
      let sourceConfig: any = {}
      if (sourceType === 's3') {
        if (!s3Config.bucket || !s3Config.key) {
          throw new Error('Bucket et clé S3 requis')
        }
        sourceConfig = {
          s3_bucket: s3Config.bucket,
          s3_key: s3Config.key,
          s3_region: s3Config.region,
          s3_endpoint: s3Config.endpoint || undefined,
          aws_access_key_id: s3Config.aws_access_key_id || undefined,
          aws_secret_access_key: s3Config.aws_secret_access_key || undefined,
          sync_interval: s3Config.sync_interval,
          auto_sync: s3Config.auto_sync,
        }
      } else if (sourceType === 'gcp') {
        if (!gcpConfig.bucket || !gcpConfig.object_name) {
          throw new Error('Bucket et nom d\'objet GCP requis')
        }
        if (!editingSource && !gcpConfig.credentials_json) {
          throw new Error('Credentials JSON GCP requis (Service Account)')
        }
        sourceConfig = {
          gcp_bucket: gcpConfig.bucket,
          gcp_object_name: gcpConfig.object_name,
          gcp_credentials_json: gcpConfig.credentials_json || undefined,
          sync_interval: gcpConfig.sync_interval,
          auto_sync: gcpConfig.auto_sync,
        }
      } else if (sourceType === 'azure') {
        if (!azureConfig.account_name || !azureConfig.container || !azureConfig.blob_name) {
          throw new Error('Nom de compte, container et nom de blob Azure requis')
        }
        sourceConfig = {
          azure_account_name: azureConfig.account_name,
          azure_container: azureConfig.container,
          azure_blob_name: azureConfig.blob_name,
          azure_account_key: azureConfig.account_key || undefined,
          azure_connection_string: azureConfig.connection_string || undefined,
          sync_interval: azureConfig.sync_interval,
          auto_sync: azureConfig.auto_sync,
        }
      }
      if (editingSource) {
        return await terraformSourceService.updateSource(editingSource.id, {
          state_file_id: stateFileID,
          type: sourceType,
          config: sourceConfig,
          enabled: editingSource.enabled !== false,
        })
      } else {
        return await terraformSourceService.addSource({
          state_file_id: stateFileID,
          type: sourceType,
          config: sourceConfig,
          enabled: true,
        })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['terraform-sources'] })
      queryClient.invalidateQueries({ queryKey: ['terraform-states'] })
      setSourceDialogOpen(false)
      setEditingSource(null)
      setSnackbar({ open: true, message: `Source ${sourceType.toUpperCase()} ${editingSource ? 'modifiée' : 'ajoutée'} avec succès`, severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || (editingSource ? 'Erreur lors de la modification de la source' : 'Erreur lors de l\'ajout de la source'),
        severity: 'error',
      })
    },
  })

  const syncSourceMutation = useMutation({
    mutationFn: (sourceId: string) => terraformSourceService.syncSource(sourceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['terraform-states'] })
      queryClient.invalidateQueries({ queryKey: ['terraform-sources'] })
      setSnackbar({ open: true, message: 'Synchronisation lancée', severity: 'success' })
    },
    onError: () => {
      setSnackbar({ open: true, message: 'Erreur lors de la synchronisation', severity: 'error' })
    },
  })

  const deleteSourceMutation = useMutation({
    mutationFn: (sourceId: string) => terraformSourceService.deleteSource(sourceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['terraform-sources'] })
      setSnackbar({ open: true, message: 'Source supprimée avec succès', severity: 'success' })
    },
    onError: () => {
      setSnackbar({ open: true, message: 'Erreur lors de la suppression de la source', severity: 'error' })
    },
  })

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          Erreur lors du chargement des états Terraform : {error instanceof Error ? error.message : 'Erreur inconnue'}
        </Alert>
        <Button onClick={() => refetch()} startIcon={<RefreshIcon />}>
          Réessayer
        </Button>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 0 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 5 }}>
        <Typography 
          variant="h3" 
          component="h1"
          sx={{
            fontWeight: 600,
            background: 'linear-gradient(135deg, #00E5FF, #B388FF)',
            backgroundClip: 'text',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            fontFamily: '"Inter", sans-serif',
            letterSpacing: '0.02em',
          }}
        >
          Terraform
        </Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<CloudUploadIcon />}
            onClick={() => setUploadDialogOpen(true)}
            sx={{ mr: 1 }}
          >
            Uploader un état
          </Button>
          <Button
            variant="outlined"
            startIcon={<CloudQueueIcon />}
            onClick={handleAddSource}
            sx={{ mr: 1 }}
          >
            Lier une source cloud
          </Button>
          <Tooltip title="Actualiser">
            <IconButton onClick={() => { refetch(); refetchSources() }} disabled={isLoading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)} sx={{ mb: 4 }}>
        <Tab label="États" />
        <Tab label="Sources de synchronisation" />
      </Tabs>

      {activeTab === 0 && (
        <>

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : !states || states.items.length === 0 ? (
        <Card sx={{ animation: 'jellyfishFloat 12s ease-in-out infinite' }}>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              Aucun état Terraform
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Commencez par uploader un fichier tfstate pour voir vos ressources Terraform.
            </Typography>
            <Button variant="contained" startIcon={<CloudUploadIcon />} onClick={() => setUploadDialogOpen(true)}>
              Uploader un état Terraform
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Grid container spacing={3}>
          {states.items.map((state, idx) => (
            <Grid item xs={12} md={6} lg={4} key={state.id}>
              <Card 
                sx={{ 
                  height: '100%', 
                  display: 'flex', 
                  flexDirection: 'column',
                  '--card-delay': `${idx * 0.3}s`,
                }}
              >
                <CardContent sx={{ flexGrow: 1 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', mb: 2.5 }}>
                    <Typography 
                      variant="h6" 
                      component="div" 
                      sx={{ 
                        fontWeight: 500, 
                        color: 'rgba(255, 255, 255, 0.95)',
                        fontFamily: '"Inter", sans-serif',
                      }}
                    >
                      {state.name}
                    </Typography>
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() => handleDelete(state.id)}
                      sx={{ ml: 1 }}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                  <Box sx={{ mb: 2 }}>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                      Version: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{state.state?.version || 'N/A'}</span>
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                      Ressources: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{(state as any)._resourceCount !== undefined 
                        ? (state as any)._resourceCount 
                        : (state.state?.resources?.length || 0)}</span>
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                      Sorties: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{Object.keys(state.state?.outputs || {}).length}</span>
                    </Typography>
                    <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 1.5, fontSize: '0.75rem' }}>
                      {new Date(state.uploaded_at).toLocaleString('fr-FR')}
                    </Typography>
                  </Box>
                  <Box sx={{ mt: 'auto', display: 'flex', gap: 1.5, pt: 2.5 }}>
                    <Button
                      size="small"
                      variant="outlined"
                      onClick={() => handleViewDetails(state)}
                    >
                      Détails
                    </Button>
                    <Button
                      size="small"
                      variant="outlined"
                      color="warning"
                      startIcon={<WarningIcon />}
                      onClick={async () => {
                        setDetectingDrift(true)
                        try {
                          const result = await terraformService.detectDrift(state.id)
                          setDriftResults(result.items || [])
                          setSelectedState(state)
                          setDriftDialogOpen(true)
                        } catch (error) {
                          setSnackbar({ open: true, message: 'Erreur lors de la détection de drift', severity: 'error' })
                        } finally {
                          setDetectingDrift(false)
                        }
                      }}
                      disabled={detectingDrift}
                    >
                      {detectingDrift ? 'Vérification...' : 'Drift'}
                    </Button>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Dialog d'upload */}
      <Dialog open={uploadDialogOpen} onClose={() => setUploadDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Uploader un état Terraform</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Nom de l'état"
            value={stateName}
            onChange={(e) => setStateName(e.target.value)}
            margin="normal"
            placeholder="Ex: production.tfstate"
          />
          <Button
            variant="outlined"
            component="label"
            fullWidth
            startIcon={<CloudUploadIcon />}
            sx={{ mt: 2 }}
          >
            {selectedFile ? selectedFile.name : 'Sélectionner un fichier tfstate'}
            <input type="file" hidden accept=".tfstate,application/json" onChange={handleFileChange} />
          </Button>
          {selectedFile && (
            <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
              Fichier sélectionné : {selectedFile.name} ({(selectedFile.size / 1024).toFixed(2)} KB)
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUploadDialogOpen(false)}>Annuler</Button>
          <Button
            onClick={handleUpload}
            variant="contained"
            disabled={!selectedFile || uploadMutation.isPending}
            startIcon={uploadMutation.isPending ? <CircularProgress size={20} /> : <CloudUploadIcon />}
          >
            Uploader
          </Button>
        </DialogActions>
      </Dialog>

      {/* Dialog de détails */}
      <Dialog
        open={detailDialogOpen}
        onClose={() => setDetailDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Détails de l'état : {selectedState?.name}</DialogTitle>
        <DialogContent>
          {selectedState && (
            <Box>
              <Typography variant="subtitle1" gutterBottom sx={{ mt: 2 }}>
                Informations générales
              </Typography>
              <List>
                <ListItem>
                  <ListItemText primary="Version" secondary={selectedState.state?.version} />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Version Terraform" secondary={selectedState.state?.terraform_version || 'N/A'} />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Serial" secondary={selectedState.state?.serial} />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Nombre de ressources" secondary={selectedState.state?.resources?.length || 0} />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Nombre de sorties" secondary={Object.keys(selectedState.state?.outputs || {}).length} />
                </ListItem>
              </List>
              <Divider sx={{ my: 2 }} />
              <Typography variant="subtitle1" gutterBottom>
                Ressources ({selectedState.state?.resources?.length || 0})
              </Typography>
              {selectedState.state?.resources && selectedState.state.resources.length > 0 ? (
                <TableContainer component={Paper} sx={{ maxHeight: 400, mt: 1 }}>
                  <Table size="small" stickyHeader>
                    <TableHead>
                      <TableRow>
                        <TableCell sx={{ width: 60 }}>#</TableCell>
                        <TableCell>Type</TableCell>
                        <TableCell>Nom</TableCell>
                        <TableCell>Provider</TableCell>
                        <TableCell>Mode</TableCell>
                        <TableCell>Module</TableCell>
                        <TableCell>Instances</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {selectedState.state.resources.map((resource, idx) => (
                        <TableRow 
                          key={idx}
                          hover
                          sx={{ cursor: 'pointer' }}
                          onClick={() => {
                            setSelectedResource(resource)
                            setResourceDialogOpen(true)
                          }}
                        >
                          <TableCell sx={{ fontWeight: 'bold', color: 'text.secondary' }}>{idx + 1}</TableCell>
                          <TableCell>
                            <Typography variant="body2" fontFamily="monospace" fontWeight="bold">
                              {resource.type}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2" fontFamily="monospace">
                              {resource.name}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="caption" color="text.secondary">
                              {resource.provider}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip label={resource.mode} size="small" color={resource.mode === 'managed' ? 'primary' : 'secondary'} />
                          </TableCell>
                          <TableCell>
                            <Typography variant="caption" color="text.secondary">
                              {resource.module || '-'}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip 
                              label={resource.instances?.length || 0} 
                              size="small" 
                              variant="outlined"
                              color="default"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  Aucune ressource
                </Typography>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailDialogOpen(false)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Dialog des résultats de drift */}
      <Dialog open={driftDialogOpen} onClose={() => setDriftDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>
          Résultats de détection de drift : {selectedState?.name}
        </DialogTitle>
        <DialogContent>
          {driftResults.length === 0 ? (
            <Alert severity="info" sx={{ mt: 2 }}>
              Aucun résultat de drift disponible. La détection compare le tfstate avec l'état réel de l'infrastructure.
            </Alert>
          ) : (
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {driftResults.length} ressource(s) analysée(s)
              </Typography>
              <TableContainer component={Paper} sx={{ maxHeight: 500, mt: 2 }}>
                <Table size="small" stickyHeader>
                  <TableHead>
                    <TableRow>
                      <TableCell>Ressource</TableCell>
                      <TableCell>Type</TableCell>
                      <TableCell>Statut</TableCell>
                      <TableCell>Message</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {driftResults.map((result: TerraformDriftResult, idx: number) => (
                      <TableRow key={idx}>
                        <TableCell>
                          <Typography variant="body2" fontFamily="monospace">
                            {result.resource_address}
                          </Typography>
                        </TableCell>
                        <TableCell>{result.resource_type}</TableCell>
                        <TableCell>
                          <Chip
                            label={result.status}
                            size="small"
                            color={
                              result.status === 'in_sync'
                                ? 'success'
                                : result.status === 'drifted' || result.status === 'missing'
                                ? 'error'
                                : 'warning'
                            }
                          />
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary">
                            {result.message || 'Aucun message'}
                          </Typography>
                          {result.differences && result.differences.length > 0 && (
                            <Box sx={{ mt: 1 }}>
                              <Typography variant="caption" color="error" fontWeight="bold">
                                {result.differences.length} différence(s) détectée(s)
                              </Typography>
                              {result.differences.slice(0, 3).map((diff, diffIdx) => (
                                <Typography key={diffIdx} variant="caption" display="block" sx={{ fontFamily: 'monospace', fontSize: '0.7rem' }}>
                                  • {diff.attribute}: attendu {JSON.stringify(diff.expected)}, actuel {JSON.stringify(diff.actual)}
                                </Typography>
                              ))}
                            </Box>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
              <Alert severity="warning" sx={{ mt: 2 }}>
                Note : La détection de drift utilise les APIs réelles des providers cloud (GCP, AWS, Azure) pour comparer l'état réel avec le tfstate.
              </Alert>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDriftDialogOpen(false)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Dialog de détails d'une ressource Terraform */}
      <Dialog 
        open={resourceDialogOpen} 
        onClose={() => { setResourceDialogOpen(false); setSelectedResource(null) }} 
        maxWidth="lg" 
        fullWidth
      >
        <DialogTitle>
          Détails de la ressource : {selectedResource?.type}.{selectedResource?.name}
        </DialogTitle>
        <DialogContent>
          {selectedResource && (
            <Box>
              <Typography variant="subtitle1" gutterBottom sx={{ mt: 2 }}>
                Informations de base
              </Typography>
              <List>
                <ListItem>
                  <ListItemText 
                    primary="Type de ressource" 
                    secondary={
                      <Typography variant="body2" fontFamily="monospace" fontWeight="bold">
                        {selectedResource.type}
                      </Typography>
                    } 
                  />
                </ListItem>
                <ListItem>
                  <ListItemText 
                    primary="Nom" 
                    secondary={
                      <Typography variant="body2" fontFamily="monospace">
                        {selectedResource.name}
                      </Typography>
                    } 
                  />
                </ListItem>
                <ListItem>
                  <ListItemText 
                    primary="Provider" 
                    secondary={selectedResource.provider} 
                  />
                </ListItem>
                <ListItem>
                  <ListItemText 
                    primary="Mode" 
                    secondary={
                      <Chip 
                        label={selectedResource.mode} 
                        size="small" 
                        color={selectedResource.mode === 'managed' ? 'primary' : 'secondary'} 
                      />
                    } 
                  />
                </ListItem>
                {selectedResource.module && (
                  <ListItem>
                    <ListItemText 
                      primary="Module" 
                      secondary={
                        <Typography variant="body2" fontFamily="monospace">
                          {selectedResource.module}
                        </Typography>
                      } 
                    />
                  </ListItem>
                )}
                <ListItem>
                  <ListItemText 
                    primary="Nombre d'instances" 
                    secondary={selectedResource.instances?.length || 0} 
                  />
                </ListItem>
              </List>

              {selectedResource.instances && selectedResource.instances.length > 0 && (
                <>
                  <Divider sx={{ my: 2 }} />
                  <Typography variant="subtitle1" gutterBottom>
                    Instances ({selectedResource.instances.length})
                  </Typography>
                  {selectedResource.instances.map((instance: any, instanceIdx: number) => (
                    <Box key={instanceIdx} sx={{ mb: 3 }}>
                      <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                        Instance {instanceIdx + 1}
                        {instance.schema_version && (
                          <Chip 
                            label={`Schema v${instance.schema_version}`} 
                            size="small" 
                            variant="outlined"
                            sx={{ ml: 1 }}
                          />
                        )}
                      </Typography>
                      
                      {instance.attributes && Object.keys(instance.attributes).length > 0 && (
                        <Box>
                          <Typography variant="body2" color="text.secondary" gutterBottom>
                            Attributs ({Object.keys(instance.attributes).length})
                          </Typography>
                          <TableContainer component={Paper} sx={{ maxHeight: 300, mt: 1 }}>
                            <Table size="small">
                              <TableHead>
                                <TableRow>
                                  <TableCell>Attribut</TableCell>
                                  <TableCell>Valeur</TableCell>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {Object.entries(instance.attributes).slice(0, 20).map(([key, value]) => (
                                  <TableRow key={key}>
                                    <TableCell>
                                      <Typography variant="body2" fontFamily="monospace" fontWeight="bold">
                                        {key}
                                      </Typography>
                                    </TableCell>
                                    <TableCell>
                                      <Typography variant="body2" fontFamily="monospace">
                                        {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
                                      </Typography>
                                    </TableCell>
                                  </TableRow>
                                ))}
                                {Object.keys(instance.attributes).length > 20 && (
                                  <TableRow>
                                    <TableCell colSpan={2} align="center">
                                      <Typography variant="caption" color="text.secondary">
                                        ... et {Object.keys(instance.attributes).length - 20} autres attributs
                                      </Typography>
                                    </TableCell>
                                  </TableRow>
                                )}
                              </TableBody>
                            </Table>
                          </TableContainer>
                        </Box>
                      )}

                      {instance.dependencies && (
                        <Box sx={{ mt: 2 }}>
                          <Typography variant="body2" color="text.secondary" gutterBottom>
                            Dépendances
                          </Typography>
                          <Typography variant="body2" fontFamily="monospace" sx={{ bgcolor: 'background.paper', p: 1, borderRadius: 1 }}>
                            {typeof instance.dependencies === 'string' 
                              ? instance.dependencies 
                              : JSON.stringify(instance.dependencies, null, 2)}
                          </Typography>
                        </Box>
                      )}
                    </Box>
                  ))}
                </>
              )}

              <Divider sx={{ my: 2 }} />
              <Typography variant="subtitle1" gutterBottom>
                Bloc Terraform
              </Typography>
              <Box sx={{ bgcolor: 'background.paper', p: 2, borderRadius: 1, border: '1px solid', borderColor: 'divider' }}>
                <Typography variant="body2" fontFamily="monospace" component="pre" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                  {(() => {
                    const instances = selectedResource.instances || []
                    const instanceBlocks = instances.map((instance: any, idx: number) => {
                      if (!instance.attributes) return ''
                      const attrs = Object.entries(instance.attributes)
                        .filter(([key]) => !key.startsWith('_'))
                        .map(([key, value]) => {
                          if (typeof value === 'object' && value !== null) {
                            return `  ${key} = ${JSON.stringify(value, null, 2).split('\n').map((line: string, i: number) => i === 0 ? line : '  ' + line).join('\n')}`
                          }
                          if (typeof value === 'string' && value.includes('\n')) {
                            return `  ${key} = <<-EOT\n${value}\nEOT`
                          }
                          return `  ${key} = ${JSON.stringify(value)}`
                        })
                        .join('\n')
                      return attrs ? (instances.length > 1 ? `  # Instance ${idx + 1}\n${attrs}` : attrs) : ''
                    }).filter(Boolean).join('\n\n')
                    
                    const moduleLine = selectedResource.module ? `  module = "${selectedResource.module}"\n` : ''
                    return `resource "${selectedResource.type}" "${selectedResource.name}" {
${moduleLine}${instanceBlocks}
}`
                  })()}
                </Typography>
              </Box>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setResourceDialogOpen(false); setSelectedResource(null) }}>Fermer</Button>
        </DialogActions>
      </Dialog>
      </>
      )}

      {/* Dialog d'ajout/modification de source cloud - Déplacé en dehors des onglets pour être toujours disponible */}
      <Dialog 
        key={editingSource?.id || 'new'} 
        open={sourceDialogOpen} 
        onClose={handleCloseDialog}
        maxWidth="md" 
        fullWidth
      >
        <DialogTitle>{editingSource ? 'Modifier une source cloud' : 'Lier une source cloud (S3, Azure, GCP)'}</DialogTitle>
        <DialogContent>
          <FormControl fullWidth margin="normal">
            <InputLabel>Type de source</InputLabel>
            <Select
              value={sourceType}
              onChange={(e) => setSourceType(e.target.value as 's3' | 'azure' | 'gcp')}
              label="Type de source"
              disabled={!!editingSource}
            >
              <MenuItem value="gcp">GCP Cloud Storage</MenuItem>
              <MenuItem value="s3">AWS S3</MenuItem>
              <MenuItem value="azure">Azure Blob Storage</MenuItem>
            </Select>
          </FormControl>

          {!editingSource && (
            <FormControlLabel
              control={
                <Switch
                  checked={createStateFromSource}
                  onChange={(e) => setCreateStateFromSource(e.target.checked)}
                />
              }
              label="Créer un nouvel état depuis cette source"
              sx={{ mt: 2, mb: 2 }}
            />
          )}

          {createStateFromSource && !editingSource ? (
            <TextField
              fullWidth
              label="Nom du nouvel état"
              value={newStateName}
              onChange={(e) => setNewStateName(e.target.value)}
              margin="normal"
              required
              placeholder="Ex: production-gcp"
            />
          ) : !editingSource ? (
            <TextField
              fullWidth
              select
              label="État Terraform existant"
              value={selectedStateForSource}
              onChange={(e) => setSelectedStateForSource(e.target.value)}
              margin="normal"
              SelectProps={{
                native: true,
              }}
            >
              <option value="">Sélectionner un état...</option>
              {states?.items.map((state) => (
                <option key={state.id} value={state.id}>
                  {state.name}
                </option>
              ))}
            </TextField>
          ) : null}

          <Divider sx={{ my: 2 }} />

          {sourceType === 'gcp' && (
            <>
              <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                Configuration GCP Cloud Storage
              </Typography>
              <TextField
                fullWidth
                label="Bucket GCP"
                value={gcpConfig.bucket}
                onChange={(e) => setGcpConfig({ ...gcpConfig, bucket: e.target.value })}
                margin="normal"
                required
                placeholder="mon-bucket-gcp"
              />
              <TextField
                fullWidth
                label="Chemin de l'objet (ex: path/to/default.tfstate)"
                value={gcpConfig.object_name}
                onChange={(e) => setGcpConfig({ ...gcpConfig, object_name: e.target.value })}
                margin="normal"
                required
                placeholder="terraform/states/default.tfstate"
              />
              <TextField
                fullWidth
                label={editingSource ? 'Credentials JSON (laisser vide pour conserver)' : 'Credentials JSON (Service Account) *'}
                value={editingSource && !gcpConfig.credentials_json ? '••••••••••••••••••••••••••••••••' : gcpConfig.credentials_json}
                onChange={(e) => {
                  // Si on est en mode édition et qu'on commence à taper, remplacer le masquage
                  const newValue = e.target.value === '••••••••••••••••••••••••••••••••' ? '' : e.target.value
                  setGcpConfig({ ...gcpConfig, credentials_json: newValue })
                }}
                onFocus={(e) => {
                  // Si on focus et que c'est le masquage, vider le champ
                  const target = e.target as HTMLInputElement | HTMLTextAreaElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    setGcpConfig({ ...gcpConfig, credentials_json: '' })
                  }
                }}
                onCopy={(e) => {
                  // Empêcher la copie si c'est le masquage
                  const target = e.target as HTMLInputElement | HTMLTextAreaElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    e.preventDefault()
                  }
                }}
                margin="normal"
                required={!editingSource}
                multiline
                rows={4}
                placeholder={editingSource ? 'Credentials existants (masqués)' : '{"type": "service_account", "project_id": "...", ...}'}
                helperText={editingSource ? 'Credentials existants masqués. Laisser vide pour conserver, ou saisir de nouveaux credentials pour les remplacer.' : 'JSON complet de la clé de service GCP'}
                InputProps={{
                  sx: editingSource && !gcpConfig.credentials_json ? {
                    '& input': {
                      color: 'text.secondary',
                      fontFamily: 'monospace',
                      letterSpacing: '0.1em',
                    }
                  } : {}
                }}
              />
              <TextField
                fullWidth
                label="Intervalle de synchronisation"
                value={gcpConfig.sync_interval}
                onChange={(e) => setGcpConfig({ ...gcpConfig, sync_interval: e.target.value })}
                margin="normal"
                placeholder="15m"
                helperText="Format: 15m, 1h, 30s, etc."
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={gcpConfig.auto_sync}
                    onChange={(e) => setGcpConfig({ ...gcpConfig, auto_sync: e.target.checked })}
                  />
                }
                label="Synchronisation automatique"
                sx={{ mt: 1 }}
              />
            </>
          )}

          {sourceType === 's3' && (
            <>
              <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                Configuration AWS S3
              </Typography>
              <TextField
                fullWidth
                label="Bucket S3"
                value={s3Config.bucket}
                onChange={(e) => setS3Config({ ...s3Config, bucket: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label="Clé S3 (chemin)"
                value={s3Config.key}
                onChange={(e) => setS3Config({ ...s3Config, key: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label="Région"
                value={s3Config.region}
                onChange={(e) => setS3Config({ ...s3Config, region: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label="Endpoint (optionnel, pour S3-compatible)"
                value={s3Config.endpoint}
                onChange={(e) => setS3Config({ ...s3Config, endpoint: e.target.value })}
                margin="normal"
              />
              <TextField
                fullWidth
                label="AWS Access Key ID"
                value={editingSource && !s3Config.aws_access_key_id ? '••••••••••••••••' : s3Config.aws_access_key_id}
                onChange={(e) => {
                  const newValue = e.target.value === '••••••••••••••••' ? '' : e.target.value
                  setS3Config({ ...s3Config, aws_access_key_id: newValue })
                }}
                onFocus={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••') {
                    setS3Config({ ...s3Config, aws_access_key_id: '' })
                  }
                }}
                onCopy={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••') {
                    e.preventDefault()
                  }
                }}
                margin="normal"
                type="password"
                placeholder={editingSource ? 'Access Key existant (masqué)' : ''}
                helperText={editingSource ? 'Laisser vide pour conserver, ou saisir une nouvelle clé pour la remplacer' : ''}
              />
              <TextField
                fullWidth
                label="AWS Secret Access Key"
                value={editingSource && !s3Config.aws_secret_access_key ? '••••••••••••••••••••••••••••••••' : s3Config.aws_secret_access_key}
                onChange={(e) => {
                  const newValue = e.target.value === '••••••••••••••••••••••••••••••••' ? '' : e.target.value
                  setS3Config({ ...s3Config, aws_secret_access_key: newValue })
                }}
                onFocus={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    setS3Config({ ...s3Config, aws_secret_access_key: '' })
                  }
                }}
                onCopy={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    e.preventDefault()
                  }
                }}
                margin="normal"
                type="password"
                placeholder={editingSource ? 'Secret Key existant (masqué)' : ''}
                helperText={editingSource ? 'Laisser vide pour conserver, ou saisir une nouvelle clé pour la remplacer' : ''}
              />
              <TextField
                fullWidth
                label="Intervalle de synchronisation"
                value={s3Config.sync_interval}
                onChange={(e) => setS3Config({ ...s3Config, sync_interval: e.target.value })}
                margin="normal"
                placeholder="15m"
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={s3Config.auto_sync}
                    onChange={(e) => setS3Config({ ...s3Config, auto_sync: e.target.checked })}
                  />
                }
                label="Synchronisation automatique"
                sx={{ mt: 1 }}
              />
            </>
          )}

          {sourceType === 'azure' && (
            <>
              <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                Configuration Azure Blob Storage
              </Typography>
              <TextField
                fullWidth
                label="Nom du compte de stockage"
                value={azureConfig.account_name}
                onChange={(e) => setAzureConfig({ ...azureConfig, account_name: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label="Container"
                value={azureConfig.container}
                onChange={(e) => setAzureConfig({ ...azureConfig, container: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label="Nom du blob"
                value={azureConfig.blob_name}
                onChange={(e) => setAzureConfig({ ...azureConfig, blob_name: e.target.value })}
                margin="normal"
                required
              />
              <TextField
                fullWidth
                label={`Account Key ${!editingSource ? '*' : ''}`}
                value={editingSource && !azureConfig.account_key ? '••••••••••••••••••••••••••••••••' : azureConfig.account_key}
                onChange={(e) => {
                  const newValue = e.target.value === '••••••••••••••••••••••••••••••••' ? '' : e.target.value
                  setAzureConfig({ ...azureConfig, account_key: newValue })
                }}
                onFocus={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    setAzureConfig({ ...azureConfig, account_key: '' })
                  }
                }}
                onCopy={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    e.preventDefault()
                  }
                }}
                margin="normal"
                type="password"
                required={!editingSource}
                placeholder={editingSource ? 'Account Key existant (masqué)' : ''}
                helperText={editingSource ? 'Account Key existant masqué. Laisser vide pour conserver, ou saisir une nouvelle clé pour la remplacer.' : 'OBLIGATOIRE : Clé d\'accès du compte de stockage Azure (trouvable dans Azure Portal / Storage Account / Access Keys)'}
              />
              <TextField
                fullWidth
                label="Connection String (optionnel, prioritaire sur Account Key)"
                value={editingSource && !azureConfig.connection_string ? '••••••••••••••••••••••••••••••••' : azureConfig.connection_string}
                onChange={(e) => {
                  const newValue = e.target.value === '••••••••••••••••••••••••••••••••' ? '' : e.target.value
                  setAzureConfig({ ...azureConfig, connection_string: newValue })
                }}
                onFocus={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    setAzureConfig({ ...azureConfig, connection_string: '' })
                  }
                }}
                onCopy={(e) => {
                  const target = e.target as HTMLInputElement
                  if (target.value === '••••••••••••••••••••••••••••••••') {
                    e.preventDefault()
                  }
                }}
                margin="normal"
                type="password"
                placeholder={editingSource ? 'Connection String existant (masqué)' : ''}
                helperText={editingSource ? 'Connection String existant masqué. Laisser vide pour conserver, ou saisir une nouvelle chaîne pour la remplacer.' : 'Alternative : Connection String complète (remplace Account Key si fournie)'}
              />
              <TextField
                fullWidth
                label="Intervalle de synchronisation"
                value={azureConfig.sync_interval}
                onChange={(e) => setAzureConfig({ ...azureConfig, sync_interval: e.target.value })}
                margin="normal"
                placeholder="15m"
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={azureConfig.auto_sync}
                    onChange={(e) => setAzureConfig({ ...azureConfig, auto_sync: e.target.checked })}
                  />
                }
                label="Synchronisation automatique"
                sx={{ mt: 1 }}
              />
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Annuler</Button>
          <Button
            onClick={() => addSourceMutation.mutate()}
            variant="contained"
            disabled={addSourceMutation.isPending}
          >
            {editingSource ? 'Modifier' : createStateFromSource ? 'Créer et lier' : 'Lier'}
          </Button>
        </DialogActions>
      </Dialog>

      {activeTab === 1 && (
        <>
          {sourcesLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress />
            </Box>
          ) : !sources || sources.items.length === 0 ? (
            <Card sx={{ animation: 'jellyfishFloat 12s ease-in-out infinite' }}>
              <CardContent sx={{ textAlign: 'center', py: 6 }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  Aucune source de synchronisation
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                  Liez une source cloud pour synchroniser automatiquement vos états Terraform.
                </Typography>
                <Button variant="contained" startIcon={<CloudQueueIcon />} onClick={() => setSourceDialogOpen(true)}>
                  Lier une source cloud
                </Button>
              </CardContent>
            </Card>
          ) : (
            <Grid container spacing={3}>
              {sources.items.map((source, idx) => (
                <Grid item xs={12} md={6} lg={4} key={source.id}>
                  <Card 
                    sx={{ 
                      height: '100%', 
                      display: 'flex', 
                      flexDirection: 'column',
                      '--card-delay': `${idx * 0.3}s`,
                    }}
                  >
                    <CardContent sx={{ flexGrow: 1 }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', mb: 2.5 }}>
                        <Typography 
                          variant="h6" 
                          component="div" 
                          sx={{ 
                            fontWeight: 500, 
                            color: 'rgba(255, 255, 255, 0.95)', 
                            fontSize: '1rem',
                            fontFamily: '"JetBrains Mono", monospace',
                          }}
                        >
                          {source.type === 'gcp' && `${source.config.gcp_bucket} /${source.config.gcp_object_name}`}
                          {source.type === 's3' && `${source.config.s3_bucket}/${source.config.s3_key}`}
                          {source.type === 'azure' && `${source.config.azure_account_name}/${source.config.azure_container}/${source.config.azure_blob_name}`}
                        </Typography>
                        <Chip
                          label={source.enabled ? 'Actif' : 'Inactif'}
                          size="small"
                          color={source.enabled ? 'success' : 'default'}
                          sx={{ fontSize: '0.75rem', height: 24 }}
                        />
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                          Type: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{source.type.toUpperCase()}</span>
                        </Typography>
                        {source.last_sync && (
                          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                            Dernière sync: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{new Date(source.last_sync).toLocaleString('fr-FR')}</span>
                          </Typography>
                        )}
                        {source.next_sync && (
                          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                            Prochaine sync: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{new Date(source.next_sync).toLocaleString('fr-FR')}</span>
                          </Typography>
                        )}
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.875rem' }}>
                          Intervalle: <span style={{ color: 'rgba(255, 255, 255, 0.85)' }}>{source.config.sync_interval || '15m'}</span>
                        </Typography>
                      </Box>
                      <Box sx={{ mt: 'auto', display: 'flex', gap: 1.5, pt: 2.5 }}>
                        <Button
                          size="small"
                          variant="outlined"
                          startIcon={<SyncIcon />}
                          onClick={() => syncSourceMutation.mutate(source.id)}
                          disabled={syncSourceMutation.isPending}
                        >
                          Synchroniser
                        </Button>
                        <Button
                          size="small"
                          variant="outlined"
                          startIcon={<EditIcon />}
                          onClick={() => handleEditSource(source)}
                          sx={{ mr: 1 }}
                        >
                          Modifier
                        </Button>
                        <Button
                          size="small"
                          variant="outlined"
                          color="error"
                          startIcon={<DeleteIcon />}
                          onClick={() => {
                            if (window.confirm('Êtes-vous sûr de vouloir supprimer cette source de synchronisation ?')) {
                              deleteSourceMutation.mutate(source.id)
                            }
                          }}
                          disabled={deleteSourceMutation.isPending}
                        >
                          Supprimer
                        </Button>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>
          )}
        </>
      )}

      {/* Snackbar pour les notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  )
}
