import { useState } from 'react'
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
} from '@mui/material'
import {
  CloudUpload as CloudUploadIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Warning as WarningIcon,
} from '@mui/icons-material'
import { terraformService, TerraformState } from '../services/terraformService'

export default function TerraformPage() {
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [stateName, setStateName] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [selectedState, setSelectedState] = useState<TerraformState | null>(null)
  const [detailDialogOpen, setDetailDialogOpen] = useState(false)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })
  const queryClient = useQueryClient()

  const { data: states, isLoading, error, refetch } = useQuery({
    queryKey: ['terraform-states'],
    queryFn: () => terraformService.getStates(),
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
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          États Terraform
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
          <Tooltip title="Actualiser">
            <IconButton onClick={() => refetch()} disabled={isLoading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : !states || states.items.length === 0 ? (
        <Card>
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
          {states.items.map((state) => (
            <Grid item xs={12} md={6} lg={4} key={state.id}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', mb: 2 }}>
                    <Typography variant="h6" component="div">
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
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Version: {state.state?.version || 'N/A'}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Ressources: {state.state?.resources?.length || 0}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Sorties: {Object.keys(state.state?.outputs || {}).length}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 1 }}>
                    Uploadé le : {new Date(state.uploaded_at).toLocaleString('fr-FR')}
                  </Typography>
                  <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
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
                        try {
                          await terraformService.detectDrift(state.id)
                          setSnackbar({ open: true, message: 'Détection de drift lancée', severity: 'success' })
                        } catch (error) {
                          setSnackbar({ open: true, message: 'Erreur lors de la détection de drift', severity: 'error' })
                        }
                      }}
                    >
                      Drift
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
                        <TableCell>Type</TableCell>
                        <TableCell>Nom</TableCell>
                        <TableCell>Provider</TableCell>
                        <TableCell>Mode</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {selectedState.state.resources.map((resource, idx) => (
                        <TableRow key={idx}>
                          <TableCell>{resource.type}</TableCell>
                          <TableCell>{resource.name}</TableCell>
                          <TableCell>{resource.provider}</TableCell>
                          <TableCell>
                            <Chip label={resource.mode} size="small" color={resource.mode === 'managed' ? 'primary' : 'secondary'} />
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
