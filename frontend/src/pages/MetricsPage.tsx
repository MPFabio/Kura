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
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import CancelIcon from '@mui/icons-material/Cancel'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import { ModuleSubtitle } from '../components/ModuleText'
import {
  metricsService,
  type ServiceHealth,
  type ServiceMetric,
  type Overview,
} from '../services/metricsService'

const colors = {
  turquoise: '#00FFFF',
  violet: '#BF00FF',
  magenta: '#FF00BF',
  green: '#00FF88',
  red: '#FF4444',
  grayLight: '#A0A0A0',
  grayMedium: '#808080',
  cardBg: '#1a1d2e',
}

// URL Grafana — pointe vers le dashboard kura-overview en mode kiosk.
// En local : http://localhost:3000 ; en prod, adapter via variable d'env.
// En production, Grafana est exposé à /grafana via le reverse proxy.
// En développement local, on utilise VITE_GRAFANA_URL ou localhost:3000.
const isProd = window.location.protocol === 'https:'
const GRAFANA_URL = isProd
  ? `${window.location.origin}/grafana`
  : ((import.meta as any).env?.VITE_GRAFANA_URL ?? 'http://localhost:3000')
const GRAFANA_DASHBOARD_URL = `${GRAFANA_URL}/d/kura-overview/kura-e28094-platform-overview?orgId=1&kiosk=tv&refresh=30s`

function HealthBadge({ up }: { up: boolean }) {
  return up ? (
    <Chip
      icon={<CheckCircleIcon style={{ color: colors.green }} />}
      label="UP"
      size="small"
      sx={{ color: colors.green, borderColor: colors.green, fontFamily: '"JetBrains Mono", monospace' }}
      variant="outlined"
    />
  ) : (
    <Chip
      icon={<CancelIcon style={{ color: colors.red }} />}
      label="DOWN"
      size="small"
      sx={{ color: colors.red, borderColor: colors.red, fontFamily: '"JetBrains Mono", monospace' }}
      variant="outlined"
    />
  )
}

function KPICard({ label, value, unit }: { label: string; value: string | number; unit?: string }) {
  return (
    <ModuleCard sx={{ p: 2.5, textAlign: 'center' }}>
      <Typography
        sx={{ fontSize: '0.75rem', color: colors.grayLight, fontFamily: '"JetBrains Mono", monospace', mb: 1 }}
      >
        {label}
      </Typography>
      <Typography
        sx={{ fontSize: '2rem', color: colors.turquoise, fontFamily: '"JetBrains Mono", monospace', fontWeight: 700 }}
      >
        {value}
        {unit && (
          <Typography component="span" sx={{ fontSize: '0.875rem', color: colors.grayMedium, ml: 0.5 }}>
            {unit}
          </Typography>
        )}
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
      } catch (err: any) {
        setError('Impossible de joindre le metrics-service. Vérifiez que la stack est démarrée.')
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
        <CircularProgress sx={{ color: colors.turquoise }} />
      </Box>
    )
  }

  return (
    <Box>
      <ModuleTitle>Monitoring</ModuleTitle>

      {error && (
        <Alert
          severity="warning"
          sx={{ mb: 3, background: '#2c2f3f', border: `1px solid ${colors.violet}`, color: colors.grayLight }}
        >
          {error}
        </Alert>
      )}

      {/* Section 1 : KPI Overview */}
      {overview && (
        <Grid container spacing={2} sx={{ mb: 3 }}>
          <Grid item xs={6} sm={3}>
            <KPICard label="Services actifs" value={overview.services_up} />
          </Grid>
          <Grid item xs={6} sm={3}>
            <KPICard label="Services hors ligne" value={overview.services_down} />
          </Grid>
          <Grid item xs={6} sm={3}>
            <KPICard label="Goroutines totales" value={Math.round(overview.total_goroutines)} />
          </Grid>
          <Grid item xs={6} sm={3}>
            <KPICard label="Mémoire totale" value={overview.total_memory_mb.toFixed(0)} unit="MB" />
          </Grid>
        </Grid>
      )}

      {/* Section 2 : Health Cards */}
      {health.length > 0 && (
        <Grid container spacing={2} sx={{ mb: 3 }}>
          {health.map((svc) => (
            <Grid item xs={12} sm={6} md={4} key={svc.job}>
              <ModuleCard sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography
                    sx={{ fontFamily: '"JetBrains Mono", monospace', color: '#fff', fontWeight: 600, fontSize: '0.9rem' }}
                  >
                    {svc.name}
                  </Typography>
                  <Typography sx={{ fontFamily: '"JetBrains Mono", monospace', color: colors.grayLight, fontSize: '0.72rem' }}>
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

      {/* Section 3 : Tableau détaillé par service */}
      {services.length > 0 && (
        <ModuleCard sx={{ mb: 3 }}>
          <ModuleSubtitle sx={{ p: 2, pb: 0 }}>Métriques par service</ModuleSubtitle>
          <Table size="small" sx={{ fontFamily: '"JetBrains Mono", monospace' }}>
            <TableHead>
              <TableRow>
                {['Service', 'État', 'Goroutines', 'CPU (rate)', 'Mémoire (MB)'].map((h) => (
                  <TableCell
                    key={h}
                    sx={{ color: colors.grayLight, borderColor: '#333', fontSize: '0.75rem', fontFamily: 'inherit' }}
                  >
                    {h}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {services.map((svc) => (
                <TableRow key={svc.job} hover sx={{ '&:hover': { background: '#1f2235' } }}>
                  <TableCell sx={{ color: '#fff', borderColor: '#333', fontFamily: '"JetBrains Mono", monospace' }}>
                    {svc.name}
                  </TableCell>
                  <TableCell sx={{ borderColor: '#333' }}>
                    <HealthBadge up={svc.up} />
                  </TableCell>
                  <TableCell sx={{ color: colors.turquoise, borderColor: '#333', fontFamily: 'inherit' }}>
                    {svc.goroutines > 0 ? Math.round(svc.goroutines) : '—'}
                  </TableCell>
                  <TableCell sx={{ color: colors.magenta, borderColor: '#333', fontFamily: 'inherit' }}>
                    {svc.cpu_rate > 0 ? svc.cpu_rate.toFixed(4) : '—'}
                  </TableCell>
                  <TableCell sx={{ color: colors.violet, borderColor: '#333', fontFamily: 'inherit' }}>
                    {svc.memory_mb > 0 ? svc.memory_mb.toFixed(1) : '—'}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </ModuleCard>
      )}

      {/* Section 4 : Dashboard Grafana iframe */}
      <ModuleCard sx={{ p: 0, overflow: 'hidden' }}>
        <Box sx={{ p: 2, borderBottom: `1px solid #333` }}>
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
