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
  Drawer,
  FormControlLabel,
  Switch,
  Divider,
  InputAdornment,
  Tabs,
  Tab,
  Autocomplete,
} from '@mui/material'
import {
  Refresh as RefreshIcon,
  Sync as SyncIcon,
  Add as AddIcon,
  History as HistoryIcon,
  Replay as RollbackIcon,
  Search as SearchIcon,
  VerifiedUser as VerifiedUserIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import HelmIcon from '../components/icons/HelmIcon'
import ZotIcon from '../components/icons/ZotIcon'
import { ModuleSubtitle, ModuleSecondaryText } from '../components/ModuleText'
import { argocdService, ArgoApplication, CreateApplicationRequest, GitOpsInfo, HelmChartSummary } from '../services/argocdService'
import { registryService } from '../services/registryService'
import { kuraColors } from '../theme'

const emptyForm: CreateApplicationRequest = {
  name: '',
  project: 'default',
  source_type: 'git',
  repo_url: '',
  path: '',
  chart: '',
  helm_values: '',
  target_revision: 'HEAD',
  dest_namespace: '',
  dest_server: 'https://kubernetes.default.svc',
  sync_policy_automated: false,
  prune: false,
  self_heal: false,
  branch: '',
  create_branch_from: '',
}

const NEW_BRANCH_VALUE = '__new_branch__'

function syncStatusColor(status: string): string {
  switch (status) {
    case 'Synced':
      return kuraColors.success
    case 'OutOfSync':
      return kuraColors.warning
    default:
      return kuraColors.text2
  }
}

function healthStatusColor(status: string): string {
  switch (status) {
    case 'Healthy':
      return kuraColors.success
    case 'Degraded':
      return kuraColors.error
    case 'Progressing':
      return kuraColors.info
    default:
      return kuraColors.text2
  }
}

interface BranchSelectorProps {
  gitopsInfo?: GitOpsInfo
  loading: boolean
  branch: string
  newBranchName: string
  createBranchFrom: string
  onChangeBranch: (branch: string) => void
  onChangeNewBranchName: (name: string) => void
  onChangeCreateBranchFrom: (branch: string) => void
}

function BranchSelector({
  gitopsInfo,
  loading,
  branch,
  newBranchName,
  createBranchFrom,
  onChangeBranch,
  onChangeNewBranchName,
  onChangeCreateBranchFrom,
}: BranchSelectorProps) {
  const branches = gitopsInfo?.branches ?? []
  const isNewBranch = branch === NEW_BRANCH_VALUE
  const options = [...branches, NEW_BRANCH_VALUE]

  return (
    <Box sx={{ mb: 2 }}>
      <Autocomplete
        options={options}
        value={branch || null}
        loading={loading}
        getOptionLabel={(option) => (option === NEW_BRANCH_VALUE ? 'Nouvelle branche...' : option)}
        onChange={(_, value) => onChangeBranch(value ?? '')}
        renderInput={(params) => (
          <TextField
            {...params}
            label="Branche du dépôt GitOps"
            size="small"
            helperText={gitopsInfo?.repository ? `Dépôt : ${gitopsInfo.repository}` : undefined}
          />
        )}
      />
      {isNewBranch && (
        <>
          <TextField
            label="Nom de la nouvelle branche"
            value={newBranchName}
            onChange={(e) => onChangeNewBranchName(e.target.value)}
            size="small"
            fullWidth
            sx={{ mt: 2 }}
          />
          <Autocomplete
            options={branches}
            value={createBranchFrom || null}
            onChange={(_, value) => onChangeCreateBranchFrom(value ?? '')}
            sx={{ mt: 2 }}
            renderInput={(params) => (
              <TextField {...params} label="...à partir de la branche" size="small" />
            )}
          />
        </>
      )}
    </Box>
  )
}

function StatusChip({ label, color }: { label: string; color: string }) {
  return (
    <Chip
      label={label || 'Inconnu'}
      size="small"
      sx={{
        fontSize: '0.6875rem',
        bgcolor: `${color}22`,
        border: `1px solid ${color}`,
        color,
      }}
    />
  )
}

export default function ArgoCDPage() {
  const queryClient = useQueryClient()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [helmCatalogOpen, setHelmCatalogOpen] = useState(false)
  const [helmCatalogQuery, setHelmCatalogQuery] = useState('')
  const [helmCatalogTab, setHelmCatalogTab] = useState<'artifacthub' | 'registry'>('artifacthub')
  const [form, setForm] = useState<CreateApplicationRequest>(emptyForm)
  const [installBranch, setInstallBranch] = useState('')
  const [installNewBranchName, setInstallNewBranchName] = useState('')
  const [installCreateBranchFrom, setInstallCreateBranchFrom] = useState('')
  const [formNewBranchName, setFormNewBranchName] = useState('')
  const [detailApp, setDetailApp] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  const { data: status, isLoading: statusLoading } = useQuery({
    queryKey: ['argocd-status'],
    queryFn: () => argocdService.getStatus(),
    retry: false,
    refetchInterval: (query) => (query.state.data?.server_ready ? false : 5000),
  })

  const installed = status?.installed ?? false
  const serverReady = status?.server_ready ?? false
  const selfManaged = status?.self_managed ?? false

  const {
    data: applications,
    isLoading: appsLoading,
    error: appsError,
    refetch: refetchApps,
  } = useQuery({
    queryKey: ['argocd-applications'],
    queryFn: () => argocdService.listApplications(),
    enabled: serverReady,
    retry: false,
  })

  const { data: appDetail } = useQuery({
    queryKey: ['argocd-application-detail', detailApp],
    queryFn: () => argocdService.getApplication(detailApp as string),
    enabled: !!detailApp,
    retry: false,
  })

  const { data: helmCharts, isLoading: helmCatalogLoading } = useQuery({
    queryKey: ['argocd-helm-catalog', helmCatalogQuery],
    queryFn: () => argocdService.searchHelmCatalog(helmCatalogQuery),
    enabled: helmCatalogOpen && helmCatalogTab === 'artifacthub',
    retry: false,
  })

  const { data: gitopsInfo, isLoading: gitopsInfoLoading } = useQuery({
    queryKey: ['argocd-gitops-info'],
    queryFn: () => argocdService.getGitOpsInfo(),
    retry: false,
  })

  const { data: registryCharts, isLoading: registryCatalogLoading } = useQuery({
    queryKey: ['argocd-registry-catalog'],
    queryFn: async () => {
      const repos = await registryService.listRepositories()
      const details = await Promise.all(repos.map((repo) => registryService.getRepository(repo.name)))
      return details
        .map((detail) => {
          const helmTags = detail.tags.filter((tag) => tag.type === 'helm-chart')
          if (helmTags.length === 0) return null
          const latest = helmTags.find((tag) => tag.name === 'latest') ?? helmTags[0]
          return { name: detail.name, tags: helmTags, latest, signed: latest.signed }
        })
        .filter((entry): entry is NonNullable<typeof entry> => entry !== null)
    },
    enabled: helmCatalogOpen && helmCatalogTab === 'registry',
    retry: false,
  })

  const installMutation = useMutation({
    mutationFn: () =>
      argocdService.installArgoCD(
        installBranch === NEW_BRANCH_VALUE ? installNewBranchName : installBranch,
        installBranch === NEW_BRANCH_VALUE ? installCreateBranchFrom : undefined
      ),
    onSuccess: () => {
      setSnackbar({ open: true, message: "Installation d'ArgoCD lancée", severity: 'success' })
      queryClient.invalidateQueries({ queryKey: ['argocd-status'] })
      queryClient.invalidateQueries({ queryKey: ['argocd-gitops-info'] })
    },
    onError: () => {
      setSnackbar({ open: true, message: "Erreur lors de l'installation d'ArgoCD", severity: 'error' })
    },
  })

  const createMutation = useMutation({
    mutationFn: (req: CreateApplicationRequest) => argocdService.createApplication(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['argocd-applications'] })
      queryClient.invalidateQueries({ queryKey: ['argocd-gitops-info'] })
      setCreateDialogOpen(false)
      setForm(emptyForm)
      setFormNewBranchName('')
      setSnackbar({ open: true, message: 'Application créée avec succès', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      const truncated = errorMessage.length > 400 ? `${errorMessage.slice(0, 400)}...` : errorMessage
      setSnackbar({ open: true, message: `Erreur lors de la création de l'Application : ${truncated}`, severity: 'error' })
    },
  })

  const syncMutation = useMutation({
    mutationFn: ({ name, prune }: { name: string; prune?: boolean }) => argocdService.syncApplication(name, prune),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['argocd-applications'] })
      queryClient.invalidateQueries({ queryKey: ['argocd-application-detail'] })
      setSnackbar({ open: true, message: 'Synchronisation lancée', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur lors de la synchronisation : ${errorMessage}`, severity: 'error' })
    },
  })

  const refreshMutation = useMutation({
    mutationFn: (name: string) => argocdService.refreshApplication(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['argocd-applications'] })
      queryClient.invalidateQueries({ queryKey: ['argocd-application-detail'] })
      setSnackbar({ open: true, message: 'Rafraîchissement effectué', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur lors du rafraîchissement : ${errorMessage}`, severity: 'error' })
    },
  })

  const rollbackMutation = useMutation({
    mutationFn: ({ name, id }: { name: string; id: number }) => argocdService.rollbackApplication(name, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['argocd-applications'] })
      queryClient.invalidateQueries({ queryKey: ['argocd-application-detail'] })
      setSnackbar({ open: true, message: 'Retour à la version précédente effectué', severity: 'success' })
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur lors du rollback : ${errorMessage}`, severity: 'error' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (name: string) => argocdService.deleteApplication(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['argocd-applications'] })
      setSnackbar({ open: true, message: 'Application supprimée', severity: 'success' })
      setDeleteTarget(null)
      setDetailApp(null)
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
      setSnackbar({ open: true, message: `Erreur lors de la suppression de l'Application : ${errorMessage}`, severity: 'error' })
      setDeleteTarget(null)
    },
  })

  const handleCreate = () => {
    if (form.source_type === 'helm') {
      if (!form.name.trim() || !form.repo_url.trim() || !form.chart?.trim() || !form.dest_namespace.trim()) {
        setSnackbar({ open: true, message: 'Nom, dépôt Helm, chart et namespace de destination sont requis', severity: 'error' })
        return
      }
    } else if (!form.name.trim() || !form.repo_url.trim() || !form.path.trim() || !form.dest_namespace.trim()) {
      setSnackbar({ open: true, message: 'Nom, dépôt, chemin et namespace de destination sont requis', severity: 'error' })
      return
    }
    if (!form.branch || (form.branch === NEW_BRANCH_VALUE && !formNewBranchName.trim())) {
      setSnackbar({ open: true, message: 'Une branche du dépôt GitOps doit être sélectionnée', severity: 'error' })
      return
    }
    const branch = form.branch === NEW_BRANCH_VALUE ? formNewBranchName.trim() : form.branch
    const createBranchFrom = form.branch === NEW_BRANCH_VALUE ? form.create_branch_from : undefined
    createMutation.mutate({ ...form, branch, create_branch_from: createBranchFrom })
  }

  const handleSelectHelmChart = (chart: HelmChartSummary) => {
    setForm({
      ...emptyForm,
      name: chart.name,
      source_type: 'helm',
      repo_url: chart.repo_url,
      chart: chart.name,
      target_revision: chart.version,
      dest_namespace: chart.name,
      helm_values: '',
    })
    setFormNewBranchName('')
    setHelmCatalogOpen(false)
    setCreateDialogOpen(true)
  }

  const handleSelectRegistryChart = (entry: { name: string; latest: { name: string } }) => {
    setForm({
      ...emptyForm,
      name: entry.name,
      source_type: 'helm',
      repo_url: 'oci://zot.zot.svc.cluster.local:5000',
      chart: entry.name,
      target_revision: entry.latest.name,
      dest_namespace: entry.name,
      helm_values: '',
    })
    setFormNewBranchName('')
    setHelmCatalogOpen(false)
    setCreateDialogOpen(true)
  }

  return (
    <Box>
      <ModuleTitle>ArgoCD</ModuleTitle>

      <ModuleCard sx={{ mb: 2 }}>
        <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography sx={{ fontWeight: 600 }}>État ArgoCD</Typography>
            {statusLoading ? (
              <CircularProgress size={16} />
            ) : (
              <>
                <Chip
                  label={installed ? 'Installé' : 'Non installé'}
                  size="small"
                  sx={{
                    color: installed ? kuraColors.success : kuraColors.warning,
                    borderColor: installed ? kuraColors.success : kuraColors.warning,
                  }}
                  variant="outlined"
                />
                {installed && (
                  <Chip
                    label={serverReady ? 'Serveur prêt' : 'Démarrage en cours...'}
                    size="small"
                    sx={{
                      color: serverReady ? kuraColors.success : kuraColors.info,
                      borderColor: serverReady ? kuraColors.success : kuraColors.info,
                    }}
                    variant="outlined"
                  />
                )}
                {status?.version && (
                  <Chip label={`v${status.version}`} size="small" variant="outlined" />
                )}
              </>
            )}
          </Box>
        </Box>
        {!installed && (
          <Box sx={{ px: 2, pb: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Alert severity="info">
              L'installation d'ArgoCD est suivie via le dépôt GitOps du projet : le manifest
              d'auto-gestion d'ArgoCD sera committé sur la branche choisie avant l'installation.
            </Alert>
            <Box sx={{ maxWidth: 480 }}>
              <BranchSelector
                gitopsInfo={gitopsInfo}
                loading={gitopsInfoLoading}
                branch={installBranch}
                newBranchName={installNewBranchName}
                createBranchFrom={installCreateBranchFrom}
                onChangeBranch={setInstallBranch}
                onChangeNewBranchName={setInstallNewBranchName}
                onChangeCreateBranchFrom={setInstallCreateBranchFrom}
              />
            </Box>
            <Box>
              <Button
                variant="contained"
                onClick={() => installMutation.mutate()}
                disabled={
                  installMutation.isPending ||
                  !installBranch ||
                  (installBranch === NEW_BRANCH_VALUE && !installNewBranchName.trim())
                }
              >
                {installMutation.isPending ? 'Installation en cours...' : 'Installer ArgoCD'}
              </Button>
            </Box>
          </Box>
        )}
        {installed && !serverReady && (
          <Box sx={{ px: 2, pb: 2 }}>
            <Alert severity="info">
              ArgoCD est installé, en attente du démarrage du serveur (argocd-server)...
            </Alert>
          </Box>
        )}
        {installed && serverReady && !selfManaged && (
          <Box sx={{ px: 2, pb: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Alert severity="warning">
              ArgoCD est installé mais son auto-gestion GitOps (commit du bootstrap et
              Application "argocd") n'a pas pu être finalisée. Choisissez une branche et
              relancez le bootstrap.
            </Alert>
            <Box sx={{ maxWidth: 480 }}>
              <BranchSelector
                gitopsInfo={gitopsInfo}
                loading={gitopsInfoLoading}
                branch={installBranch}
                newBranchName={installNewBranchName}
                createBranchFrom={installCreateBranchFrom}
                onChangeBranch={setInstallBranch}
                onChangeNewBranchName={setInstallNewBranchName}
                onChangeCreateBranchFrom={setInstallCreateBranchFrom}
              />
            </Box>
            <Box>
              <Button
                variant="contained"
                onClick={() => installMutation.mutate()}
                disabled={
                  installMutation.isPending ||
                  !installBranch ||
                  (installBranch === NEW_BRANCH_VALUE && !installNewBranchName.trim())
                }
              >
                {installMutation.isPending ? 'Bootstrap en cours...' : 'Relancer le bootstrap GitOps'}
              </Button>
            </Box>
          </Box>
        )}
      </ModuleCard>

      {serverReady && (
        <ModuleCard>
          <Box sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <ModuleSubtitle>Applications</ModuleSubtitle>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<HelmIcon sx={{ width: 18, height: 18 }} active />}
                  onClick={() => setHelmCatalogOpen(true)}
                >
                  Depuis le catalogue Helm
                </Button>
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => { setForm(emptyForm); setFormNewBranchName(''); setCreateDialogOpen(true) }}
                >
                  Nouvelle Application
                </Button>
                <IconButton onClick={() => refetchApps()} color="primary">
                  <RefreshIcon />
                </IconButton>
              </Box>
            </Box>

            {appsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : appsError ? (
              <Alert severity="error">Erreur lors de la récupération des Applications.</Alert>
            ) : (applications?.length ?? 0) > 0 ? (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Nom</TableCell>
                      <TableCell>Projet</TableCell>
                      <TableCell>Synchro</TableCell>
                      <TableCell>Santé</TableCell>
                      <TableCell>Source</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {applications!.map((app: ArgoApplication) => {
                      const isSyncing = syncMutation.isPending && syncMutation.variables?.name === app.name
                      const isRefreshing = refreshMutation.isPending && refreshMutation.variables === app.name
                      const isDeleting = deleteMutation.isPending && deleteMutation.variables === app.name
                      const operationInProgress = app.health_status === 'Progressing'
                      const actionsDisabled = isSyncing || isRefreshing || isDeleting || operationInProgress

                      return (
                      <TableRow
                        key={app.name}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() => setDetailApp(app.name)}
                      >
                        <TableCell sx={{ fontWeight: 600 }}>{app.name}</TableCell>
                        <TableCell>{app.project}</TableCell>
                        <TableCell>
                          <StatusChip label={app.sync_status} color={syncStatusColor(app.sync_status)} />
                        </TableCell>
                        <TableCell>
                          <StatusChip label={app.health_status} color={healthStatusColor(app.health_status)} />
                        </TableCell>
                        <TableCell>
                          <ModuleSecondaryText>
                            {app.repo_url} ({app.path})
                          </ModuleSecondaryText>
                        </TableCell>
                        <TableCell align="right">
                          <Tooltip title={operationInProgress ? 'Une opération est déjà en cours' : 'Synchroniser'}>
                            <span>
                              <IconButton
                                size="small"
                                disabled={actionsDisabled}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  syncMutation.mutate({ name: app.name })
                                }}
                              >
                                {isSyncing ? <CircularProgress size={16} /> : <SyncIcon fontSize="small" />}
                              </IconButton>
                            </span>
                          </Tooltip>
                          <Tooltip title="Rafraîchir">
                            <span>
                              <IconButton
                                size="small"
                                disabled={isRefreshing || isDeleting}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  refreshMutation.mutate(app.name)
                                }}
                              >
                                {isRefreshing ? <CircularProgress size={16} /> : <RefreshIcon fontSize="small" />}
                              </IconButton>
                            </span>
                          </Tooltip>
                          <Tooltip title="Historique">
                            <IconButton
                              size="small"
                              onClick={(e) => {
                                e.stopPropagation()
                                setDetailApp(app.name)
                              }}
                            >
                              <HistoryIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Supprimer">
                            <span>
                              <IconButton
                                size="small"
                                disabled={isDeleting}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  setDeleteTarget(app.name)
                                }}
                              >
                                {isDeleting ? <CircularProgress size={16} /> : <DeleteIcon fontSize="small" sx={{ color: kuraColors.error }} />}
                              </IconButton>
                            </span>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Alert severity="info">Aucune Application ArgoCD pour le moment.</Alert>
            )}
          </Box>
        </ModuleCard>
      )}

      {/* Dialog de création d'une Application */}
      <Dialog open={createDialogOpen} onClose={() => { setCreateDialogOpen(false); createMutation.reset() }} maxWidth="sm" fullWidth>
        <DialogTitle>
          {form.source_type === 'helm' ? 'Nouvelle Application ArgoCD — Chart Helm' : 'Nouvelle Application ArgoCD'}
        </DialogTitle>
        <DialogContent>
          {createMutation.isError && (
            <Alert severity="error" sx={{ mb: 2, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
              {(() => {
                const error: any = createMutation.error
                const errorMessage = error?.response?.data?.error || error?.message || 'Erreur inconnue'
                return errorMessage.length > 600 ? `${errorMessage.slice(0, 600)}...` : errorMessage
              })()}
            </Alert>
          )}
          <TextField
            label="Nom"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            size="small"
            fullWidth
            sx={{ mt: 1, mb: 2 }}
          />
          <TextField
            label="Projet"
            value={form.project}
            onChange={(e) => setForm({ ...form, project: e.target.value })}
            size="small"
            fullWidth
            sx={{ mb: 2 }}
          />
          {form.source_type === 'helm' ? (
            <>
              <TextField
                label="URL du dépôt Helm"
                placeholder="https://prometheus-community.github.io/helm-charts"
                value={form.repo_url}
                onChange={(e) => setForm({ ...form, repo_url: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
              <TextField
                label="Chart"
                value={form.chart}
                onChange={(e) => setForm({ ...form, chart: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
              <TextField
                label="Version du chart"
                value={form.target_revision}
                onChange={(e) => setForm({ ...form, target_revision: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
              <TextField
                label="Values (YAML)"
                value={form.helm_values}
                onChange={(e) => setForm({ ...form, helm_values: e.target.value })}
                size="small"
                fullWidth
                multiline
                minRows={6}
                placeholder={'# Surcharges values.yaml\nreplicaCount: 1'}
                sx={{ mb: 2, '& textarea': { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8125rem' } }}
              />
            </>
          ) : (
            <>
              <TextField
                label="URL du dépôt Git"
                placeholder="https://github.com/org/repo.git"
                value={form.repo_url}
                onChange={(e) => setForm({ ...form, repo_url: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
              <TextField
                label="Chemin"
                placeholder="manifests/"
                value={form.path}
                onChange={(e) => setForm({ ...form, path: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
              <TextField
                label="Révision cible"
                value={form.target_revision}
                onChange={(e) => setForm({ ...form, target_revision: e.target.value })}
                size="small"
                fullWidth
                sx={{ mb: 2 }}
              />
            </>
          )}
          <TextField
            label="Namespace de destination"
            value={form.dest_namespace}
            onChange={(e) => setForm({ ...form, dest_namespace: e.target.value })}
            size="small"
            fullWidth
            sx={{ mb: 2 }}
          />
          <TextField
            label="Cluster de destination"
            value={form.dest_server}
            onChange={(e) => setForm({ ...form, dest_server: e.target.value })}
            size="small"
            fullWidth
            sx={{ mb: 2 }}
          />
          <Divider sx={{ my: 1 }} />
          <BranchSelector
            gitopsInfo={gitopsInfo}
            loading={gitopsInfoLoading}
            branch={form.branch}
            newBranchName={formNewBranchName}
            createBranchFrom={form.create_branch_from ?? ''}
            onChangeBranch={(branch) => setForm({ ...form, branch })}
            onChangeNewBranchName={setFormNewBranchName}
            onChangeCreateBranchFrom={(createBranchFrom) => setForm({ ...form, create_branch_from: createBranchFrom })}
          />
          <Divider sx={{ my: 1 }} />
          <FormControlLabel
            control={
              <Switch
                checked={form.sync_policy_automated}
                onChange={(e) => setForm({ ...form, sync_policy_automated: e.target.checked })}
              />
            }
            label="Synchronisation automatique"
          />
          <FormControlLabel
            control={
              <Switch
                checked={form.prune}
                onChange={(e) => setForm({ ...form, prune: e.target.checked })}
                disabled={!form.sync_policy_automated}
              />
            }
            label="Supprimer les ressources obsolètes (prune)"
          />
          <FormControlLabel
            control={
              <Switch
                checked={form.self_heal}
                onChange={(e) => setForm({ ...form, self_heal: e.target.checked })}
                disabled={!form.sync_policy_automated}
              />
            }
            label="Auto-correction (self-heal)"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setCreateDialogOpen(false); createMutation.reset() }}>Annuler</Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={
              createMutation.isPending ||
              !form.branch ||
              (form.branch === NEW_BRANCH_VALUE && !formNewBranchName.trim())
            }
          >
            {createMutation.isPending ? 'Création...' : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Dialog du catalogue Helm (ArtifactHub) */}
      <Dialog open={helmCatalogOpen} onClose={() => setHelmCatalogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <HelmIcon sx={{ width: 24, height: 24 }} active />
          Catalogue Helm
        </DialogTitle>
        <Tabs
          value={helmCatalogTab}
          onChange={(_, value) => setHelmCatalogTab(value)}
          sx={{ px: 3, borderBottom: `1px solid ${kuraColors.border1}` }}
        >
          <Tab label="ArtifactHub" value="artifacthub" />
          <Tab label="Zot" value="registry" />
        </Tabs>
        <DialogContent>
          {helmCatalogTab === 'artifacthub' && (
          <TextField
            placeholder="Rechercher un chart (prometheus, cert-manager, ingress-nginx...)"
            value={helmCatalogQuery}
            onChange={(e) => setHelmCatalogQuery(e.target.value)}
            size="small"
            fullWidth
            sx={{ mt: 1, mb: 2 }}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" />
                </InputAdornment>
              ),
            }}
          />
          )}

          {helmCatalogTab === 'artifacthub' ? (helmCatalogLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (helmCharts?.length ?? 0) > 0 ? (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1.5 }}>
              {helmCharts!.map((chart) => (
                <Box
                  key={chart.package_id}
                  onClick={() => handleSelectHelmChart(chart)}
                  sx={{
                    width: { xs: '100%', sm: 'calc(50% - 6px)' },
                    p: 1.5,
                    borderRadius: '8px',
                    border: `1px solid ${kuraColors.border1}`,
                    cursor: 'pointer',
                    display: 'flex',
                    gap: 1.5,
                    transition: 'border-color 0.15s ease, background-color 0.15s ease',
                    '&:hover': { borderColor: kuraColors.accent, bgcolor: kuraColors.bg3 },
                  }}
                >
                  {chart.logo_url ? (
                    <Box
                      component="img"
                      src={chart.logo_url}
                      alt=""
                      sx={{ width: 40, height: 40, objectFit: 'contain', flexShrink: 0, borderRadius: '4px' }}
                    />
                  ) : (
                    <Box sx={{ width: 40, height: 40, flexShrink: 0, borderRadius: '4px', bgcolor: kuraColors.bg3 }} />
                  )}
                  <Box sx={{ minWidth: 0, flex: 1 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, flexWrap: 'wrap' }}>
                      <Typography sx={{ fontWeight: 600, fontSize: '0.875rem' }}>{chart.display_name || chart.name}</Typography>
                      {chart.cncf && <Chip label="CNCF" size="small" sx={{ fontSize: '0.625rem', height: 18 }} />}
                      {chart.official && <Chip label="Officiel" size="small" sx={{ fontSize: '0.625rem', height: 18 }} />}
                    </Box>
                    <Typography
                      sx={{
                        fontSize: '0.75rem', color: kuraColors.text2, mt: 0.25,
                        overflow: 'hidden', textOverflow: 'ellipsis',
                        display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
                      }}
                    >
                      {chart.description}
                    </Typography>
                    <Typography sx={{ fontSize: '0.6875rem', color: kuraColors.text2, mt: 0.5, fontFamily: '"JetBrains Mono", monospace' }}>
                      {chart.repo_name} · v{chart.version}
                    </Typography>
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info">Aucun chart trouvé pour cette recherche.</Alert>
          )) : registryCatalogLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (registryCharts?.length ?? 0) > 0 ? (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1.5, mt: 1 }}>
              {registryCharts!.map((entry) => (
                <Box
                  key={entry.name}
                  onClick={() => handleSelectRegistryChart(entry)}
                  sx={{
                    width: { xs: '100%', sm: 'calc(50% - 6px)' },
                    p: 1.5,
                    borderRadius: '8px',
                    border: `1px solid ${kuraColors.border1}`,
                    cursor: 'pointer',
                    display: 'flex',
                    gap: 1.5,
                    transition: 'border-color 0.15s ease, background-color 0.15s ease',
                    '&:hover': { borderColor: kuraColors.accent, bgcolor: kuraColors.bg3 },
                  }}
                >
                  <ZotIcon sx={{ width: 40, height: 40 }} active />
                  <Box sx={{ minWidth: 0, flex: 1 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, flexWrap: 'wrap' }}>
                      <Typography sx={{ fontWeight: 600, fontSize: '0.875rem' }}>{entry.name}</Typography>
                      {entry.signed && (
                        <Chip
                          icon={<VerifiedUserIcon sx={{ fontSize: 14 }} />}
                          label="Signé"
                          size="small"
                          sx={{ fontSize: '0.625rem', height: 18, color: kuraColors.success, borderColor: kuraColors.success }}
                          variant="outlined"
                        />
                      )}
                    </Box>
                    <Typography sx={{ fontSize: '0.6875rem', color: kuraColors.text2, mt: 0.5, fontFamily: '"JetBrains Mono", monospace' }}>
                      Zot · {entry.latest.name}
                    </Typography>
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info">Aucun chart Helm trouvé dans le registre interne.</Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setHelmCatalogOpen(false)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Tiroir de détails / historique */}
      <Drawer anchor="right" open={!!detailApp} onClose={() => setDetailApp(null)}>
        <Box sx={{ width: 420, p: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
            <Typography variant="h6">{detailApp}</Typography>
            <Tooltip title="Supprimer">
              <IconButton
                size="small"
                onClick={() => detailApp && setDeleteTarget(detailApp)}
              >
                <DeleteIcon fontSize="small" sx={{ color: kuraColors.error }} />
              </IconButton>
            </Tooltip>
          </Box>
          {appDetail ? (
            <>
              <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
                <StatusChip label={appDetail.sync_status} color={syncStatusColor(appDetail.sync_status)} />
                <StatusChip label={appDetail.health_status} color={healthStatusColor(appDetail.health_status)} />
              </Box>
              <ModuleSecondaryText>Projet : {appDetail.project}</ModuleSecondaryText>
              <ModuleSecondaryText>Dépôt : {appDetail.repo_url}</ModuleSecondaryText>
              <ModuleSecondaryText>Chemin : {appDetail.path}</ModuleSecondaryText>
              <ModuleSecondaryText>Révision : {appDetail.target_revision}</ModuleSecondaryText>
              <ModuleSecondaryText>Namespace : {appDetail.dest_namespace}</ModuleSecondaryText>

              <Divider sx={{ my: 2 }} />
              <ModuleSubtitle>Historique des déploiements</ModuleSubtitle>
              {appDetail.history?.length > 0 ? (
                <Table size="small" sx={{ mt: 1 }}>
                  <TableHead>
                    <TableRow>
                      <TableCell>Révision</TableCell>
                      <TableCell>Date</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {[...appDetail.history].reverse().map((entry) => (
                      <TableRow key={entry.id}>
                        <TableCell sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}>
                          {entry.revision_id?.slice(0, 8)}
                        </TableCell>
                        <TableCell>
                          {entry.deployed_at ? new Date(entry.deployed_at).toLocaleString('fr-FR') : '-'}
                        </TableCell>
                        <TableCell align="right">
                          <Tooltip title="Revenir à cette version">
                            <IconButton
                              size="small"
                              onClick={() =>
                                detailApp && rollbackMutation.mutate({ name: detailApp, id: entry.id })
                              }
                              disabled={rollbackMutation.isPending}
                            >
                              <RollbackIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Alert severity="info" sx={{ mt: 1 }}>Aucun historique disponible.</Alert>
              )}
            </>
          ) : (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress size={24} />
            </Box>
          )}
        </Box>
      </Drawer>

      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>Supprimer l&apos;Application ?</DialogTitle>
        <DialogContent>
          <Typography>
            L&apos;Application <strong>{deleteTarget}</strong> et toutes les ressources qu&apos;elle a déployées dans le cluster (namespace, deployment, service...) seront supprimées. Cette action est irréversible.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)}>Annuler</Button>
          <Button
            color="error"
            variant="contained"
            disabled={deleteMutation.isPending}
            onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget)}
          >
            {deleteMutation.isPending ? 'Suppression...' : 'Supprimer'}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={snackbar.severity === 'error' ? 15000 : 4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        sx={{ top: { xs: 16, sm: 24 }, zIndex: (theme) => theme.zIndex.modal + 1 }}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          sx={{ maxWidth: 600, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  )
}
