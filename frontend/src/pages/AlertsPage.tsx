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
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle } from '../components/ModuleText'
import { metricsService, type ServiceHealth } from '../services/metricsService'
import { pipelineService, type PipelineRun } from '../services/pipelineService'
import { kuraColors } from '../theme'

interface KuraAlert {
  id: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  message: string
  source: string
  timestamp: Date
}

const severityConfig = {
  critical: { label: 'Critique', color: kuraColors.error },
  warning:  { label: 'Warning',  color: kuraColors.warning },
  info:     { label: 'Info',     color: kuraColors.success },
}

function SeverityBadge({ severity }: { severity: KuraAlert['severity'] }) {
  const { label, color } = severityConfig[severity]
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
        letterSpacing: '0.01em',
      }}
    />
  )
}

function SummaryBadge({ count, label, color }: { count: number; label: string; color: string }) {
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Box
        sx={{
          minWidth: 22,
          height: 22,
          borderRadius: '4px',
          bgcolor: `${color}22`,
          border: `1px solid ${color}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          px: 0.75,
        }}
      >
        <Typography sx={{ fontSize: '0.75rem', fontWeight: 700, color, fontFamily: '"JetBrains Mono", monospace', lineHeight: 1 }}>
          {count}
        </Typography>
      </Box>
      <Typography sx={{ fontSize: '0.8125rem', color: kuraColors.text2, fontWeight: 400 }}>
        {label}
      </Typography>
    </Box>
  )
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<KuraAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const buildAlerts = async () => {
    const result: KuraAlert[] = []

    try {
      const health: ServiceHealth[] = await metricsService.getHealth()
      const down = health.filter((s) => !s.up)
      const up = health.filter((s) => s.up)

      down.forEach((svc) => {
        result.push({
          id: `down-${svc.job}`,
          severity: 'critical',
          title: `Service hors ligne — ${svc.name}`,
          message: `${svc.job} ne répond plus à son health check`,
          source: 'Monitoring',
          timestamp: new Date(),
        })
      })

      if (down.length === 0 && up.length > 0) {
        result.push({
          id: 'all-up',
          severity: 'info',
          title: 'Tous les services opérationnels',
          message: `${up.length} service${up.length > 1 ? 's' : ''} actif${up.length > 1 ? 's' : ''}`,
          source: 'Monitoring',
          timestamp: new Date(),
        })
      }
    } catch {
      result.push({
        id: 'metrics-error',
        severity: 'warning',
        title: 'Monitoring indisponible',
        message: 'Le metrics-service ne répond pas',
        source: 'Monitoring',
        timestamp: new Date(),
      })
    }

    try {
      const runs = await pipelineService.getRuns({ limit: 20 })
      const recentRuns: PipelineRun[] = runs?.runs ?? []
      const failed = recentRuns.filter((r) => (r.status as string) === 'failure' || (r.status as string) === 'failed')
      const seen = new Set<string>()
      failed.forEach((r) => {
        const key = `${r.repository}-${r.workflow_name}`
        if (seen.has(key)) return
        seen.add(key)
        result.push({
          id: `pipeline-${r.id}`,
          severity: 'warning',
          title: `Pipeline en échec — ${r.workflow_name}`,
          message: `${r.repository} · ${r.branch}`,
          source: 'Pipelines',
          timestamp: r.started_at ? new Date(r.started_at) : new Date(),
        })
      })
    } catch { /* Pipeline service indisponible */ }

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

  const critical = alerts.filter((a) => a.severity === 'critical').length
  const warnings = alerts.filter((a) => a.severity === 'warning').length
  const infos = alerts.filter((a) => a.severity === 'info').length

  return (
    <Box>
      <ModuleTitle>Alertes</ModuleTitle>

      {/* Résumé compact */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 3, mb: 3, flexWrap: 'wrap' }}>
        <SummaryBadge count={critical} label={`critique${critical !== 1 ? 's' : ''}`} color={kuraColors.error} />
        <SummaryBadge count={warnings} label={`warning${warnings !== 1 ? 's' : ''}`} color={kuraColors.warning} />
        <SummaryBadge count={infos} label={`info${infos !== 1 ? 's' : ''}`} color={kuraColors.success} />
        <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, ml: 'auto', fontFamily: '"JetBrains Mono", monospace' }}>
          {lastRefresh.toLocaleTimeString('fr-FR')} · rafraîchissement auto toutes les 30s
        </Typography>
      </Box>

      {loading ? (
        <Box display="flex" justifyContent="center" py={6}>
          <CircularProgress size={28} sx={{ color: kuraColors.accent }} />
        </Box>
      ) : (
        <ModuleCard>
          <Box sx={{ px: 2.5, py: 2, borderBottom: `1px solid ${kuraColors.border0}` }}>
            <ModuleSubtitle>Alertes actives</ModuleSubtitle>
          </Box>
          <Table size="small">
            <TableHead>
              <TableRow>
                {['Sévérité', 'Titre', 'Message', 'Source', 'Heure'].map((h) => (
                  <TableCell key={h}>{h}</TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {alerts.map((alert) => (
                <TableRow key={alert.id}>
                  <TableCell sx={{ width: 100 }}>
                    <SeverityBadge severity={alert.severity} />
                  </TableCell>
                  <TableCell sx={{ fontWeight: 500, color: kuraColors.text0 }}>
                    {alert.title}
                  </TableCell>
                  <TableCell sx={{ color: kuraColors.text2 }}>
                    {alert.message}
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={alert.source}
                      size="small"
                      sx={{
                        fontSize: '0.6875rem',
                        bgcolor: kuraColors.bg3,
                        color: kuraColors.text1,
                        border: `1px solid ${kuraColors.border1}`,
                      }}
                    />
                  </TableCell>
                  <TableCell sx={{ color: kuraColors.text2, fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem', whiteSpace: 'nowrap' }}>
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
