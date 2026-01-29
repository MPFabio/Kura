import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
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
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ansibleService, AnsibleJobSummary, AnsibleInventorySummary, AnsibleJobTemplateSummary } from '../services/ansibleService'

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
  const [tabValue, setTabValue] = useState(0)
  const [selectedJob, setSelectedJob] = useState<AnsibleJobSummary | null>(null)
  const [jobDetailDialogOpen, setJobDetailDialogOpen] = useState(false)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  // Requêtes pour les données
  const {
    data: jobsData,
    isLoading: jobsLoading,
    refetch: refetchJobs,
  } = useQuery({
    queryKey: ['ansible-jobs'],
    queryFn: () => ansibleService.getJobs(),
    refetchInterval: 30000, // Rafraîchir toutes les 30 secondes
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

  const handleViewJob = async (job: AnsibleJobSummary) => {
    try {
      const detail = await ansibleService.getJob(job.id)
      setSelectedJob({ ...job, ...detail } as AnsibleJobSummary)
      setJobDetailDialogOpen(true)
    } catch (error) {
      setSnackbar({ open: true, message: 'Erreur lors de la récupération des détails du job', severity: 'error' })
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
      case 'successful':
      case 'success':
        return 'success'
      case 'failed':
      case 'error':
        return 'error'
      case 'running':
      case 'pending':
        return 'info'
      case 'canceled':
      case 'cancelled':
        return 'warning'
      default:
        return 'default'
    }
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

  // Vérifier si le service Ansible Tower est configuré
  const isServiceAvailable = jobsData !== undefined || inventoriesData !== undefined

  return (
    <Box>
      <ModuleTitle>Ansible</ModuleTitle>

      {!isServiceAvailable && (
        <ModuleCard>
          <Alert severity="info">
            Le service Ansible sera bientôt disponible. Cette page affichera les jobs Ansible Tower, les inventaires et
            l'historique d'exécution.
          </Alert>
        </ModuleCard>
      )}

      {isServiceAvailable && (
        <ModuleCard>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab label="Jobs" />
              <Tab label="Historique" />
              <Tab label="Inventaires" />
              <Tab label="Templates" />
            </Tabs>
          </Box>

          {/* Onglet Jobs */}
          <TabPanel value={tabValue} index={0}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6">Jobs en cours</Typography>
              <IconButton onClick={() => refetchJobs()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {jobsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : jobsData && jobsData.items.length > 0 ? (
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
                    {jobsData.items.map((job) => (
                      <TableRow key={job.id}>
                        <TableCell>{job.id}</TableCell>
                        <TableCell>{job.name}</TableCell>
                        <TableCell>
                          <Chip label={job.status} color={getStatusColor(job.status) as any} size="small" />
                        </TableCell>
                        <TableCell>{job.job_template_name || '-'}</TableCell>
                        <TableCell>{job.inventory_name || '-'}</TableCell>
                        <TableCell>{formatDate(job.started)}</TableCell>
                        <TableCell>{formatElapsed(job.elapsed)}</TableCell>
                        <TableCell>
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
              <Alert severity="info">Aucun job trouvé</Alert>
            )}
          </TabPanel>

          {/* Onglet Historique */}
          <TabPanel value={tabValue} index={1}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6">Historique des jobs</Typography>
              <IconButton onClick={() => refetchJobHistory()} color="primary">
                <RefreshIcon />
              </IconButton>
            </Box>
            {jobHistoryLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : jobHistoryData && jobHistoryData.items.length > 0 ? (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>ID</TableCell>
                      <TableCell>Nom</TableCell>
                      <TableCell>Statut</TableCell>
                      <TableCell>Template</TableCell>
                      <TableCell>Démarré</TableCell>
                      <TableCell>Terminé</TableCell>
                      <TableCell>Durée</TableCell>
                      <TableCell>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {jobHistoryData.items.map((job) => (
                      <TableRow key={job.id}>
                        <TableCell>{job.id}</TableCell>
                        <TableCell>{job.name}</TableCell>
                        <TableCell>
                          <Chip label={job.status} color={getStatusColor(job.status) as any} size="small" />
                        </TableCell>
                        <TableCell>{job.job_template_name || '-'}</TableCell>
                        <TableCell>{formatDate(job.started)}</TableCell>
                        <TableCell>{formatDate(job.finished)}</TableCell>
                        <TableCell>{formatElapsed(job.elapsed)}</TableCell>
                        <TableCell>
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
              <Typography variant="h6">Inventaires</Typography>
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
                    <Card>
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
              <Alert severity="info">Aucun inventaire trouvé</Alert>
            )}
          </TabPanel>

          {/* Onglet Templates */}
          <TabPanel value={tabValue} index={3}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6">Templates de jobs</Typography>
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
                    <Card>
                      <CardContent>
                        <Typography variant="h6">{template.name}</Typography>
                        {template.description && (
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                            {template.description}
                          </Typography>
                        )}
                        <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                          {template.job_type && <Chip label={template.job_type} size="small" />}
                          {template.playbook && <Chip label={template.playbook} size="small" variant="outlined" />}
                        </Box>
                        <Box sx={{ mt: 2 }}>
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
              <Alert severity="info">Aucun template trouvé</Alert>
            )}
          </TabPanel>
        </ModuleCard>
      )}

      {/* Dialog de détails du job */}
      <Dialog open={jobDetailDialogOpen} onClose={() => setJobDetailDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Détails du job #{selectedJob?.id}</DialogTitle>
        <DialogContent>
          {selectedJob && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2">Nom</Typography>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {selectedJob.name}
              </Typography>
              <Typography variant="subtitle2">Statut</Typography>
              <Chip label={selectedJob.status} color={getStatusColor(selectedJob.status) as any} sx={{ mb: 2 }} />
              <Typography variant="subtitle2">Template</Typography>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {selectedJob.job_template_name || '-'}
              </Typography>
              <Typography variant="subtitle2">Inventaire</Typography>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {selectedJob.inventory_name || '-'}
              </Typography>
              <Typography variant="subtitle2">Démarré</Typography>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {formatDate(selectedJob.started)}
              </Typography>
              {selectedJob.finished && (
                <>
                  <Typography variant="subtitle2">Terminé</Typography>
                  <Typography variant="body1" sx={{ mb: 2 }}>
                    {formatDate(selectedJob.finished)}
                  </Typography>
                </>
              )}
              <Typography variant="subtitle2">Durée</Typography>
              <Typography variant="body1">{formatElapsed(selectedJob.elapsed)}</Typography>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setJobDetailDialogOpen(false)}>Fermer</Button>
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
