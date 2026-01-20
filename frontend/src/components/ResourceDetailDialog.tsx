import { useState } from 'react'
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

interface ResourceDetailDialogProps {
  open: boolean
  onClose: () => void
  resourceType: 'pod' | 'deployment' | 'service'
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
        default:
          return Promise.resolve('')
      }
    },
    enabled: open && activeTab === 1,
  })

  const { data: logs, isLoading: logsLoading, error: logsError, refetch: refetchLogs } = useQuery({
    queryKey: ['pod-logs', namespace, name],
    queryFn: () => k8sService.getPodLogs(namespace, name, undefined, 100),
    enabled: open && activeTab === 2 && resourceType === 'pod',
  })

  const eventsTabIndex = resourceType === 'pod' ? 4 : 2
  const terminalTabIndex = resourceType === 'pod' ? 3 : -1
  const { data: events, isLoading: eventsLoading, error: eventsError, refetch: refetchEvents } = useQuery({
    queryKey: ['events', namespace],
    queryFn: () => k8sService.getEvents(namespace),
    enabled: open && activeTab === eventsTabIndex,
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
    switch (type.toLowerCase()) {
      case 'warning':
        return 'warning'
      case 'error':
      case 'failed':
        return 'error'
      default:
        return 'info'
    }
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
            <Tab label="Events" />
          </Tabs>
        </Box>

        {activeTab === 0 && (
          <Box>
            <Typography variant="body1" sx={{ mb: 2 }}>
              <strong>Namespace:</strong> {namespace}
            </Typography>
            <Typography variant="body1" sx={{ mb: 2 }}>
              <strong>Name:</strong> {name}
            </Typography>
            <Typography variant="body1" sx={{ mb: 2 }}>
              <strong>Type:</strong> {resourceType}
            </Typography>
            <Alert severity="info">
              Les détails complets seront disponibles dans une prochaine version.
            </Alert>
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
            {logsLoading ? (
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
            <Terminal
              namespace={namespace}
              pod={name}
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
                      .filter((e) => e.involvedObjectKind === resourceType.toUpperCase() && e.involvedObject.includes(name))
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
