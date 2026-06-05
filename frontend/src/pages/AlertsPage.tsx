import { useEffect, useState } from 'react'
import {
  Box,
  Alert,
  AlertTitle,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  CircularProgress,
  Typography,
} from '@mui/material'
import {
  Error as ErrorIcon,
  Warning as WarningIcon,
  CheckCircle as CheckCircleIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle } from '../components/ModuleText'
import { metricsService, type ServiceHealth } from '../services/metricsService'
import { pipelineService, type PipelineRun } from '../services/pipelineService'

interface KuraAlert {
  id: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  message: string
  source: string
  timestamp: Date
}

const colors = {
  turquoise: '#00FFFF',
  red: '#FF4444',
  orange: '#FF8800',
  green: '#00FF88',
  grayLight: '#A0A0A0',
}

function severityColor(s: KuraAlert['severity']) {
  return s === 'critical' ? colors.red : s === 'warning' ? colors.orange : colors.green
}

function severityIcon(s: KuraAlert['severity']) {
  if (s === 'critical') return <ErrorIcon sx={{ color: colors.red }} />
  if (s === 'warning') return <WarningIcon sx={{ color: colors.orange }} />
  return <CheckCircleIcon sx={{ color: colors.green }} />
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<KuraAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const buildAlerts = async () => {
    const result: KuraAlert[] = []

    // ── 1. Services hors ligne (metrics-service) ──────────────────────────────
    try {
      const health: ServiceHealth[] = await metricsService.getHealth()
      const down = health.filter((s) => !s.up)
      const up = health.filter((s) => s.up)

      down.forEach((svc) => {
        result.push({
          id: `down-${svc.job}`,
          severity: 'critical',
          title: `Service hors ligne : ${svc.name}`,
          message: `Le service ${svc.job} ne répond plus à son health check. Vérifiez les logs du conteneur.`,
          source: 'Monitoring',
          timestamp: new Date(),
        })
      })

      if (down.length === 0 && up.length > 0) {
        result.push({
          id: 'all-up',
          severity: 'info',
          title: 'Tous les services sont opérationnels',
          message: `${up.length} service${up.length > 1 ? 's' : ''} actif${up.length > 1 ? 's' : ''}`,
          source: 'Monitoring',
          timestamp: new Date(),
        })
      }
    } catch {
      result.push({
        id: 'metrics-error',
        severity: 'warning',
        title: 'Monitoring non disponible',
        message: 'Le metrics-service ne répond pas. Impossible de vérifier l\'état des services.',
        source: 'Monitoring',
        timestamp: new Date(),
      })
    }

    // ── 2. Échecs de pipelines récents ────────────────────────────────────────
    try {
      const runs = await pipelineService.getRuns({ limit: 20 })
      const recentRuns: PipelineRun[] = runs?.runs ?? []

      const failed = recentRuns.filter(
        (r) => (r.status as string) === 'failure' || (r.status as string) === 'failed'
      )

      const seen = new Set<string>()
      failed.forEach((r) => {
        const key = `${r.repository}-${r.workflow_name}`
        if (seen.has(key)) return
        seen.add(key)
        result.push({
          id: `pipeline-${r.id}`,
          severity: 'warning',
          title: `Pipeline en échec : ${r.workflow_name}`,
          message: `Dépôt : ${r.repository} · Branche : ${r.branch}`,
          source: 'Pipelines',
          timestamp: r.started_at ? new Date(r.started_at) : new Date(),
        })
      })
    } catch {
      // Pipeline service indisponible — pas bloquant
    }

    // Tri : critiques en premier, puis warnings, puis info
    const order = { critical: 0, warning: 1, info: 2 }
    result.sort((a, b) => order[a.severity] - order[b.severity])

    setAlerts(result)
    setLastRefresh(new Date())
    setLoading(false)
  }

  useEffect(() => {
    buildAlerts()
    const interval = setInterval(buildAlerts, 30_000)
    return () => clearInterval(interval)
  }, [])

  const critical = alerts.filter((a) => a.severity === 'critical')
  const warnings = alerts.filter((a) => a.severity === 'warning')
  const infos = alerts.filter((a) => a.severity === 'info')

  return (
    <Box>
      <ModuleTitle>Alertes</ModuleTitle>

      {/* Résumé */}
      <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
        <Chip
          icon={<ErrorIcon sx={{ color: `${colors.red} !important` }} />}
          label={`${critical.length} critique${critical.length !== 1 ? 's' : ''}`}
          sx={{ color: colors.red, borderColor: colors.red, fontFamily: '"JetBrains Mono", monospace' }}
          variant="outlined"
        />
        <Chip
          icon={<WarningIcon sx={{ color: `${colors.orange} !important` }} />}
          label={`${warnings.length} warning${warnings.length !== 1 ? 's' : ''}`}
          sx={{ color: colors.orange, borderColor: colors.orange, fontFamily: '"JetBrains Mono", monospace' }}
          variant="outlined"
        />
        <Chip
          icon={<CheckCircleIcon sx={{ color: `${colors.green} !important` }} />}
          label={`${infos.length} info${infos.length !== 1 ? 's' : ''}`}
          sx={{ color: colors.green, borderColor: colors.green, fontFamily: '"JetBrains Mono", monospace' }}
          variant="outlined"
        />
        <Typography sx={{ color: colors.grayLight, fontSize: '0.75rem', ml: 'auto', alignSelf: 'center', fontFamily: '"JetBrains Mono", monospace' }}>
          Mis à jour : {lastRefresh.toLocaleTimeString('fr-FR')} · Rafraîchissement auto toutes les 30s
        </Typography>
      </Box>

      {loading ? (
        <Box display="flex" justifyContent="center" py={4}>
          <CircularProgress sx={{ color: colors.turquoise }} />
        </Box>
      ) : (
        <ModuleCard>
          <ModuleSubtitle sx={{ p: 2, pb: 1 }}>Alertes actives</ModuleSubtitle>
          <Table size="small">
            <TableHead>
              <TableRow>
                {['Sévérité', 'Titre', 'Message', 'Source', 'Heure'].map((h) => (
                  <TableCell
                    key={h}
                    sx={{ color: colors.grayLight, borderColor: '#333', fontSize: '0.75rem', fontFamily: '"JetBrains Mono", monospace' }}
                  >
                    {h}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {alerts.map((alert) => (
                <TableRow key={alert.id} hover sx={{ '&:hover': { background: '#1f2235' } }}>
                  <TableCell sx={{ borderColor: '#333' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      {severityIcon(alert.severity)}
                      <Typography sx={{ color: severityColor(alert.severity), fontSize: '0.75rem', fontFamily: '"JetBrains Mono", monospace' }}>
                        {alert.severity.toUpperCase()}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell sx={{ color: '#fff', borderColor: '#333', fontFamily: '"JetBrains Mono", monospace', fontSize: '0.85rem' }}>
                    {alert.title}
                  </TableCell>
                  <TableCell sx={{ color: colors.grayLight, borderColor: '#333', fontSize: '0.8rem' }}>
                    {alert.message}
                  </TableCell>
                  <TableCell sx={{ borderColor: '#333' }}>
                    <Chip label={alert.source} size="small" sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.7rem' }} />
                  </TableCell>
                  <TableCell sx={{ color: colors.grayLight, borderColor: '#333', fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}>
                    {alert.timestamp.toLocaleTimeString('fr-FR')}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {alerts.length === 0 && (
            <Alert severity="success" sx={{ m: 2 }}>
              <AlertTitle>Aucune alerte</AlertTitle>
              Tous les services sont opérationnels et aucun pipeline n&apos;est en échec.
            </Alert>
          )}
        </ModuleCard>
      )}
    </Box>
  )
}
