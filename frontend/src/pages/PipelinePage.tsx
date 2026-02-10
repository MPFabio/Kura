import { useQuery } from '@tanstack/react-query'
import {
  Box,
  Alert,
  Typography,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  CircularProgress,
  IconButton,
  Tooltip,
  Link,
} from '@mui/material'
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Schedule as ScheduleIcon,
  Cached as CachedIcon,
  OpenInNew as OpenInNewIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { pipelineService, type PipelineRun, type PipelineRunStatus } from '../services/pipelineService'
import { jellyfishColors } from '../theme'

const providerLabels: Record<string, string> = {
  github: 'GitHub Actions',
  gitlab: 'GitLab CI',
  jenkins: 'Jenkins',
}

const providerColors: Record<string, string> = {
  github: jellyfishColors.cyanSoft,
  gitlab: jellyfishColors.violetMedium,
  jenkins: jellyfishColors.magenta,
}

function StatusChip({ status }: { status: PipelineRunStatus }) {
  const config: Record<
    PipelineRunStatus,
    { label: string; color: string; icon: React.ReactNode }
  > = {
    success: {
      label: 'Succès',
      color: jellyfishColors.successSoft,
      icon: <CheckCircleIcon sx={{ fontSize: 16 }} />,
    },
    failure: {
      label: 'Échec',
      color: jellyfishColors.errorSoft,
      icon: <ErrorIcon sx={{ fontSize: 16 }} />,
    },
    running: {
      label: 'En cours',
      color: jellyfishColors.infoSoft,
      icon: <CachedIcon sx={{ fontSize: 16 }} />,
    },
    pending: {
      label: 'En attente',
      color: jellyfishColors.warningSoft,
      icon: <ScheduleIcon sx={{ fontSize: 16 }} />,
    },
    cancelled: {
      label: 'Annulé',
      color: '#808080',
      icon: <ErrorIcon sx={{ fontSize: 16 }} />,
    },
    skipped: {
      label: 'Ignoré',
      color: '#808080',
      icon: <ScheduleIcon sx={{ fontSize: 16 }} />,
    },
  }
  const { label, color } = config[status] ?? config.pending
  return (
    <Chip
      size="small"
      label={label}
      sx={{
        fontWeight: 600,
        fontSize: '0.6875rem',
        backgroundColor: `${color}22`,
        border: `1px solid ${color}`,
        color,
      }}
    />
  )
}

function formatDuration(ms?: number) {
  if (ms == null) return '—'
  if (ms < 1000) return `${ms} ms`
  const sec = Math.round(ms / 1000)
  if (sec < 60) return `${sec} s`
  const min = Math.floor(sec / 60)
  const s = sec % 60
  return s ? `${min} min ${s} s` : `${min} min`
}

