import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Box,
  Alert,
  CircularProgress,
  Tabs,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Snackbar,
} from '@mui/material'
import {
  Refresh as RefreshIcon,
  PlayArrow as PlayIcon,
  Visibility as ViewIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
  Schedule as ScheduleIcon,
  Link as LinkIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import {
  ansibleService,
  AnsibleJobSummary,
  AnsibleInventorySummary,
  AnsibleJobTemplateSummary,
  AnsibleJobTemplateDetail,
  AnsibleInventoryDetail,
  AnsibleHost,
} from '../services/ansibleService'
import { ModuleSubtitle, ModuleBodyText, ModuleSecondaryText } from '../components/ModuleText'
import CodeBlock from '../components/CodeBlock'

interface TabPanelProps {
  children?: React.ReactNode
  index: number
  value: number
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props
  return (
    <div role="tabpanel" hidden={value !== index} {...other}>
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  )
}

export default function AnsiblePage() {
  const queryClient = useQueryClient()
  const [configExpanded, setConfigExpanded] = useState(false)
  const [semaphoreUrl, setSemaphoreUrl] = useState('')
  const [semaphoreToken, setSemaphoreToken] = useState('')
  const [semaphoreProjectId, setSemaphoreProjectId] = useState('1')
  const [tabValue, setTabValue] = useState(0)
  const [selectedJob, setSelectedJob] = useState<AnsibleJobSummary | null>(null)
  const [jobDetailDialogOpen, setJobDetailDialogOpen] = useState(false)
  const [selectedInventory, setSelectedInventory] = useState<AnsibleInventorySummary | null>(null)
  const [inventoryDetail, setInventoryDetail] = useState<AnsibleInventoryDetail | null>(null)
  const [inventoryHosts, setInventoryHosts] = useState<AnsibleHost[]>([])
  const [inventoryDetailDialogOpen, setInventoryDetailDialogOpen] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<AnsibleJobTemplateSummary | null>(null)
  const [templateDetail, setTemplateDetail] = useState<AnsibleJobTemplateDetail | null>(null)
  const [templateDetailDialogOpen, setTemplateDetailDialogOpen] = useState(false)
  const [jobStdout, setJobStdout] = useState<string | null>(null)
  const [loadingStdout, setLoadingStdout] = useState(false)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  // Configuration Semaphore
  const { data: configData, refetch: refetchConfig } = useQuery({
    queryKey: ['ansible-config'],
    queryFn: () => ansibleService.getConfig(),
    retry: false,
  })

  const saveConfigMutation = useMutation({
    mutationFn: (data: { semaphore_url?: string; token?: string; semaphore_project_id?: number }) =>
      ansibleService.setConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ansible-config'] })
      queryClient.invalidateQueries({ queryKey: ['ansible-jobs'] })
      refetchConfig()
      setSemaphoreToken('')
      setConfigExpanded(false)
    },
  })

  const handleSaveConfig = () => {
    saveConfigMutation.mutate({
      ...(semaphoreUrl && { semaphore_url: semaphoreUrl }),
      ...(semaphoreToken && { token: semaphoreToken }),
      semaphore_project_id: parseInt(semaphoreProjectId) || 1,
    })
  }

  // Requêtes pour les données
  const {
    data: jobsData,
    isLoading: jobsLoading,
    isError: jobsError,
    refetch: refetchJobs,
  } = useQuery({
    queryKey: ['ansible-jobs'],
    queryFn: () => ansibleService.getJobs(),
    refetchInterval: 30000,
    retry: false, // Ne pas réessayer en boucle si l'API n'est pas joignable
  })

  const {
    data: inventoriesData,
    isLoading: inventoriesLoading,
    refetch: refetchInventories,
  } = useQuery({
    queryKey: ['ansible-inventories'],
    queryFn: () => ansibleService.getInventories(),
  })

  const {
    data: templatesData,
    isLoading: templatesLoading,
    refetch: refetchTemplates,
  } = useQuery({
    queryKey: ['ansible-templates'],
    queryFn: () => ansibleService.getJobTemplates(),
  })

  const {
    data: jobHistoryData,
    isLoading: jobHistoryLoading,
    refetch: refetchJobHistory,
  } = useQuery({
    queryKey: ['ansible-job-history'],
    queryFn: () => ansibleService.getJobHistory(50),
  })

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue)
  }

  const handleViewInventory = async (inventory: AnsibleInventorySummary) => {
    try {
      setSelectedInventory(inventory)
      setInventoryDetailDialogOpen(true)
      const [detail, hostsResponse] = await Promise.all([
        ansibleService.getInventory(inventory.id),
        ansibleService.getInventoryHosts(inventory.id),
      ])
      setInventoryDetail(detail)
      setInventoryHosts(hostsResponse?.items ?? [])
    } catch (error) {
      setSnackbar({ open: true, message: "Erreur lors de la récupération de l'inventaire", severity: 'error' })
    }
  }

  const handleViewJob = async (job: AnsibleJobSummary) => {
    try {
      setJobStdout(null)
      const detail = await ansibleService.getJob(job.id)
      setSelectedJob({ ...job, ...detail } as AnsibleJobSummary)
      setJobDetailDialogOpen(true)
    } catch (error) {
      setSnackbar({ open: true, message: 'Erreur lors de la récupération des détails du job', severity: 'error' })
    }
  }

  const handleLoadJobStdout = async () => {
    if (!selectedJob) return
    try {
      setLoadingStdout(true)
      const detail = await ansibleService.getJob(selectedJob.id, true)
      setJobStdout(detail.stdout || null)
    } catch (error) {
      setSnackbar({ open: true, message: 'Erreur lors du chargement de la sortie', severity: 'error' })
    } finally {
      setLoadingStdout(false)
    }
  }

  const handleViewTemplate = async (template: AnsibleJobTemplateSummary) => {
    try {
      setSelectedTemplate(template)
      setTemplateDetailDialogOpen(true)
      const detail = await ansibleService.getJobTemplate(template.id)
      setTemplateDetail(detail)
    } catch (error) {
      setSnackbar({ open: true, message: 'Erreur lors de la récupération du template', severity: 'error' })
    }
  }

  const handleLaunchTemplate = async (templateId: number) => {
    try {
      const result = await ansibleService.launchJobTemplate(templateId)
      setSnackbar({ open: true, message: `Job lancé avec succès (ID: ${result.job})`, severity: 'success' })
      refetchJobs()
      refetchJobHistory()
    } catch (error: any) {
      setSnackbar({
        open: true,
        message: error.response?.data?.detail || 'Erreur lors du lancement du job',
        severity: 'error',
      })
    }
  }

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'successful': case 'success': return 'success'
      case 'failed': case 'error': return 'error'
      case 'running': case 'pending': return 'info'
      case 'canceled': case 'cancelled': return 'warning'
      default: return 'default'
    }
  }

  const getStatusSx = (status: string) => {
    const s = status?.toLowerCase() || ''
    if (['successful','success'].includes(s))
      return { bgcolor: 'rgba(52,211,153,0.15)', border: '1px solid #34D399', color: '#34D399' }
    if (['failed','error'].includes(s))
      return { bgcolor: 'rgba(248,113,113,0.15)', border: '1px solid #F87171', color: '#F87171' }
    if (['running'].includes(s))
      return { bgcolor: 'rgba(96,165,250,0.15)', border: '1px solid #60A5FA', color: '#60A5FA' }
    if (['pending'].includes(s))
      return { bgcolor: 'rgba(251,191,36,0.15)', border: '1px solid #FBBF24', color: '#FBBF24' }
    if (['canceled','cancelled','stopped'].includes(s))
      return { bgcolor: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.2)', color: '#8C94A6' }
    return { bgcolor: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.15)', color: '#8C94A6' }
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return '-'
    return new Date(dateString).toLocaleString('fr-FR')
  }

  const formatElapsed = (seconds?: number) => {
    if (!seconds) return '-'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}m ${secs}s`
  }

  // Afficher un avertissement si l'API Ansible n'est pas joignable
  const apiUnreachable = jobsError && !templatesData?.items?.length

  // Tri : jobs en cours en premier, puis plus récent en premier
  const sortJobsByDate = (jobs: AnsibleJobSummary[]) =>
    [...jobs].sort((a, b) => {
      const runningA = ['running', 'pending'].includes((a.status || '').toLowerCase()) ? 1 : 0
      const runningB = ['running', 'pending'].includes((b.status || '').toLowerCase()) ? 1 : 0
      if (runningA !== runningB) return runningB - runningA
      const dateA = new Date(a.finished || a.started || 0).getTime()
      const dateB = new Date(b.finished || b.started || 0).getTime()
      return dateB - dateA
    })

  // Jobs = uniquement running/pending | Historique = tous les jobs
  const allJobs = jobHistoryData?.items ?? jobsData?.items ?? []
  const sortedAll = sortJobsByDate(allJobs)
  const runningJobs = sortedAll.filter((j) =>
    ['running', 'pending'].includes((j.status || '').toLowerCase())
  )
  const sortedHistory = sortedAll

  return (
    <Box>
      <ModuleTitle>Semaphore</ModuleTitle>

      {/* Panneau de configuration Semaphore */}
      <ModuleCard sx={{ mb: 2 }}>
        <Box
          sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
          onClick={() => setConfigExpanded(!configExpanded)}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <LinkIcon sx={{ color: 'primary.main' }} />
            <Typography sx={{ fontWeight: 600 }}>Connecter un backend Ansible (Semaphore)</Typography>
            {configData?.configured ? (
              <Chip label="Connecté" size="small" sx={{ color: '#00FF88', borderColor: '#00FF88' }} variant="outlined" />
            ) : (
              <Chip label="Non configuré" size="small" color="warning" variant="outlined" />
            )}
          </Box>
          {configExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Box>

        {configExpanded && (
          <Box sx={{ px: 2, pb: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            {configData?.semaphore_url && (
              <Alert severity="info" sx={{ mb: 1 }}>
                URL actuelle : <strong>{configData.semaphore_url}</strong> — Projet ID : <strong>{configData.semaphore_project_id}</strong>
                {configData.has_token && ' — Token configuré ✓'}
              </Alert>
            )}
            <TextField
              label="URL Semaphore"
              placeholder="http://semaphore:3000"
              value={semaphoreUrl}
              onChange={(e) => setSemaphoreUrl(e.target.value)}
              size="small"
              fullWidth
            />
            <TextField
              label="Token API"
              placeholder="Laisser vide pour conserver l'actuel"
              value={semaphoreToken}
              onChange={(e) => setSemaphoreToken(e.target.value)}
              type="password"
              size="small"
              fullWidth
            />
            <TextField
              label="Project ID"
              placeholder="1"
              value={semaphoreProjectId}
              onChange={(e) => setSemaphoreProjectId(e.target.value)}
              size="small"
              sx={{ width: 120 }}
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

      {apiUnreachable && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          Le service Ansible n'est pas joignable. Vérifiez que Kong (port 8000) et ansible-service (8083) tournent et que
          le frontend utilise <code>VITE_API_BASE_URL=http://localhost:8000</code>.
        </Alert>
      )}

      <ModuleCard>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab label="Jobs" />
              <Tab label="Historique" />
              <Tab label="Inventaires" />
              <Tab label="Templates" />
            </Tabs>
          </Box>

          {/* Onglet Jobs (en cours uniquement) */}
          <TabPanel value={tabValue} index={0}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <ModuleSubtitle>Jobs en cours</ModuleSubtitle>
              <IconButton onClick={() => { refetchJobs(); refetchJobHistory() }} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {(jobHistoryLoading || jobsLoading) ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : runningJobs.length > 0 ? (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>Nom</TableCell>
                      <TableCell>Statut</TableCell>
                      <TableCell>Template</TableCell>
                      <TableCell>Inventaire</TableCell>
                      <TableCell>Démarré</TableCell>
                      <TableCell>Durée</TableCell>
                      <TableCell>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {runningJobs.map((job) => (
                      <TableRow
                        key={job.id}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() => handleViewJob(job)}
                      >
                        <TableCell>{job.id}</TableCell>
                        <TableCell>{job.name}</TableCell>
                        <TableCell>
                          <Chip label={job.status} size="small" sx={{ fontWeight: 600, fontSize: '0.6875rem', ...getStatusSx(job.status) }} />
                        </TableCell>
                        <TableCell>{job.job_template_name || '-'}</TableCell>
                        <TableCell>{job.inventory_name || '-'}</TableCell>
                        <TableCell>{formatDate(job.started)}</TableCell>
                        <TableCell>{formatElapsed(job.elapsed)}</TableCell>
                        <TableCell onClick={(e) => e.stopPropagation()}>
                          <Tooltip title="Voir les détails">
                            <IconButton size="small" onClick={() => handleViewJob(job)}>
                              <ViewIcon />
                            </IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Alert severity="info">Aucun job en cours. Les jobs terminés sont dans l&apos;onglet Historique.</Alert>
            )}
          </TabPanel>

          {/* Onglet Historique */}
          <TabPanel value={tabValue} index={1}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <ModuleSubtitle>Historique des jobs</ModuleSubtitle>
              <IconButton onClick={() => refetchJobHistory()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {jobHistoryLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : sortedHistory.length > 0 ? (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>Nom</TableCell>
                      <TableCell>Statut</TableCell>
                      <TableCell>Template</TableCell>
                      <TableCell>Inventaire</TableCell>
                      <TableCell>Démarré</TableCell>
                      <TableCell>Terminé</TableCell>
                      <TableCell>Durée</TableCell>
                      <TableCell>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {sortedHistory.map((job) => (
                      <TableRow
                        key={job.id}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() => handleViewJob(job)}
                      >
                        <TableCell>{job.id}</TableCell>
                        <TableCell>{job.name}</TableCell>
                        <TableCell>
                          <Chip label={job.status} size="small" sx={{ fontWeight: 600, fontSize: '0.6875rem', ...getStatusSx(job.status) }} />
                        </TableCell>
                        <TableCell>{job.job_template_name || '-'}</TableCell>
                        <TableCell>{job.inventory_name || '-'}</TableCell>
                        <TableCell>{formatDate(job.started)}</TableCell>
                        <TableCell>{formatDate(job.finished)}</TableCell>
                        <TableCell>{formatElapsed(job.elapsed)}</TableCell>
                        <TableCell onClick={(e) => e.stopPropagation()}>
                          <Tooltip title="Voir les détails">
                            <IconButton size="small" onClick={() => handleViewJob(job)}>
                              <ViewIcon />
                            </IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Alert severity="info">Aucun historique disponible</Alert>
            )}
          </TabPanel>

          {/* Onglet Inventaires */}
          <TabPanel value={tabValue} index={2}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <ModuleSubtitle>Inventaires</ModuleSubtitle>
              <IconButton onClick={() => refetchInventories()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {inventoriesLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : inventoriesData && inventoriesData.items.length > 0 ? (
              <Grid container spacing={2}>
                {inventoriesData.items.map((inventory) => (
                  <Grid item xs={12} md={6} lg={4} key={inventory.id}>
                    <Card
                      sx={{ cursor: 'pointer', '&:hover': { boxShadow: 4 } }}
                      onClick={() => handleViewInventory(inventory)}
                    >
                      <CardContent>
                        <Typography variant="h6">{inventory.name}</Typography>
                        {inventory.description && (
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                            {inventory.description}
                          </Typography>
                        )}
                        <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                          {inventory.host_count !== undefined && (
                            <Chip label={`${inventory.host_count} hosts`} size="small" />
                          )}
                          {inventory.organization_name && (
                            <Chip label={inventory.organization_name} size="small" variant="outlined" />
                          )}
                          {inventory.kind && <Chip label={inventory.kind} size="small" variant="outlined" />}
                        </Box>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            ) : (
              <Alert severity="info">Aucun inventaire trouvé dans ce projet Semaphore.</Alert>
            )}
          </TabPanel>

          {/* Onglet Templates */}
          <TabPanel value={tabValue} index={3}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <ModuleSubtitle>Templates de jobs</ModuleSubtitle>
              <IconButton onClick={() => refetchTemplates()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {templatesLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : templatesData && templatesData.items.length > 0 ? (
              <Grid container spacing={2}>
                {templatesData.items.map((template) => (
                  <Grid item xs={12} md={6} lg={4} key={template.id}>
                    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flex: 1 }}>
                        <Typography variant="h6">{template.name}</Typography>
                        {template.description && (
                          <ModuleSecondaryText sx={{ mt: 1, display: 'block' }}>
                            {template.description}
                          </ModuleSecondaryText>
                        )}
                        <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                          {template.job_type && <Chip label={template.job_type} size="small" />}
                          {template.playbook && <Chip label={template.playbook} size="small" variant="outlined" />}
                          {(template.inventory_name || template.summary_fields?.inventory?.name) && (
                            <Chip
                              label={template.inventory_name || template.summary_fields?.inventory?.name}
                              size="small"
                              variant="outlined"
                              color="primary"
                            />
                          )}
                        </Box>
                        <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                          <Button
                            variant="outlined"
                            startIcon={<ViewIcon />}
                            onClick={() => handleViewTemplate(template)}
                            size="small"
                          >
                            Détails
                          </Button>
                          <Button
                            variant="contained"
                            startIcon={<PlayIcon />}
                            onClick={() => handleLaunchTemplate(template.id)}
                            size="small"
                          >
                            Lancer
                          </Button>
                        </Box>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            ) : (
              <Alert severity="info">Aucun template trouvé dans ce projet Semaphore.</Alert>
            )}
          </TabPanel>
        </ModuleCard>

      {/* Dialog de détails du job */}
      <Dialog
        open={jobDetailDialogOpen}
        onClose={() => { setJobDetailDialogOpen(false); setJobStdout(null) }}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Détails du job #{selectedJob?.id}</DialogTitle>
        <DialogContent>
          {selectedJob && (
            <Box sx={{ mt: 2 }}>
              <ModuleSubtitle sx={{ mb: 1 }}>Informations</ModuleSubtitle>
              <Grid container spacing={2} sx={{ mb: 2 }}>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Nom</ModuleSecondaryText>
                  <ModuleBodyText>{selectedJob.name}</ModuleBodyText>
                </Grid>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Statut</ModuleSecondaryText>
                  <Chip label={selectedJob.status} size="small" sx={{ fontWeight: 600, fontSize: '0.6875rem', ...getStatusSx(selectedJob.status) }} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Template</ModuleSecondaryText>
                  <ModuleBodyText>{selectedJob.job_template_name || '-'}</ModuleBodyText>
                </Grid>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Inventaire</ModuleSecondaryText>
                  <ModuleBodyText>{selectedJob.inventory_name || '-'}</ModuleBodyText>
                </Grid>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Démarré</ModuleSecondaryText>
                  <ModuleBodyText>{formatDate(selectedJob.started)}</ModuleBodyText>
                </Grid>
                <Grid item xs={12} md={6}>
                  <ModuleSecondaryText>Durée</ModuleSecondaryText>
                  <ModuleBodyText>{formatElapsed((selectedJob as any).elapsed)}</ModuleBodyText>
                </Grid>
              </Grid>
              <ModuleSubtitle sx={{ mb: 1 }}>Sortie standard</ModuleSubtitle>
              {jobStdout !== null ? (
                <Box sx={{ bgcolor: 'grey.900', p: 2, borderRadius: 1, maxHeight: 400, overflow: 'auto' }}>
                  <CodeBlock language="text" showLineNumbers={false}>{jobStdout}</CodeBlock>
                </Box>
              ) : (
                <Button
                  variant="outlined"
                  onClick={handleLoadJobStdout}
                  disabled={loadingStdout}
                  startIcon={<ViewIcon />}
                  size="small"
                >
                  {loadingStdout ? 'Chargement...' : 'Voir la sortie'}
                </Button>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setJobDetailDialogOpen(false); setJobStdout(null) }}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Dialog détails inventaire */}
      <Dialog
        open={inventoryDetailDialogOpen}
        onClose={() => setInventoryDetailDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Inventaire {selectedInventory?.name}</DialogTitle>
        <DialogContent>
          {inventoryDetail && (
            <Box sx={{ mt: 2 }}>
              {inventoryDetail.description && (
                <>
                  <Typography variant="subtitle2">Description</Typography>
                  <Typography variant="body1" sx={{ mb: 2 }}>
                    {inventoryDetail.description}
                  </Typography>
                </>
              )}
              <Typography variant="subtitle2">Organisation</Typography>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {inventoryDetail.organization_name || '-'}
              </Typography>
              <Typography variant="subtitle2">Hôtes ({inventoryHosts.length})</Typography>
              {inventoryHosts.length > 0 ? (
                <TableContainer component={Paper} sx={{ mt: 1 }}>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Nom</TableCell>
                        <TableCell>Activé</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {inventoryHosts.map((host) => (
                        <TableRow key={host.id}>
                          <TableCell>{host.name}</TableCell>
                          <TableCell>{host.enabled ? 'Oui' : 'Non'}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                  Aucun hôte
                </Typography>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setInventoryDetailDialogOpen(false)}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Dialog détails template */}
      <Dialog
        open={templateDetailDialogOpen}
        onClose={() => { setTemplateDetailDialogOpen(false); setTemplateDetail(null) }}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Template {selectedTemplate?.name}</DialogTitle>
        <DialogContent>
          {templateDetail && (
            <Box sx={{ mt: 2 }}>
              {templateDetail.description && (
                <>
                  <ModuleSubtitle sx={{ mb: 1 }}>Description</ModuleSubtitle>
                  <ModuleBodyText sx={{ mb: 2 }}>{templateDetail.description}</ModuleBodyText>
                </>
              )}
              <ModuleSubtitle sx={{ mb: 1 }}>Configuration</ModuleSubtitle>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                <Box>
                  <ModuleSecondaryText>Type</ModuleSecondaryText>
                  <ModuleBodyText>{templateDetail.job_type || 'run'}</ModuleBodyText>
                </Box>
                <Box>
                  <ModuleSecondaryText>Playbook</ModuleSecondaryText>
                  <ModuleBodyText>{templateDetail.playbook || '-'}</ModuleBodyText>
                </Box>
                <Box>
                  <ModuleSecondaryText>Projet</ModuleSecondaryText>
                  <ModuleBodyText>{templateDetail.project_name || '-'}</ModuleBodyText>
                </Box>
                <Box>
                  <ModuleSecondaryText>Inventaire</ModuleSecondaryText>
                  <ModuleBodyText>{templateDetail.inventory_name || '-'}</ModuleBodyText>
                </Box>
                {templateDetail.limit && (
                  <Box>
                    <ModuleSecondaryText>Limit</ModuleSecondaryText>
                    <ModuleBodyText>{templateDetail.limit}</ModuleBodyText>
                  </Box>
                )}
              </Box>
              <Box sx={{ mt: 3 }}>
                <Button
                  variant="contained"
                  startIcon={<PlayIcon />}
                  onClick={() => {
                    if (selectedTemplate) handleLaunchTemplate(selectedTemplate.id)
                    setTemplateDetailDialogOpen(false)
                  }}
                  fullWidth
                >
                  Lancer ce template
                </Button>
              </Box>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setTemplateDetailDialogOpen(false); setTemplateDetail(null) }}>Fermer</Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar pour les notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        message={snackbar.message}
      />
    </Box>
  )
}
