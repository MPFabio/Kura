import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link as RouterLink } from 'react-router-dom'
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
  TextField,
  Button,
  Collapse,
} from '@mui/material'
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Schedule as ScheduleIcon,
  Cached as CachedIcon,
  OpenInNew as OpenInNewIcon,
  Sync as SyncIcon,
  Link as LinkIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  ContentCopy as ContentCopyIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import {
  pipelineService,
  type PipelineRun,
  type PipelineRunStatus,
  type PipelineConfig,
} from '../services/pipelineService'
import { projectService } from '../services/projectService'
import { useProject } from '../contexts/ProjectContext'
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
  const queryClient = useQueryClient()
  const { currentProject } = useProject()
  const [configExpanded, setConfigExpanded] = useState(false)
  const [tokenInput, setTokenInput] = useState('')
  const [reposInput, setReposInput] = useState('')

  const {
    data: runsData,
    isLoading: runsLoading,
    error: runsError,
    refetch: refetchRuns,
  } = useQuery({
    queryKey: ['pipeline-runs'],
    queryFn: () => pipelineService.getRuns({ limit: 50 }),
  })

  const { data: configData, refetch: refetchConfig } = useQuery({
    queryKey: ['pipeline-config'],
    queryFn: () => pipelineService.getConfig(),
  })

  const saveConfigMutation = useMutation({
    mutationFn: async (data: { github_token?: string; github_repos?: string[] }) => {
      const result = await pipelineService.setConfig(data)
      if (currentProject && data.github_repos?.length) {
        for (const repo of data.github_repos) {
          try {
            await projectService.createProjectMapping(currentProject.id, { github_repository: repo })
          } catch {
            // Ignore si mapping existe déjà (contrainte unique)
          }
        }
      }
      return result
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline-config'] })
      queryClient.invalidateQueries({ queryKey: ['project-mappings', currentProject?.id] })
      refetchConfig()
      setTokenInput('')
    },
  })

  const handleSaveConfig = () => {
    const repos = reposInput
      .split(/[,;\n]/)
      .map((r) => r.trim())
      .filter(Boolean)
    saveConfigMutation.mutate({
      ...(tokenInput && { github_token: tokenInput }),
      github_repos: repos,
    })
  }

  const syncMutation = useMutation({
    mutationFn: () => pipelineService.sync(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['pipeline-runs'] })
      if (data.runs > 0) {
        refetchRuns()
      }
    },
  })

  const { data: providersData } = useQuery({
    queryKey: ['pipeline-providers'],
    queryFn: () => pipelineService.getProviders(),
  })

  const runsRaw = runsData?.runs ?? []
  const runs = [...runsRaw].sort((a, b) => {
    const dateB = new Date(b.finished_at ?? b.started_at ?? b.created_at).getTime()
    const dateA = new Date(a.finished_at ?? a.started_at ?? a.created_at).getTime()
    return dateB - dateA
  })
  const providers = providersData?.providers ?? []
  const hasRuns = runs.length > 0
  const config = configData as PipelineConfig | undefined
  const isLinked = config?.linked ?? false

  const webhookBaseUrl =
    (typeof import.meta !== 'undefined' && import.meta.env?.VITE_PUBLIC_URL
      ? String(import.meta.env.VITE_PUBLIC_URL).replace(/\/$/, '')
      : null) ??
    (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_BASE_URL != null
      ? String(import.meta.env.VITE_API_BASE_URL).replace(/\/$/, '')
      : '') ??
    (typeof window !== 'undefined' ? window.location.origin : '')
  const webhookUrl = webhookBaseUrl ? `${webhookBaseUrl}/api/v1/pipeline/webhooks/github` : ''

  const handleCopyWebhookUrl = () => {
    if (webhookUrl) {
      navigator.clipboard.writeText(webhookUrl)
      // Feedback visuel optionnel via snackbar si disponible
    }
  }

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
          <Tooltip title="Synchroniser GitHub">
            <span>
              <IconButton
                onClick={() => syncMutation.mutate()}
                disabled={syncMutation.isPending}
                sx={{
                  color: jellyfishColors.cyanSoft,
                  '&:hover': { backgroundColor: 'rgba(0, 229, 255, 0.1)' },
                }}
              >
                <SyncIcon />
              </IconButton>
            </span>
          </Tooltip>
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
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              cursor: 'pointer',
              py: 0.5,
            }}
            onClick={() => setConfigExpanded(!configExpanded)}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
              <LinkIcon sx={{ color: jellyfishColors.cyanSoft }} />
              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                Connecter un dépôt GitHub
              </Typography>
              {isLinked && (
                <Chip
                  size="small"
                  label="Connecté"
                  sx={{
                    border: `1px solid ${jellyfishColors.successSoft}`,
                    color: jellyfishColors.successSoft,
                    backgroundColor: 'transparent',
                    fontWeight: 600,
                  }}
                />
              )}
            </Box>
            {configExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </Box>
          <Collapse in={configExpanded}>
            <Box sx={{ mt: 2, pt: 2, borderTop: '1px solid rgba(0,229,255,0.2)' }}>
              <Typography variant="body2" sx={{ color: '#a0a0a0', mb: 2 }}>
                Liez vos dépôts GitHub pour afficher les exécutions GitHub Actions. Créez un{' '}
                <Link
                  href="https://github.com/settings/tokens"
                  target="_blank"
                  rel="noopener"
                  sx={{ color: jellyfishColors.cyanSoft }}
                >
                  Personal Access Token
                </Link>{' '}
                (scope <code style={{ background: '#2c2f3f', padding: '1px 4px' }}>repo</code> ou Actions read).
              </Typography>
              <TextField
                fullWidth
                size="small"
                label="Token GitHub"
                type="password"
                placeholder={isLinked ? '•••••••• (laisser vide pour conserver)' : 'ghp_xxx...'}
                value={tokenInput}
                onChange={(e) => setTokenInput(e.target.value)}
                sx={{ mb: 2, '& .MuiOutlinedInput-root': { backgroundColor: '#1a1d24' } }}
              />
              <TextField
                fullWidth
                size="small"
                label="Dépôts (owner/repo)"
                placeholder={
                  config?.github_repos?.length
                    ? config.github_repos.join(', ')
                    : 'owner/repo ou owner/repo1, owner/repo2'
                }
                value={reposInput}
                onChange={(e) => setReposInput(e.target.value)}
                helperText="Format: owner/repo. Plusieurs dépôts séparés par des virgules."
                sx={{ mb: 2, '& .MuiOutlinedInput-root': { backgroundColor: '#1a1d24' } }}
              />
              <Box sx={{ display: 'flex', gap: 2 }}>
                <Button
                  variant="contained"
                  onClick={handleSaveConfig}
                  disabled={saveConfigMutation.isPending}
                  sx={{
                    backgroundColor: jellyfishColors.cyanSoft,
                    color: '#0a0d12',
                    '&:hover': { backgroundColor: 'rgba(0, 229, 255, 0.9)' },
                  }}
                >
                  {saveConfigMutation.isPending ? 'Enregistrement...' : 'Enregistrer'}
                </Button>
                {saveConfigMutation.isSuccess && (
                  <Typography sx={{ color: jellyfishColors.successSoft, alignSelf: 'center' }}>
                    ✓ Configuration enregistrée
                  </Typography>
                )}
              </Box>

              <Box sx={{ mt: 3, pt: 2, borderTop: '1px solid rgba(0,229,255,0.15)' }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1.5, color: jellyfishColors.cyanSoft }}>
                  Option temps réel (webhooks)
                </Typography>
                <Typography variant="body2" sx={{ color: '#a0a0a0', mb: 1.5 }}>
                  Pour une mise à jour immédiate sans polling, configurez un webhook GitHub en pointant vers l&apos;URL
                  ci-dessous. Voir la{' '}
                  <Link component={RouterLink} to="/documentation?section=pipelines" sx={{ color: jellyfishColors.cyanSoft }}>
                    documentation
                  </Link>
                  .
                </Typography>
                {webhookUrl ? (
                  <TextField
                    fullWidth
                    size="small"
                    label="URL du webhook GitHub"
                    value={webhookUrl}
                    readOnly
                    InputProps={{
                      readOnly: true,
                      endAdornment: (
                        <Tooltip title="Copier">
                          <IconButton
                            size="small"
                            onClick={handleCopyWebhookUrl}
                            sx={{ color: jellyfishColors.cyanSoft }}
                          >
                            <ContentCopyIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      ),
                    }}
                    sx={{ '& .MuiOutlinedInput-root': { backgroundColor: '#1a1d24' } }}
                  />
                ) : (
                  <Typography variant="body2" sx={{ color: '#808080', fontFamily: 'monospace' }}>
                    Définissez VITE_PUBLIC_URL ou VITE_API_BASE_URL pour afficher l&apos;URL.
                  </Typography>
                )}
              </Box>
            </Box>
          </Collapse>
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
              Aucune exécution enregistrée. Connectez un dépôt GitHub ci-dessus, puis cliquez sur
              Sync pour afficher les runs.
            </Typography>
          </Box>
        </ModuleCard>
      )}
    </Box>
  )
}