function formatDate(s?: string) {
  if (!s) return '—'
  const d = new Date(s)
  return d.toLocaleString('fr-FR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function PipelinePage() {
  const {
    data: runsData,
    isLoading: runsLoading,
    error: runsError,
    refetch: refetchRuns,
  } = useQuery({
    queryKey: ['pipeline-runs'],
    queryFn: () => pipelineService.getRuns({ limit: 50 }),
  })

  const { data: providersData } = useQuery({
    queryKey: ['pipeline-providers'],
    queryFn: () => pipelineService.getProviders(),
  })

  const runs = runsData?.runs ?? []
  const providers = providersData?.providers ?? []
  const hasRuns = runs.length > 0

  return (
    <Box>
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          mb: 4,
          pb: 3,
          borderBottom: '2px solid rgba(0, 229, 255, 0.15)',
        }}
      >
        <ModuleTitle>Pipelines CI/CD</ModuleTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Tooltip title="Actualiser">
            <IconButton
              onClick={() => refetchRuns()}
              sx={{
                color: jellyfishColors.cyanSoft,
                '&:hover': { backgroundColor: 'rgba(0, 229, 255, 0.1)' },
              }}
            >
              <CachedIcon />
            </IconButton>
          </Tooltip>
          {hasRuns && (
            <Typography
              sx={{
                fontFamily: '"JetBrains Mono", monospace',
                color: jellyfishColors.cyanSoft,
                fontSize: '0.875rem',
              }}
            >
              {runs.length} exécution{runs.length > 1 ? 's' : ''}
            </Typography>
          )}
        </Box>
      </Box>

      <Box sx={{ mb: 4 }}>
        <ModuleCard>
          <Alert
            severity="info"
            sx={{
              borderRadius: 0,
              borderLeft: `4px solid ${jellyfishColors.cyanSoft}`,
              backgroundColor: 'rgba(0, 229, 255, 0.05)',
            }}
          >
            <Typography variant="body1" sx={{ mb: 1 }}>
              Agrégation des pipelines GitHub Actions, GitLab CI et Jenkins.
            </Typography>
            <Typography variant="body2" sx={{ color: '#a0a0a0' }}>
              Configurez des webhooks vers{' '}
              <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>
                /api/v1/pipeline/webhooks/github
              </code>
              ,{' '}
              <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>
                /api/v1/pipeline/webhooks/gitlab
              </code>
              ou{' '}
              <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>
                /api/v1/pipeline/webhooks/jenkins
              </code>
              pour recevoir les événements en temps réel.
            </Typography>
          </Alert>
        </ModuleCard>
      </Box>

      {providers.length > 0 && (
        <Box sx={{ mb: 4 }}>
          <Typography
            sx={{
              textTransform: 'uppercase',
              letterSpacing: '0.15em',
              mb: 2,
              fontWeight: 700,
              color: '#808080',
              fontSize: '0.75rem',
            }}
          >
            Providers supportés
          </Typography>
          <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
            {providers.map((p) => (
              <Chip
                key={p.id}
                label={p.name}
                sx={{
                  border: `1px solid ${providerColors[p.id] ?? jellyfishColors.cyanSoft}`,
                  color: providerColors[p.id] ?? jellyfishColors.cyanSoft,
                  backgroundColor: 'transparent',
                  fontWeight: 600,
                }}
              />
            ))}
          </Box>
        </Box>
      )}

      {runsLoading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
          <CircularProgress sx={{ color: jellyfishColors.cyanSoft }} />
        </Box>
      )}

      {runsError && (
        <ModuleCard>
          <Alert severity="error" sx={{ mb: 2 }}>
            Impossible de charger les exécutions.
          </Alert>
          <Typography variant="body2" sx={{ color: '#b8b8b8', fontFamily: 'monospace', fontSize: '0.75rem', mb: 2 }}>
            {runsError instanceof Error ? runsError.message : String(runsError)}
          </Typography>
          <Typography variant="body2" sx={{ color: '#808080', fontSize: '0.8125rem' }}>
            Vérifications : lancez <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>./scripts/check-services.sh</code> ou testez :<br />
            Kong : <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>curl http://localhost:8000/api/v1/pipeline/runs</code><br />
            Direct : <code style={{ background: '#2c2f3f', padding: '2px 6px' }}>curl http://localhost:8084/health</code>
          </Typography>
        </ModuleCard>
      )}

      {!runsLoading && !runsError && hasRuns && (
        <ModuleCard sx={{ overflow: 'hidden' }}>
          <Typography
            sx={{
              mb: 2,
              fontWeight: 700,
              fontSize: '1rem',
              color: '#f0f0f0',
            }}
          >
            Historique des exécutions
          </Typography>
          <TableContainer component={Paper} sx={{ boxShadow: 'none', background: 'transparent' }}>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell sx={{ fontWeight: 700 }}>Provider</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Repository</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Branche</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Workflow</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Statut</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Durée</TableCell>
                  <TableCell sx={{ fontWeight: 700 }}>Date</TableCell>
                  <TableCell sx={{ fontWeight: 700 }} align="right">
                    Lien
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {runs.map((run: PipelineRun) => (
                  <TableRow key={run.id} hover>
                    <TableCell>
                      <Chip
                        size="small"
                        label={providerLabels[run.provider] ?? run.provider}
                        sx={{
                          border: `1px solid ${providerColors[run.provider] ?? jellyfishColors.cyanSoft}`,
                          color: providerColors[run.provider] ?? jellyfishColors.cyanSoft,
                          backgroundColor: 'transparent',
                          fontWeight: 600,
                          fontSize: '0.6875rem',
                        }}
                      />
                    </TableCell>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}>
                      {run.repository || '—'}
                    </TableCell>
                    <TableCell>{run.branch || '—'}</TableCell>
                    <TableCell>{run.workflow_name || '—'}</TableCell>
                    <TableCell>
                      <StatusChip status={run.status} />
                    </TableCell>
                    <TableCell>{formatDuration(run.duration_ms)}</TableCell>
                    <TableCell sx={{ fontSize: '0.8125rem', color: '#b8b8b8' }}>
                      {formatDate(run.finished_at ?? run.started_at ?? run.created_at)}
                    </TableCell>
                    <TableCell align="right">
                      {run.external_url && (
                        <Tooltip title="Ouvrir dans le provider">
                          <IconButton
                            component={Link}
                            href={run.external_url}
                            target="_blank"
                            rel="noopener"
                            size="small"
                            sx={{
                              color: jellyfishColors.cyanSoft,
                              '&:hover': { backgroundColor: 'rgba(0, 229, 255, 0.1)' },
                            }}
                          >
                            <OpenInNewIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </ModuleCard>
      )}

      {!runsLoading && !runsError && !hasRuns && (
        <ModuleCard>
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <ScheduleIcon sx={{ fontSize: 48, color: jellyfishColors.cyanSoft, opacity: 0.6 }} />
            <Typography sx={{ mt: 2, color: '#a0a0a0' }}>
              Aucune exécution enregistrée. Les runs apparaîtront ici une fois les webhooks
              configurés et les pipelines déclenchés.
            </Typography>
          </Box>
        </ModuleCard>
      )}
    </Box>
  )
}
