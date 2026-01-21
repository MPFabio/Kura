import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
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
  Tabs,
  Tab,
  TextField,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Tooltip,
  Snackbar,
  Checkbox,
  AppBar,
  FormControlLabel,
  Switch,
} from '@mui/material'
import {
  Delete as DeleteIcon,
  Edit as EditIcon,
  Search as SearchIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { k8sService } from '../services/k8sService'
import { clusterService, KubernetesCluster } from '../services/clusterService'
import { Deployment } from '../services/api'
import ResourceDetailDialog from '../components/ResourceDetailDialog'
import {
  Add as AddIcon,
  CheckCircle as CheckCircleIcon,
  CloudUpload as CloudUploadIcon,
} from '@mui/icons-material'

type ResourceTab = 'clusters' | 'pods' | 'deployments' | 'services' | 'configmaps' | 'secrets' | 'nodes'

export default function K8sPage() {
  const [selectedNamespace, setSelectedNamespace] = useState<string>('')
  const [activeTab, setActiveTab] = useState<ResourceTab>('clusters')
  const [searchTerm, setSearchTerm] = useState<string>('')
  const [scaleDialog, setScaleDialog] = useState<{
    open: boolean
    namespace: string
    name: string
    currentReplicas: number
  }>({
    open: false,
    namespace: '',
    name: '',
    currentReplicas: 0,
  })
  const [scaleReplicas, setScaleReplicas] = useState<number>(1)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'warning' }>({
    open: false,
    message: '',
    severity: 'success',
  })
  const queryClient = useQueryClient()
  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const [detailDialog, setDetailDialog] = useState<{
    open: boolean
    type: 'pod' | 'deployment' | 'service'
    namespace: string
    name: string
  }>({
    open: false,
    type: 'pod',
    namespace: '',
    name: '',
  })
  const [selectedResources, setSelectedResources] = useState<Set<string>>(new Set())
  const [bulkActionDialog, setBulkActionDialog] = useState<{
    open: boolean
    action: 'delete' | 'restart' | 'scale' | null
    count: number
  }>({
    open: false,
    action: null,
    count: 0,
  })

  // États pour la gestion des clusters
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

  // Queries pour les clusters
  const { data: clusters, isLoading: clustersLoading, error: clustersError } = useQuery({
    queryKey: ['clusters'],
    queryFn: () => clusterService.getClusters(),
  })

  // activeCluster non utilisé pour l'instant mais peut être utile plus tard
  // const { data: activeCluster } = useQuery({
  //   queryKey: ['active-cluster'],
  //   queryFn: () => clusterService.getActiveCluster(),
  // })

  const { data: namespaces, isLoading: namespacesLoading, error: namespacesError } = useQuery({
    queryKey: ['namespaces'],
    queryFn: () => k8sService.getNamespaces(),
    retry: false,
    enabled: activeTab !== 'clusters', // Ne charger que si on n'est pas sur l'onglet clusters
  })

  const { data: pods, isLoading: podsLoading, error: podsError } = useQuery({
    queryKey: ['pods', selectedNamespace],
    queryFn: () => k8sService.getPods(selectedNamespace),
    enabled: !!selectedNamespace && activeTab === 'pods',
    retry: 1,
  })

  const { data: deployments, isLoading: deploymentsLoading, error: deploymentsError } = useQuery({
    queryKey: ['deployments', selectedNamespace],
    queryFn: () => k8sService.getDeployments(selectedNamespace),
    enabled: !!selectedNamespace && activeTab === 'deployments',
    retry: 1,
  })

  const { data: services, isLoading: servicesLoading, error: servicesError } = useQuery({
    queryKey: ['services', selectedNamespace],
    queryFn: () => k8sService.getServices(selectedNamespace),
    enabled: !!selectedNamespace && activeTab === 'services',
    retry: 1,
  })

  const { data: configMaps, isLoading: configMapsLoading, error: configMapsError } = useQuery({
    queryKey: ['configmaps', selectedNamespace],
    queryFn: () => k8sService.getConfigMaps(selectedNamespace),
    enabled: !!selectedNamespace && activeTab === 'configmaps',
    retry: 1,
  })

  const { data: secrets, isLoading: secretsLoading, error: secretsError } = useQuery({
    queryKey: ['secrets', selectedNamespace],
    queryFn: () => k8sService.getSecrets(selectedNamespace),
    enabled: !!selectedNamespace && activeTab === 'secrets',
    retry: 1,
  })

  const { data: nodes, isLoading: nodesLoading, error: nodesError } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => k8sService.getNodes(),
    enabled: activeTab === 'nodes',
    retry: 1,
  })

  const deletePodMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      k8sService.deletePod(namespace, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['pods', selectedNamespace] })
      await queryClient.refetchQueries({ queryKey: ['pods', selectedNamespace] })
      setSnackbar({ open: true, message: 'Pod supprimé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur: ${errorMessage}`, severity: 'error' })
    },
  })

  const deleteDeploymentMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      k8sService.deleteDeployment(namespace, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['deployments', selectedNamespace] })
      await queryClient.refetchQueries({ queryKey: ['deployments', selectedNamespace] })
      setSnackbar({ open: true, message: 'Deployment supprimé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur: ${errorMessage}`, severity: 'error' })
    },
  })

  const deleteServiceMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      k8sService.deleteService(namespace, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['services', selectedNamespace] })
      await queryClient.refetchQueries({ queryKey: ['services', selectedNamespace] })
      setSnackbar({ open: true, message: 'Service supprimé avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur: ${errorMessage}`, severity: 'error' })
    },
  })

  // Nettoyer le polling quand le composant se démonte ou change de namespace
  useEffect(() => {
    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
        pollingIntervalRef.current = null
      }
    }
  }, [selectedNamespace])

  const scaleDeploymentMutation = useMutation({
    mutationFn: ({ namespace, name, replicas }: { namespace: string; name: string; replicas: number }) =>
      k8sService.scaleDeployment(namespace, name, replicas),
    onSuccess: async () => {
      setScaleDialog({ ...scaleDialog, open: false })
      setSnackbar({
        open: true,
        message: `Deployment mis à jour: ${scaleReplicas} replicas`,
        severity: 'success',
      })
      
      // Invalider et forcer un refetch immédiat
      await queryClient.invalidateQueries({ queryKey: ['deployments', selectedNamespace] })
      await queryClient.refetchQueries({ queryKey: ['deployments', selectedNamespace] })
      
      // Nettoyer l'ancien polling s'il existe
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
      }
      
      // Démarrer un polling temporaire pour suivre l'avancement
      pollingIntervalRef.current = setInterval(async () => {
        if (!selectedNamespace) {
          if (pollingIntervalRef.current) {
            clearInterval(pollingIntervalRef.current)
            pollingIntervalRef.current = null
          }
          return
        }
        
        await queryClient.refetchQueries({ queryKey: ['deployments', selectedNamespace] })
        // Récupérer les données depuis le cache
        const queryData = queryClient.getQueryData<{ items: Deployment[] }>(['deployments', selectedNamespace])
        const data = queryData
        
        // Arrêter le polling si tous les deployments sont prêts
        if (data?.items) {
          const isScaling = data.items.some(
            (dep) => dep.readyReplicas < dep.replicas && dep.replicas > 0
          )
          if (!isScaling) {
            if (pollingIntervalRef.current) {
              clearInterval(pollingIntervalRef.current)
              pollingIntervalRef.current = null
            }
          }
        }
      }, 2000)
      
      // Arrêter le polling après 2 minutes maximum
      setTimeout(() => {
        if (pollingIntervalRef.current) {
          clearInterval(pollingIntervalRef.current)
          pollingIntervalRef.current = null
        }
      }, 120000)
    },
    onError: (error: any) => {
      console.error('Erreur lors du scale:', error)
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur: ${errorMessage}`, severity: 'error' })
    },
  })

  const handleDelete = (type: 'pod' | 'deployment' | 'service', namespace: string, name: string) => {
    if (window.confirm(`Êtes-vous sûr de vouloir supprimer ${type} ${name} ?`)) {
      if (type === 'pod') {
        deletePodMutation.mutate({ namespace, name })
      } else if (type === 'deployment') {
        deleteDeploymentMutation.mutate({ namespace, name })
      } else if (type === 'service') {
        deleteServiceMutation.mutate({ namespace, name })
      }
    }
  }

  const handleScaleClick = (namespace: string, name: string, currentReplicas: number) => {
    setScaleReplicas(currentReplicas)
    setScaleDialog({ open: true, namespace, name, currentReplicas })
  }

  const handleScale = () => {
    scaleDeploymentMutation.mutate({
      namespace: scaleDialog.namespace,
      name: scaleDialog.name,
      replicas: scaleReplicas,
    })
  }

  // Mutations pour les clusters
  const createClusterMutation = useMutation({
    mutationFn: (cluster: any) => clusterService.createCluster(cluster),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      queryClient.invalidateQueries({ queryKey: ['namespaces'] })
      setClusterDialogOpen(false)
      resetClusterForm()
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

  const updateClusterMutation = useMutation({
    mutationFn: ({ id, cluster }: { id: string; cluster: any }) => clusterService.updateCluster(id, cluster),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      queryClient.invalidateQueries({ queryKey: ['namespaces'] })
      setClusterDialogOpen(false)
      resetClusterForm()
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

  const deleteClusterMutation = useMutation({
    mutationFn: (id: string) => clusterService.deleteCluster(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      queryClient.invalidateQueries({ queryKey: ['namespaces'] })
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

  const activateClusterMutation = useMutation({
    mutationFn: (id: string) => clusterService.setActiveCluster(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clusters'] })
      queryClient.invalidateQueries({ queryKey: ['active-cluster'] })
      queryClient.invalidateQueries({ queryKey: ['namespaces'] })
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

  const resetClusterForm = () => {
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

  const handleOpenClusterDialog = (cluster?: KubernetesCluster) => {
    if (cluster) {
      setEditingCluster(cluster)
      setClusterForm({
        name: cluster.name,
        description: cluster.description || '',
        endpoint: cluster.endpoint || '',
        kubeconfig: '', // Ne pas pré-remplir pour la sécurité
        is_active: cluster.is_active,
      })
    } else {
      resetClusterForm()
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

  const handleClusterSubmit = () => {
    if (!clusterForm.name || !clusterForm.kubeconfig) {
      setSnackbar({ open: true, message: 'Le nom et le kubeconfig sont requis', severity: 'error' })
      return
    }

    if (editingCluster) {
      updateClusterMutation.mutate({ id: editingCluster.id, cluster: clusterForm })
    } else {
      createClusterMutation.mutate(clusterForm)
    }
  }

  const handleDeleteCluster = (id: string) => {
    if (window.confirm('Êtes-vous sûr de vouloir supprimer ce cluster ?')) {
      deleteClusterMutation.mutate(id)
    }
  }

  const handleActivateCluster = (id: string) => {
    activateClusterMutation.mutate(id)
  }

  const filterResources = <T extends { name: string }>(items: T[] | undefined): T[] => {
    if (!items) return []
    if (!searchTerm) return items
    return items.filter((item) => item.name.toLowerCase().includes(searchTerm.toLowerCase()))
  }

  // Gestion de la sélection multiple
  const handleSelectAll = (resourceNames: string[]) => {
    const allSelected = resourceNames.every((name) => selectedResources.has(name))
    if (allSelected) {
      setSelectedResources(new Set())
    } else {
      setSelectedResources(new Set(resourceNames))
    }
  }

  const handleSelectOne = (resourceName: string) => {
    const newSelected = new Set(selectedResources)
    if (newSelected.has(resourceName)) {
      newSelected.delete(resourceName)
    } else {
      newSelected.add(resourceName)
    }
    setSelectedResources(newSelected)
  }

  const isSelected = (resourceName: string) => selectedResources.has(resourceName)

  // Réinitialiser la sélection quand on change d'onglet ou de namespace
  useEffect(() => {
    setSelectedResources(new Set())
  }, [activeTab, selectedNamespace])

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
      case 'ready':
      case 'active':
        return 'success'
      case 'pending':
        return 'warning'
      case 'failed':
      case 'error':
      case 'notready':
        return 'error'
      default:
        return 'default'
    }
  }

  const renderPodsTable = () => {
    if (!selectedNamespace) {
      return (
        <Alert severity="info" sx={{ mt: 2 }}>
          Veuillez sélectionner un namespace pour afficher les pods.
        </Alert>
      )
    }
    if (podsLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (podsError) {
      return (
        <Alert severity="error">
          Erreur: {(podsError as any)?.response?.data?.error || (podsError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredPods = filterResources(pods?.items)
    if (!pods?.items || pods.items.length === 0) {
      return <Alert severity="info">Aucun pod trouvé</Alert>
    }
    if (filteredPods.length === 0) {
      return <Alert severity="info">Aucun pod ne correspond à la recherche</Alert>
    }
    const podNames = filteredPods.map((p) => p.name)
    const allSelected = podNames.length > 0 && podNames.every((name) => selectedResources.has(name))
    const someSelected = podNames.some((name) => selectedResources.has(name))

    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  indeterminate={someSelected && !allSelected}
                  checked={allSelected}
                  onChange={() => handleSelectAll(podNames)}
                />
              </TableCell>
              <TableCell>Nom</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Statut</TableCell>
              <TableCell>Node</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredPods.map((pod) => (
              <TableRow
                key={pod.name}
                hover
                selected={isSelected(pod.name)}
                sx={{ cursor: 'pointer' }}
                onClick={() =>
                  setDetailDialog({
                    open: true,
                    type: 'pod',
                    namespace: pod.namespace,
                    name: pod.name,
                  })
                }
              >
                <TableCell padding="checkbox" onClick={(e) => e.stopPropagation()}>
                  <Checkbox
                    checked={isSelected(pod.name)}
                    onChange={() => handleSelectOne(pod.name)}
                  />
                </TableCell>
                <TableCell>{pod.name}</TableCell>
                <TableCell>{pod.namespace}</TableCell>
                <TableCell>
                  <Chip label={pod.status} color={getStatusColor(pod.status) as any} size="small" />
                </TableCell>
                <TableCell>{pod.node || '-'}</TableCell>
                <TableCell align="right" onClick={(e) => e.stopPropagation()}>
                  <Tooltip title="Supprimer">
                    <IconButton
                      size="small"
                      color="error"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDelete('pod', pod.namespace, pod.name)
                      }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  const renderDeploymentsTable = () => {
    if (!selectedNamespace) {
      return (
        <Alert severity="info" sx={{ mt: 2 }}>
          Veuillez sélectionner un namespace pour afficher les deployments.
        </Alert>
      )
    }
    if (deploymentsLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (deploymentsError) {
      return (
        <Alert severity="error">
          Erreur: {(deploymentsError as any)?.response?.data?.error || (deploymentsError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredDeployments = filterResources(deployments?.items)
    if (!deployments?.items || deployments.items.length === 0) {
      return <Alert severity="info">Aucun deployment trouvé</Alert>
    }
    if (filteredDeployments.length === 0) {
      return <Alert severity="info">Aucun deployment ne correspond à la recherche</Alert>
    }
    const deploymentNames = filteredDeployments.map((d) => d.name)
    const allSelected = deploymentNames.length > 0 && deploymentNames.every((name) => selectedResources.has(name))
    const someSelected = deploymentNames.some((name) => selectedResources.has(name))

    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  indeterminate={someSelected && !allSelected}
                  checked={allSelected}
                  onChange={() => handleSelectAll(deploymentNames)}
                />
              </TableCell>
              <TableCell>Nom</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Replicas</TableCell>
              <TableCell>Ready</TableCell>
              <TableCell>Available</TableCell>
              <TableCell>Statut</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredDeployments.map((dep) => (
              <TableRow
                key={dep.name}
                hover
                selected={isSelected(dep.name)}
                sx={{ cursor: 'pointer' }}
                onClick={() =>
                  setDetailDialog({
                    open: true,
                    type: 'deployment',
                    namespace: dep.namespace,
                    name: dep.name,
                  })
                }
              >
                <TableCell padding="checkbox" onClick={(e) => e.stopPropagation()}>
                  <Checkbox
                    checked={isSelected(dep.name)}
                    onChange={() => handleSelectOne(dep.name)}
                  />
                </TableCell>
                <TableCell>{dep.name}</TableCell>
                <TableCell>{dep.namespace}</TableCell>
                <TableCell>{dep.replicas}</TableCell>
                <TableCell>
                  <Chip
                    label={`${dep.readyReplicas}/${dep.replicas}`}
                    color={dep.readyReplicas === dep.replicas ? 'success' : 'warning'}
                    size="small"
                  />
                </TableCell>
                <TableCell>{dep.availableReplicas}</TableCell>
                <TableCell>
                  {dep.readyReplicas < dep.replicas ? (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <CircularProgress size={16} />
                      <Typography variant="body2" color="text.secondary">
                        Scaling... {dep.readyReplicas}/{dep.replicas}
                      </Typography>
                    </Box>
                  ) : dep.readyReplicas === dep.replicas && dep.replicas > 0 ? (
                    <Chip label="Ready" color="success" size="small" />
                  ) : (
                    <Chip label="N/A" size="small" />
                  )}
                </TableCell>
                <TableCell align="right" onClick={(e) => e.stopPropagation()}>
                  <Tooltip title="Scale">
                    <IconButton
                      size="small"
                      color="primary"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleScaleClick(dep.namespace, dep.name, dep.replicas)
                      }}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Supprimer">
                    <IconButton
                      size="small"
                      color="error"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDelete('deployment', dep.namespace, dep.name)
                      }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  const renderServicesTable = () => {
    if (!selectedNamespace) {
      return (
        <Alert severity="info" sx={{ mt: 2 }}>
          Veuillez sélectionner un namespace pour afficher les services.
        </Alert>
      )
    }
    if (servicesLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (servicesError) {
      return (
        <Alert severity="error">
          Erreur: {(servicesError as any)?.response?.data?.error || (servicesError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredServices = filterResources(services?.items)
    if (!services?.items || services.items.length === 0) {
      return <Alert severity="info">Aucun service trouvé</Alert>
    }
    if (filteredServices.length === 0) {
      return <Alert severity="info">Aucun service ne correspond à la recherche</Alert>
    }
    const serviceNames = filteredServices.map((s) => s.name)
    const allSelected = serviceNames.length > 0 && serviceNames.every((name) => selectedResources.has(name))
    const someSelected = serviceNames.some((name) => selectedResources.has(name))

    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  indeterminate={someSelected && !allSelected}
                  checked={allSelected}
                  onChange={() => handleSelectAll(serviceNames)}
                />
              </TableCell>
              <TableCell>Nom</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Cluster IP</TableCell>
              <TableCell>Ports</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredServices.map((svc) => (
              <TableRow
                key={svc.name}
                hover
                selected={isSelected(svc.name)}
                sx={{ cursor: 'pointer' }}
                onClick={() =>
                  setDetailDialog({
                    open: true,
                    type: 'service',
                    namespace: svc.namespace,
                    name: svc.name,
                  })
                }
              >
                <TableCell padding="checkbox" onClick={(e) => e.stopPropagation()}>
                  <Checkbox
                    checked={isSelected(svc.name)}
                    onChange={() => handleSelectOne(svc.name)}
                  />
                </TableCell>
                <TableCell>{svc.name}</TableCell>
                <TableCell>{svc.namespace}</TableCell>
                <TableCell>
                  <Chip label={svc.type} size="small" />
                </TableCell>
                <TableCell>{svc.clusterIP || '-'}</TableCell>
                <TableCell>
                  {svc.ports?.map((p, idx) => (
                    <Chip key={idx} label={`${p.port}/${p.protocol}`} size="small" sx={{ mr: 0.5 }} />
                  )) || '-'}
                </TableCell>
                <TableCell align="right" onClick={(e) => e.stopPropagation()}>
                  <Tooltip title="Supprimer">
                    <IconButton
                      size="small"
                      color="error"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDelete('service', svc.namespace, svc.name)
                      }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  const renderConfigMapsTable = () => {
    if (!selectedNamespace) {
      return (
        <Alert severity="info" sx={{ mt: 2 }}>
          Veuillez sélectionner un namespace pour afficher les ConfigMaps.
        </Alert>
      )
    }
    if (configMapsLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (configMapsError) {
      return (
        <Alert severity="error">
          Erreur: {(configMapsError as any)?.response?.data?.error || (configMapsError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredConfigMaps = filterResources(configMaps?.items)
    if (!configMaps?.items || configMaps.items.length === 0) {
      return <Alert severity="info">Aucun ConfigMap trouvé</Alert>
    }
    if (filteredConfigMaps.length === 0) {
      return <Alert severity="info">Aucun ConfigMap ne correspond à la recherche</Alert>
    }
    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Nom</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Clés</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredConfigMaps.map((cm) => (
              <TableRow key={cm.name}>
                <TableCell>{cm.name}</TableCell>
                <TableCell>{cm.namespace}</TableCell>
                <TableCell>
                  {cm.dataKeys.map((key, idx) => (
                    <Chip key={idx} label={key} size="small" sx={{ mr: 0.5 }} />
                  ))}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  const renderSecretsTable = () => {
    if (!selectedNamespace) {
      return (
        <Alert severity="info" sx={{ mt: 2 }}>
          Veuillez sélectionner un namespace pour afficher les Secrets.
        </Alert>
      )
    }
    if (secretsLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (secretsError) {
      return (
        <Alert severity="error">
          Erreur: {(secretsError as any)?.response?.data?.error || (secretsError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredSecrets = filterResources(secrets?.items)
    if (!secrets?.items || secrets.items.length === 0) {
      return <Alert severity="info">Aucun secret trouvé</Alert>
    }
    if (filteredSecrets.length === 0) {
      return <Alert severity="info">Aucun secret ne correspond à la recherche</Alert>
    }
    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Nom</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Clés</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredSecrets.map((secret) => (
              <TableRow key={secret.name}>
                <TableCell>{secret.name}</TableCell>
                <TableCell>{secret.namespace}</TableCell>
                <TableCell>
                  <Chip label={secret.type} size="small" />
                </TableCell>
                <TableCell>
                  {secret.dataKeys.map((key, idx) => (
                    <Chip key={idx} label={key} size="small" sx={{ mr: 0.5 }} />
                  ))}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  const renderClustersTable = () => {
    if (clustersLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }

    if (clustersError) {
      return (
        <Alert severity="error" sx={{ mb: 2 }}>
          Erreur lors du chargement des clusters : {clustersError instanceof Error ? clustersError.message : 'Erreur inconnue'}
        </Alert>
      )
    }

    if (!clusters || clusters.items.length === 0) {
      return (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              Aucun cluster Kubernetes configuré
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Ajoutez un cluster Kubernetes pour commencer à gérer vos ressources.
            </Typography>
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenClusterDialog()}>
              Ajouter un cluster
            </Button>
          </CardContent>
        </Card>
      )
    }

    return (
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">Clusters Kubernetes</Typography>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenClusterDialog()}>
            Ajouter un cluster
          </Button>
        </Box>
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
                            onClick={() => handleActivateCluster(cluster.id)}
                            disabled={activateClusterMutation.isPending}
                          >
                            <CheckCircleIcon />
                          </IconButton>
                        </Tooltip>
                      )}
                      <Tooltip title="Modifier">
                        <IconButton size="small" onClick={() => handleOpenClusterDialog(cluster)}>
                          <EditIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Supprimer">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDeleteCluster(cluster.id)}
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
      </Box>
    )
  }

  const renderNodesTable = () => {
    if (nodesLoading) {
      return (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      )
    }
    if (nodesError) {
      return (
        <Alert severity="error">
          Erreur: {(nodesError as any)?.response?.data?.error || (nodesError as any)?.message || 'Erreur inconnue'}
        </Alert>
      )
    }
    const filteredNodes = filterResources(nodes?.items)
    if (!nodes?.items || nodes.items.length === 0) {
      return <Alert severity="info">Aucun node trouvé</Alert>
    }
    if (filteredNodes.length === 0) {
      return <Alert severity="info">Aucun node ne correspond à la recherche</Alert>
    }
    return (
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Nom</TableCell>
              <TableCell>Statut</TableCell>
              <TableCell>Version</TableCell>
              <TableCell>CPU</TableCell>
              <TableCell>Mémoire</TableCell>
              <TableCell>Pods</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredNodes.map((node) => (
              <TableRow key={node.name}>
                <TableCell>{node.name}</TableCell>
                <TableCell>
                  <Chip label={node.status} color={getStatusColor(node.status) as any} size="small" />
                </TableCell>
                <TableCell>{node.kubeletVersion}</TableCell>
                <TableCell>{node.cpu}</TableCell>
                <TableCell>{node.memory}</TableCell>
                <TableCell>{node.pods}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 600 }}>
        Kubernetes
      </Typography>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>
                Namespaces
              </Typography>
              {namespacesLoading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                  <CircularProgress />
                </Box>
              ) : namespacesError ? (
                <Alert 
                  severity="error"
                  action={
                    <Button 
                      color="inherit" 
                      size="small" 
                      onClick={() => setActiveTab('clusters')}
                    >
                      Configurer un cluster
                    </Button>
                  }
                >
                  {(namespacesError as any)?.response?.status === 503 || 
                   (namespacesError as any)?.response?.data?.error?.includes('Aucun cluster') ? (
                    <>
                      Aucun cluster Kubernetes configuré. Veuillez ajouter un cluster pour commencer.
                    </>
                  ) : (
                    <>
                      Erreur: {(namespacesError as any)?.response?.data?.error || (namespacesError as any)?.message || 'Erreur inconnue'}
                    </>
                  )}
                </Alert>
              ) : namespaces?.items && namespaces.items.length > 0 ? (
                <Box>
                  {namespaces.items.map((ns) => (
                    <Chip
                      key={ns.name}
                      label={ns.name}
                      onClick={() => setSelectedNamespace(ns.name)}
                      sx={{ m: 0.5 }}
                      color={selectedNamespace === ns.name ? 'primary' : 'default'}
                    />
                  ))}
                </Box>
              ) : (
                <Alert severity="info">Aucun namespace trouvé</Alert>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <FormControl fullWidth>
                <InputLabel>Sélectionner un namespace</InputLabel>
                <Select
                  value={selectedNamespace}
                  label="Sélectionner un namespace"
                  onChange={(e) => setSelectedNamespace(e.target.value)}
                >
                  {namespaces?.items?.map((ns) => (
                    <MenuItem key={ns.name} value={ns.name}>
                      {ns.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Card>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider', flexGrow: 1 }}>
              <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
                <Tab label="Clusters" value="clusters" />
                <Tab label="Pods" value="pods" />
                <Tab label="Deployments" value="deployments" />
                <Tab label="Services" value="services" />
                <Tab label="ConfigMaps" value="configmaps" />
                <Tab label="Secrets" value="secrets" />
                <Tab label="Nodes" value="nodes" />
              </Tabs>
            </Box>
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
              {activeTab === 'clusters' && (
                <Button
                  variant="contained"
                  startIcon={<AddIcon />}
                  onClick={() => handleOpenClusterDialog()}
                  sx={{ ml: 2 }}
                >
                  Ajouter un cluster
                </Button>
              )}
              <TextField
                placeholder="Rechercher..."
                size="small"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                InputProps={{
                  startAdornment: <SearchIcon sx={{ mr: 1, color: 'text.secondary' }} />,
                }}
                sx={{ minWidth: 250 }}
                disabled={activeTab === 'clusters'}
              />
            </Box>
          </Box>

          {activeTab === 'pods' && renderPodsTable()}
          {activeTab === 'deployments' && renderDeploymentsTable()}
          {activeTab === 'services' && renderServicesTable()}
          {activeTab === 'configmaps' && renderConfigMapsTable()}
          {activeTab === 'secrets' && renderSecretsTable()}
          {activeTab === 'clusters' && renderClustersTable()}
          {activeTab === 'nodes' && renderNodesTable()}
        </CardContent>
      </Card>

      <ResourceDetailDialog
        open={detailDialog.open}
        onClose={() => setDetailDialog({ ...detailDialog, open: false })}
        resourceType={detailDialog.type}
        namespace={detailDialog.namespace}
        name={detailDialog.name}
      />

      <Dialog open={scaleDialog.open} onClose={() => setScaleDialog({ ...scaleDialog, open: false })}>
        <DialogTitle>Scale Deployment</DialogTitle>
        <DialogContent>
          <TextField
            label="Nombre de replicas"
            type="number"
            fullWidth
            value={scaleReplicas}
            onChange={(e) => {
              const val = parseInt(e.target.value, 10)
              if (!isNaN(val) && val >= 0) {
                setScaleReplicas(val)
              } else if (e.target.value === '') {
                setScaleReplicas(0)
              }
            }}
            inputProps={{ min: 0, step: 1 }}
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setScaleDialog({ ...scaleDialog, open: false })}>Annuler</Button>
          <Button
            onClick={handleScale}
            variant="contained"
            disabled={scaleDeploymentMutation.isPending}
          >
            {scaleDeploymentMutation.isPending ? <CircularProgress size={20} /> : 'Appliquer'}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>

      {/* Barre d'actions en masse */}
      {selectedResources.size > 0 && (
        <AppBar
          position="fixed"
          sx={{
            top: 'auto',
            bottom: 0,
            backgroundColor: 'rgba(10, 10, 14, 0.95)',
            backdropFilter: 'blur(20px)',
            borderTop: '2px solid rgba(0, 255, 255, 0.3)',
            boxShadow: '0 -4px 20px rgba(0, 255, 255, 0.2)',
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', p: 2 }}>
            <Typography variant="body1" sx={{ color: '#FFFFFF', fontWeight: 600 }}>
              {selectedResources.size} ressource{selectedResources.size > 1 ? 's' : ''} sélectionnée{selectedResources.size > 1 ? 's' : ''}
            </Typography>
            <Box sx={{ display: 'flex', gap: 2 }}>
              {activeTab === 'pods' && (
                <Tooltip title="Redémarrer les pods sélectionnés">
                  <Button
                    variant="outlined"
                    startIcon={<RefreshIcon />}
                    onClick={() => {
                      setBulkActionDialog({ open: true, action: 'restart', count: selectedResources.size })
                    }}
                    sx={{
                      borderColor: '#00FFFF',
                      color: '#00FFFF',
                      '&:hover': {
                        borderColor: '#00FFFF',
                        backgroundColor: 'rgba(0, 255, 255, 0.1)',
                      },
                    }}
                  >
                    Redémarrer
                  </Button>
                </Tooltip>
              )}
              {activeTab === 'deployments' && (
                <Tooltip title="Scale les deployments sélectionnés">
                  <Button
                    variant="outlined"
                    startIcon={<EditIcon />}
                    onClick={() => {
                      setBulkActionDialog({ open: true, action: 'scale', count: selectedResources.size })
                    }}
                    sx={{
                      borderColor: '#00FFFF',
                      color: '#00FFFF',
                      '&:hover': {
                        borderColor: '#00FFFF',
                        backgroundColor: 'rgba(0, 255, 255, 0.1)',
                      },
                    }}
                  >
                    Scale
                  </Button>
                </Tooltip>
              )}
              <Button
                variant="contained"
                color="error"
                startIcon={<DeleteIcon />}
                onClick={() => {
                  setBulkActionDialog({ open: true, action: 'delete', count: selectedResources.size })
                }}
                sx={{
                  backgroundColor: '#FF1744',
                  '&:hover': {
                    backgroundColor: '#D50000',
                  },
                }}
              >
                Supprimer
              </Button>
              <Button
                variant="outlined"
                onClick={() => setSelectedResources(new Set())}
                sx={{
                  borderColor: '#A0A0A0',
                  color: '#A0A0A0',
                  '&:hover': {
                    borderColor: '#FFFFFF',
                    color: '#FFFFFF',
                  },
                }}
              >
                Annuler
              </Button>
            </Box>
          </Box>
        </AppBar>
      )}

      {/* Dialog de confirmation pour actions en masse */}
      <Dialog
        open={bulkActionDialog.open}
        onClose={() => setBulkActionDialog({ ...bulkActionDialog, open: false })}
      >
        <DialogTitle>
          Confirmer l'action en masse
        </DialogTitle>
        <DialogContent>
          <Typography>
            {bulkActionDialog.action === 'delete' && (
              <>Êtes-vous sûr de vouloir supprimer {bulkActionDialog.count} ressource{bulkActionDialog.count > 1 ? 's' : ''} ? Cette action est irréversible.</>
            )}
            {bulkActionDialog.action === 'restart' && (
              <>Êtes-vous sûr de vouloir redémarrer {bulkActionDialog.count} pod{bulkActionDialog.count > 1 ? 's' : ''} ?</>
            )}
            {bulkActionDialog.action === 'scale' && (
              <>Vous allez modifier le nombre de replicas de {bulkActionDialog.count} deployment{bulkActionDialog.count > 1 ? 's' : ''}.</>
            )}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBulkActionDialog({ ...bulkActionDialog, open: false })}>
            Annuler
          </Button>
          <Button
            variant="contained"
            color={bulkActionDialog.action === 'delete' ? 'error' : 'primary'}
            onClick={async () => {
              const selectedNames = Array.from(selectedResources)
              
              try {
                let response: { success: string[]; failed: Record<string, string>; total: number } | undefined
                
                if (bulkActionDialog.action === 'delete') {
                  if (activeTab === 'pods') {
                    response = await k8sService.bulkDeletePods(selectedNamespace, selectedNames)
                  } else if (activeTab === 'deployments') {
                    response = await k8sService.bulkDeleteDeployments(selectedNamespace, selectedNames)
                  } else if (activeTab === 'services') {
                    response = await k8sService.bulkDeleteServices(selectedNamespace, selectedNames)
                  }
                } else if (bulkActionDialog.action === 'restart' && activeTab === 'pods') {
                  response = await k8sService.bulkRestartPods(selectedNamespace, selectedNames)
                } else if (bulkActionDialog.action === 'scale' && activeTab === 'deployments') {
                  // Pour le scale, on ouvre un dialog pour demander le nombre de replicas
                  // Pour l'instant, on utilise un scale simple (à améliorer)
                  const replicas = 1 // TODO: Ajouter un input pour le nombre de replicas
                  response = await k8sService.bulkScaleDeployments(selectedNamespace, selectedNames, replicas)
                }

                if (response) {
                  const successCount = response.success.length
                  const failedCount = Object.keys(response.failed).length
                  
                  if (failedCount === 0) {
                    setSnackbar({
                      open: true,
                      message: `${successCount} ressource${successCount > 1 ? 's' : ''} ${bulkActionDialog.action === 'delete' ? 'supprimée' : bulkActionDialog.action === 'restart' ? 'redémarrée' : 'scalée'}${successCount > 1 ? 's' : ''} avec succès`,
                      severity: 'success',
                    })
                  } else if (successCount > 0) {
                    setSnackbar({
                      open: true,
                      message: `${successCount} ressource${successCount > 1 ? 's' : ''} réussie${successCount > 1 ? 's' : ''}, ${failedCount} échec${failedCount > 1 ? 's' : ''}`,
                      severity: 'error',
                    })
                  } else {
                    setSnackbar({
                      open: true,
                      message: `Échec de l'action sur toutes les ressources`,
                      severity: 'error',
                    })
                  }
                  
                  // Rafraîchir les données
                  queryClient.invalidateQueries({ queryKey: [activeTab, selectedNamespace] })
                }
                
                setBulkActionDialog({ ...bulkActionDialog, open: false })
                setSelectedResources(new Set())
              } catch (error: any) {
                setSnackbar({
                  open: true,
                  message: `Erreur: ${error?.response?.data?.error || error?.message || 'Erreur inconnue'}`,
                  severity: 'error',
                })
              }
            }}
          >
            Confirmer
          </Button>
        </DialogActions>
      </Dialog>

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
            onClick={handleClusterSubmit}
            variant="contained"
            disabled={!clusterForm.name || !clusterForm.kubeconfig || createClusterMutation.isPending || updateClusterMutation.isPending}
          >
            {editingCluster ? 'Modifier' : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
