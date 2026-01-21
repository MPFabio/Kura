import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  CircularProgress,
  Alert,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Tooltip,
  Snackbar,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Switch,
  FormControlLabel,
} from '@mui/material'
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  CheckCircle as CheckCircleIcon,
  CloudUpload as CloudUploadIcon,
} from '@mui/icons-material'
import { clusterService, KubernetesCluster } from '../services/clusterService'

export default function ClustersPage() {
  const [clusterDialogOpen, setClusterDialogOpen] = useState(false)
  const [editingCluster, setEditingCluster] = useState<KubernetesCluster | null>(null)
  const [clusterForm, setClusterForm] = useState({
    name: '',
    description: '',
    endpoint: '',
    kubeconfig: '',
    is_active: false,
  })
  const [kubeconfigFile, setKubeconfigFile] = useState<File | null>(null)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })
  const queryClient = useQueryClient()

  const { data: clusters, isLoading, error } = useQuery({
    queryKey: ['clusters'],
    queryFn: () => clusterService.getClusters(),
  })

  // activeCluster non utilisé pour l'instant
  // const { data: activeCluster } = useQuery({
  //   queryKey: ['active-cluster'],
  //   queryFn: () => clusterService.getActiveCluster(),
  // })

  const createMutation = useMutation({
    mutationFn: (cluster: any) => clusterService.createCluster(cluster),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      setClusterDialogOpen(false)
      resetForm()
      setSnackbar({ open: true, message: 'Cluster créé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || 'Erreur lors de la création du cluster',
        severity: 'error',
      })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, cluster }: { id: string; cluster: any }) => clusterService.updateCluster(id, cluster),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      setClusterDialogOpen(false)
      resetForm()
      setSnackbar({ open: true, message: 'Cluster mis à jour avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || 'Erreur lors de la mise à jour du cluster',
        severity: 'error',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => clusterService.deleteCluster(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      setSnackbar({ open: true, message: 'Cluster supprimé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || 'Erreur lors de la suppression du cluster',
        severity: 'error',
      })
    },
  })

  const activateMutation = useMutation({
    mutationFn: (id: string) => clusterService.setActiveCluster(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      setSnackbar({ open: true, message: 'Cluster activé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      setSnackbar({
        open: true,
        message: error.response?.data?.error || 'Erreur lors de l\'activation du cluster',
        severity: 'error',
      })
    },
  })

  const resetForm = () => {
    setClusterForm({
      name: '',
      description: '',
      endpoint: '',
      kubeconfig: '',
      is_active: false,
    })
    setEditingCluster(null)
    setKubeconfigFile(null)
  }

  const handleOpenDialog = (cluster?: KubernetesCluster) => {
    if (cluster) {
      setEditingCluster(cluster)
      setClusterForm({
        name: cluster.name,
        description: cluster.description || '',
        endpoint: cluster.endpoint || '',
        kubeconfig: '', // Ne pas pré-remplir le kubeconfig pour la sécurité
        is_active: cluster.is_active,
      })
    } else {
      resetForm()
    }
    setClusterDialogOpen(true)
  }

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const file = event.target.files[0]
      setKubeconfigFile(file)
      const reader = new FileReader()
      reader.onload = (e) => {
        const content = e.target?.result as string
        setClusterForm({ ...clusterForm, kubeconfig: content })
      }
      reader.readAsText(file)
    }
  }

  const handleSubmit = () => {
    if (!clusterForm.name || !clusterForm.kubeconfig) {
      setSnackbar({ open: true, message: 'Le nom et le kubeconfig sont requis', severity: 'error' })
      return
    }

    if (editingCluster) {
      updateMutation.mutate({ id: editingCluster.id, cluster: clusterForm })
    } else {
      createMutation.mutate(clusterForm)
    }
  }

  const handleDelete = (id: string) => {
    if (window.confirm('Êtes-vous sûr de vouloir supprimer ce cluster ?')) {
      deleteMutation.mutate(id)
    }
  }

  const handleActivate = (id: string) => {
    activateMutation.mutate(id)
  }

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          Erreur lors du chargement des clusters : {error instanceof Error ? error.message : 'Erreur inconnue'}
        </Alert>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Clusters Kubernetes
        </Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenDialog()}>
          Ajouter un cluster
        </Button>
      </Box>

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : !clusters || clusters.items.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              Aucun cluster Kubernetes configuré
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Ajoutez un cluster Kubernetes pour commencer à gérer vos ressources.
            </Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenDialog()}>
              Ajouter un cluster
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Nom</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Endpoint</TableCell>
                <TableCell>Statut</TableCell>
                <TableCell>Créé le</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {clusters.items.map((cluster) => (
                <TableRow key={cluster.id}>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Typography variant="body1" fontWeight="medium">
                        {cluster.name}
                      </Typography>
                      {cluster.is_active && (
                        <Chip
                          label="Actif"
                          color="success"
                          size="small"
                          icon={<CheckCircleIcon />}
                        />
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>{cluster.description || '-'}</TableCell>
                  <TableCell>{cluster.endpoint || '-'}</TableCell>
                  <TableCell>
                    {cluster.is_active ? (
                      <Chip label="Actif" color="success" size="small" />
                    ) : (
                      <Chip label="Inactif" color="default" size="small" />
                    )}
                  </TableCell>
                  <TableCell>
                    {new Date(cluster.created_at).toLocaleDateString('fr-FR')}
                  </TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: 'flex', gap: 1, justifyContent: 'flex-end' }}>
                      {!cluster.is_active && (
                        <Tooltip title="Activer ce cluster">
                          <IconButton
                            size="small"
                            color="primary"
                            onClick={() => handleActivate(cluster.id)}
                            disabled={activateMutation.isPending}
                          >
                            <CheckCircleIcon />
                          </IconButton>
                        </Tooltip>
                      )}
                      <Tooltip title="Modifier">
                        <IconButton size="small" onClick={() => handleOpenDialog(cluster)}>
                          <EditIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Supprimer">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDelete(cluster.id)}
                          disabled={cluster.is_active}
                        >
                          <DeleteIcon />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Dialog d'ajout/modification de cluster */}
      <Dialog open={clusterDialogOpen} onClose={() => setClusterDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>
          {editingCluster ? 'Modifier le cluster' : 'Ajouter un cluster Kubernetes'}
        </DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Nom du cluster"
            value={clusterForm.name}
            onChange={(e) => setClusterForm({ ...clusterForm, name: e.target.value })}
            margin="normal"
            required
            placeholder="Ex: Production, Staging, Minikube"
          />
          <TextField
            fullWidth
            label="Description (optionnel)"
            value={clusterForm.description}
            onChange={(e) => setClusterForm({ ...clusterForm, description: e.target.value })}
            margin="normal"
            multiline
            rows={2}
          />
          <TextField
            fullWidth
            label="Endpoint API (optionnel)"
            value={clusterForm.endpoint}
            onChange={(e) => setClusterForm({ ...clusterForm, endpoint: e.target.value })}
            margin="normal"
            placeholder="https://kubernetes.example.com:6443"
          />
          <Box sx={{ mt: 2 }}>
            <Button
              variant="outlined"
              component="label"
              fullWidth
              startIcon={<CloudUploadIcon />}
            >
              {kubeconfigFile ? kubeconfigFile.name : 'Sélectionner un fichier kubeconfig'}
              <input type="file" hidden accept=".yaml,.yml" onChange={handleFileChange} />
            </Button>
            {kubeconfigFile && (
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                Fichier sélectionné : {kubeconfigFile.name} ({(kubeconfigFile.size / 1024).toFixed(2)} KB)
              </Typography>
            )}
          </Box>
          <TextField
            fullWidth
            label="Ou coller le contenu du kubeconfig"
            value={clusterForm.kubeconfig}
            onChange={(e) => setClusterForm({ ...clusterForm, kubeconfig: e.target.value })}
            margin="normal"
            multiline
            rows={8}
            placeholder="apiVersion: v1&#10;clusters:&#10;..."
            sx={{ mt: 2 }}
          />
          <FormControlLabel
            control={
              <Switch
                checked={clusterForm.is_active}
                onChange={(e) => setClusterForm({ ...clusterForm, is_active: e.target.checked })}
              />
            }
            label="Activer ce cluster immédiatement"
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setClusterDialogOpen(false)}>Annuler</Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={!clusterForm.name || !clusterForm.kubeconfig || createMutation.isPending || updateMutation.isPending}
          >
            {editingCluster ? 'Modifier' : 'Créer'}
          </Button>
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
