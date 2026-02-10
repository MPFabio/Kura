import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Tabs,
  Tab,
  Box,
  Typography,
  Paper,
  CircularProgress,
  Alert,
  IconButton,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material'
import {
  Close as CloseIcon,
  ContentCopy as ContentCopyIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material'
import { k8sService } from '../services/k8sService'
import { format } from 'date-fns'
import fr from 'date-fns/locale/fr'
import Terminal from './Terminal'

export type ResourceDetailType = 'pod' | 'deployment' | 'service' | 'configmap' | 'secret' | 'node'

interface ResourceDetailDialogProps {
  open: boolean
  onClose: () => void
  resourceType: ResourceDetailType
  namespace: string
  name: string
}

export default function ResourceDetailDialog({
  open,
  onClose,
  resourceType,
  namespace,
  name,
}: ResourceDetailDialogProps) {
  const [activeTab, setActiveTab] = useState(0)
  const [copied, setCopied] = useState(false)
  const [selectedContainer, setSelectedContainer] = useState<string>('')

  const { data: podDetail, isLoading: podDetailLoading } = useQuery({
    queryKey: ['pod-detail', namespace, name],
    queryFn: () => k8sService.getPod(namespace, name),
    enabled: open && resourceType === 'pod' && !!namespace && !!name,
  })

  const containerNames: string[] = podDetail?.spec?.containers?.map((c) => c.name) ?? []
  useEffect(() => {
    if (containerNames.length > 0 && !containerNames.includes(selectedContainer)) {
      setSelectedContainer(containerNames[0])
    }
  }, [containerNames.join(','), selectedContainer])

  const { data: deploymentDetail, isLoading: deploymentDetailLoading } = useQuery({
    queryKey: ['deployment-detail', namespace, name],
    queryFn: () => k8sService.getDeployment(namespace, name),
    enabled: open && resourceType === 'deployment' && !!namespace && !!name,
  })

  const { data: yaml, isLoading: yamlLoading, error: yamlError, refetch: refetchYAML } = useQuery({
    queryKey: [`${resourceType}-yaml`, namespace, name],
    queryFn: () => {
      switch (resourceType) {
        case 'pod':
          return k8sService.getPodYAML(namespace, name)
        case 'deployment':
          return k8sService.getDeploymentYAML(namespace, name)
        case 'service':
          return k8sService.getServiceYAML(namespace, name)
        case 'configmap':
          return k8sService.getConfigMapYAML(namespace, name)
        case 'secret':
          return k8sService.getSecretYAML(namespace, name)
        case 'node':
          return k8sService.getNodeYAML(name)
        default:
          return Promise.resolve('')
      }
    },
    enabled: open && activeTab === 1 && (resourceType !== 'node' ? !!namespace : true),
  })

  const { data: logs, isLoading: logsLoading, error: logsError, refetch: refetchLogs } = useQuery({
    queryKey: ['pod-logs', namespace, name, selectedContainer],
    queryFn: () => k8sService.getPodLogs(namespace, name, selectedContainer || undefined, 100),
    enabled: open && activeTab === 2 && resourceType === 'pod' && (containerNames.length <= 1 || !!selectedContainer),
  })

  const hasEventsTab = resourceType !== 'node'
  const eventsTabIndex = resourceType === 'pod' ? 4 : 2
  const terminalTabIndex = resourceType === 'pod' ? 3 : -1
  const { data: events, isLoading: eventsLoading, error: eventsError, refetch: refetchEvents } = useQuery({
    queryKey: ['events', namespace],
    queryFn: () => k8sService.getEvents(namespace),
    enabled: open && hasEventsTab && !!namespace && activeTab === eventsTabIndex,
  })

  const handleCopy = async () => {
    const textToCopy = activeTab === 1 ? yaml : logs
    if (textToCopy) {
      await navigator.clipboard.writeText(textToCopy)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleRefresh = () => {
    if (activeTab === 1) {
      refetchYAML()
    } else if (activeTab === 2 && resourceType === 'pod') {
      refetchLogs()
    } else if (activeTab === eventsTabIndex) {
      refetchEvents()
    }
    // Le terminal se rafraîchit automatiquement via WebSocket
  }

  const getEventTypeColor = (type: string) => {
    switch (type?.toLowerCase()) {
      case 'warning':
        return 'warning'
      case 'error':
      case 'failed':
        return 'error'
      default:
        return 'info'
    }
  }

  const eventKindForResource = (t: ResourceDetailType): string => {
    const map: Record<ResourceDetailType, string> = {
      pod: 'Pod',
      deployment: 'Deployment',
      service: 'Service',
      configmap: 'ConfigMap',
      secret: 'Secret',
      node: 'Node',
    }
    return map[t] ?? t
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">
            {name} ({resourceType})
          </Typography>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>
      <DialogContent>
          <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
          <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
            <Tab label="Overview" />
            <Tab label="YAML" />
            {resourceType === 'pod' && <Tab label="Logs" />}
            {resourceType === 'pod' && <Tab label="Terminal" />}
            {hasEventsTab && <Tab label="Events" />}
          </Tabs>
        </Box>

        {activeTab === 0 && (
          <Box>
            {resourceType === 'deployment' && (
              <>
                {deploymentDetailLoading ? (
                  <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                    <CircularProgress size={24} />
                  </Box>
                ) : deploymentDetail ? (
                  <>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Namespace:</strong> {deploymentDetail.metadata?.namespace ?? namespace}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Name:</strong> {deploymentDetail.metadata?.name ?? name}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Replicas souhaitées:</strong> {deploymentDetail.spec?.replicas ?? '-'}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Replicas prêtes:</strong> {deploymentDetail.status?.readyReplicas ?? 0}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Replicas disponibles:</strong> {deploymentDetail.status?.availableReplicas ?? 0}
                    </Typography>
                    <Typography variant="body1" sx={{ mb: 1 }}>
                      <strong>Replicas à jour:</strong> {deploymentDetail.status?.updatedReplicas ?? 0}
                    </Typography>
                    {deploymentDetail.spec?.selector?.matchLabels && Object.keys(deploymentDetail.spec.selector.matchLabels).length > 0 && (
                      <Typography variant="body2" sx={{ mt: 2 }}>
                        <strong>Selector:</strong>{' '}
                        {Object.entries(deploymentDetail.spec.selector.matchLabels).map(([k, v]) => `${k}=${v}`).join(', ')}
                      </Typography>
                    )}
                  </>
                ) : (
                  <Alert severity="info">Impossible de charger les détails du deployment.</Alert>
                )}
              </>
            )}
            {(resourceType === 'pod' || resourceType === 'service' || resourceType === 'configmap' || resourceType === 'secret' || resourceType === 'node') && (
              <>
                {resourceType !== 'node' && (
                  <Typography variant="body1" sx={{ mb: 1 }}>
                    <strong>Namespace:</strong> {namespace || '-'}
                  </Typography>
                )}
                <Typography variant="body1" sx={{ mb: 1 }}>
                  <strong>Name:</strong> {name}
                </Typography>
                <Typography variant="body1" sx={{ mb: 1 }}>
                  <strong>Type:</strong> {resourceType}
                </Typography>
                {resourceType === 'pod' && (
                  <>
                    {podDetailLoading ? (
                      <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                        <CircularProgress size={24} />
                      </Box>
                    ) : podDetail ? (
                      <>
                        <Typography variant="body1" sx={{ mb: 1 }}>
                          <strong>Phase:</strong> <Chip label={podDetail.status?.phase ?? '-'} size="small" sx={{ verticalAlign: 'middle' }} />
                        </Typography>
                        {podDetail.spec?.containers && podDetail.spec.containers.length > 0 && (
                          <Typography variant="body2" sx={{ mt: 1 }}>
                            <strong>Containers:</strong> {podDetail.spec.containers.map((c) => c.name).join(', ')}
                          </Typography>
                        )}
                      </>
                    ) : null}
                  </>
                )}
              </>
            )}
          </Box>
        )}

        {activeTab === 1 && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1, gap: 1 }}>
              <Tooltip title={copied ? 'Copié!' : 'Copier'}>
                <IconButton onClick={handleCopy} size="small">
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
              <Tooltip title="Rafraîchir">
                <IconButton onClick={handleRefresh} size="small">
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Box>
            {yamlLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : yamlError ? (
              <Alert severity="error">
                Erreur: {(yamlError as any)?.response?.data?.error || (yamlError as any)?.message || 'Erreur inconnue'}
              </Alert>
            ) : (
              <Paper sx={{ p: 2, bgcolor: '#1e1e1e', color: '#d4d4d4', fontFamily: 'monospace', fontSize: '0.875rem', overflow: 'auto', maxHeight: '60vh' }}>
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                  {yaml}
                </pre>
              </Paper>
            )}
          </Box>
        )}

        {activeTab === 2 && resourceType === 'pod' && (
          <Box>
            {containerNames.length > 1 && (
              <FormControl size="small" sx={{ minWidth: 220, mb: 2 }}>
                <InputLabel>Container</InputLabel>
                <Select
                  value={selectedContainer}
                  label="Container"
                  onChange={(e) => setSelectedContainer(e.target.value)}
                >
                  {containerNames.map((cn) => (
                    <MenuItem key={cn} value={cn}>{cn}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1, gap: 1 }}>
              <Tooltip title={copied ? 'Copié!' : 'Copier'}>
                <IconButton onClick={handleCopy} size="small">
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
              <Tooltip title="Rafraîchir">
                <IconButton onClick={handleRefresh} size="small">
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Box>
            {containerNames.length > 1 && !selectedContainer ? (
              <Alert severity="info">Sélectionnez un container pour afficher les logs.</Alert>
            ) : logsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : logsError ? (
              <Alert severity="error">
                Erreur: {(logsError as any)?.response?.data?.error || (logsError as any)?.message || 'Erreur inconnue'}
              </Alert>
            ) : (
              <Paper sx={{ p: 2, bgcolor: '#1e1e1e', color: '#d4d4d4', fontFamily: 'monospace', fontSize: '0.875rem', overflow: 'auto', maxHeight: '60vh' }}>
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                  {logs || 'Aucun log disponible'}
                </pre>
              </Paper>
            )}
          </Box>
        )}

        {activeTab === terminalTabIndex && resourceType === 'pod' && (
          <Box>
            {containerNames.length > 1 && (
              <FormControl size="small" sx={{ minWidth: 220, mb: 2 }}>
                <InputLabel>Container</InputLabel>
                <Select
                  value={selectedContainer}
                  label="Container"
                  onChange={(e) => setSelectedContainer(e.target.value)}
                >
                  {containerNames.map((cn) => (
                    <MenuItem key={cn} value={cn}>{cn}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
            <Terminal
              namespace={namespace}
              pod={name}
              container={selectedContainer || undefined}
              open={open && activeTab === terminalTabIndex}
            />
          </Box>
        )}

        {activeTab === eventsTabIndex && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1 }}>
              <Tooltip title="Rafraîchir">
                <IconButton onClick={handleRefresh} size="small">
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Box>
            {eventsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                <CircularProgress />
              </Box>
            ) : eventsError ? (
              <Alert severity="error">
                Erreur: {(eventsError as any)?.response?.data?.error || (eventsError as any)?.message || 'Erreur inconnue'}
              </Alert>
            ) : !events?.items || events.items.length === 0 ? (
              <Alert severity="info">Aucun événement trouvé</Alert>
            ) : (
              <TableContainer component={Paper}>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>Type</TableCell>
                      <TableCell>Raison</TableCell>
                      <TableCell>Message</TableCell>
                      <TableCell>Objet</TableCell>
                      <TableCell>Dernière occurrence</TableCell>
                      <TableCell>Compte</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {events.items
                      .filter((e) => (e as { involvedObjectKind?: string; involvedObject?: string }).involvedObjectKind === eventKindForResource(resourceType) && (e as { involvedObject?: string }).involvedObject?.includes(name))
                      .map((event) => (
                        <TableRow key={event.name}>
                          <TableCell>
                            <Chip
                              label={event.type}
                              color={getEventTypeColor(event.type) as any}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>{event.reason}</TableCell>
                          <TableCell sx={{ maxWidth: 400, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                            {event.message}
                          </TableCell>
                          <TableCell>{event.involvedObject}</TableCell>
                          <TableCell>
                            {format(new Date(event.lastTimestamp), 'PPpp', { locale: fr })}
                          </TableCell>
                          <TableCell>{event.count}</TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Fermer</Button>
      </DialogActions>
    </Dialog>
  )
}
