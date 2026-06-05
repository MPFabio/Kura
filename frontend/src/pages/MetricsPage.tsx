import { useEffect, useState } from 'react'
import {
  Box,
  Grid,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  CircularProgress,
  Typography,
  Alert,
} from '@mui/material'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import { ModuleSubtitle } from '../components/ModuleText'
import {
  metricsService,
  type ServiceHealth,
  type ServiceMetric,
  type Overview,
} from '../services/metricsService'
import { kuraColors } from '../theme'

const isProd = window.location.protocol === 'https:'
const GRAFANA_URL = isProd
  ? `${window.location.origin}/grafana`
  : ((import.meta as any).env?.VITE_GRAFANA_URL ?? 'http://localhost:3000')
const GRAFANA_DASHBOARD_URL = `${GRAFANA_URL}/d/kura-overview/kura-e28094-platform-overview?orgId=1&kiosk=tv&refresh=30s`

function HealthBadge({ up }: { up: boolean }) {
  const color = up ? kuraColors.success : kuraColors.error
  const label = up ? 'UP' : 'DOWN'
  return (
    <Chip
      size="small"
      label={label}
      sx={{
        fontWeight: 600,
        fontSize: '0.6875rem',
        fontFamily: '"JetBrains Mono", monospace',
        backgroundColor: `${color}22`,
        border: `1px solid ${color}`,
        color,
      }}
    />
  )
}

function KPICard({ label, value, unit }: { label: string; value: string | number; unit?: string }) {
  return (
    <ModuleCard sx={{ p: 2.5, textAlign: 'center' }}>
      <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, mb: 1 }}>
        {label}
      </Typography>
      <Typography sx={{ fontSize: '1.75rem', color: kuraColors.text0, fontWeight: 600, lineHeight: 1 }}>
        {value}
        {unit && <Typography component="span" sx={{ fontSize: '0.875rem', color: kuraColors.text2, ml: 0.5 }}>{unit}</Typography>}
      </Typography>
    </ModuleCard>
  )
}

export default function MetricsPage() {
  const [health, setHealth] = useState<ServiceHealth[]>([])
  const [services, setServices] = useState<ServiceMetric[]>([])
  const [overview, setOverview] = useState<Overview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      try {
        const [h, s, o] = await Promise.all([
          metricsService.getHealth(),
          metricsService.getServices(),
          metricsService.getOverview(),
        ])
        setHealth(h)
        setServices(s)
        setOverview(o)
        setError(null)
      } catch {
        setError('Impossible de joindre le metrics-service.')
      } finally {
        setLoading(false)
      }
    }
    load()
    const interval = setInterval(load, 30_000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress size={28} sx={{ color: kuraColors.accent }} />
      </Box>
    )
  }

  return (
    <Box>
      <ModuleTitle>Monitoring</ModuleTitle>

      {error && <Alert severity="warning" sx={{ mb: 3 }}>{error}</Alert>}

      {/* KPI */}
      {overview && (
        <Grid container spacing={2} sx={{ mb: 3 }}>
          {[
            { label: 'Services actifs', value: overview.services_up },
            { label: 'Hors ligne', value: overview.services_down },
            { label: 'Goroutines', value: Math.round(overview.total_goroutines) },
            { label: 'Mémoire', value: overview.total_memory_mb.toFixed(0), unit: 'MB' },
          ].map((kpi) => (
            <Grid item xs={6} sm={3} key={kpi.label}>
              <KPICard label={kpi.label} value={kpi.value} unit={(kpi as any).unit} />
            </Grid>
          ))}
        </Grid>
      )}

      {/* Health cards */}
      {health.length > 0 && (
        <Grid container spacing={2} sx={{ mb: 3 }}>
          {health.map((svc) => (
            <Grid item xs={12} sm={6} md={4} key={svc.job}>
              <ModuleCard sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography sx={{ color: kuraColors.text0, fontWeight: 500, fontSize: '0.9rem', mb: 0.25 }}>
                    {svc.name}
                  </Typography>
                  <Typography sx={{ color: kuraColors.text2, fontSize: '0.75rem', fontFamily: '"JetBrains Mono", monospace' }}>
                    {svc.goroutines > 0 ? `${Math.round(svc.goroutines)} goroutines` : '—'}
                    {svc.memory_mb > 0 ? ` · ${svc.memory_mb.toFixed(1)} MB` : ''}
                  </Typography>
                </Box>
                <HealthBadge up={svc.up} />
              </ModuleCard>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Tableau métriques */}
      {services.length > 0 && (
        <ModuleCard sx={{ mb: 3 }}>
          <Box sx={{ px: 2.5, py: 2, borderBottom: `1px solid ${kuraColors.border0}` }}>
            <ModuleSubtitle>Métriques par service</ModuleSubtitle>
          </Box>
          <Table size="small">
            <TableHead>
              <TableRow>
                {['Service', 'État', 'Goroutines', 'CPU (rate)', 'Mémoire (MB)'].map((h) => (
                  <TableCell key={h}>{h}</TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {services.map((svc) => (
                <TableRow key={svc.job}>
                  <TableCell sx={{ fontWeight: 500, color: kuraColors.text0 }}>{svc.name}</TableCell>
                  <TableCell><HealthBadge up={svc.up} /></TableCell>
                  <TableCell sx={{ color: kuraColors.text1, fontFamily: '"JetBrains Mono", monospace' }}>
                    {svc.goroutines > 0 ? Math.round(svc.goroutines) : '—'}
                  </TableCell>
                  <TableCell sx={{ color: kuraColors.text1, fontFamily: '"JetBrains Mono", monospace' }}>
                    {svc.cpu_rate > 0 ? svc.cpu_rate.toFixed(4) : '—'}
                  </TableCell>
                  <TableCell sx={{ color: kuraColors.text1, fontFamily: '"JetBrains Mono", monospace' }}>
                    {svc.memory_mb > 0 ? svc.memory_mb.toFixed(1) : '—'}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </ModuleCard>
      )}

      {/* Grafana iframe */}
      <ModuleCard sx={{ p: 0, overflow: 'hidden' }}>
        <Box sx={{ px: 2.5, py: 2, borderBottom: `1px solid ${kuraColors.border0}` }}>
          <ModuleSubtitle>Dashboard Grafana — Kura Platform Overview</ModuleSubtitle>
        </Box>
        <iframe
          src={GRAFANA_DASHBOARD_URL}
          title="Kura Grafana Dashboard"
          width="100%"
          height="600"
          style={{ border: 'none', display: 'block' }}
          sandbox="allow-scripts allow-same-origin allow-forms"
        />
      </ModuleCard>
    </Box>
  )
}
