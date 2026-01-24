import { Box, Grid, Alert } from '@mui/material'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import { ModuleSubtitle } from '../components/ModuleText'

// Données d'exemple pour les métriques
const sampleData = [
  { name: 'Lun', cpu: 45, memory: 60, requests: 1200 },
  { name: 'Mar', cpu: 52, memory: 65, requests: 1500 },
  { name: 'Mer', cpu: 48, memory: 62, requests: 1300 },
  { name: 'Jeu', cpu: 55, memory: 70, requests: 1800 },
  { name: 'Ven', cpu: 50, memory: 68, requests: 1600 },
  { name: 'Sam', cpu: 42, memory: 58, requests: 1100 },
  { name: 'Dim', cpu: 38, memory: 55, requests: 900 },
]

// Couleurs Abyssal Glow
const colors = {
  turquoise: '#00FFFF',
  turquoiseLight: '#00E0E0',
  violet: '#BF00FF',
  magenta: '#FF00BF',
  grayLight: '#A0A0A0',
  grayMedium: '#808080',
}

export default function MetricsPage() {
  return (
    <Box>
      <ModuleTitle>Métriques</ModuleTitle>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <ModuleCard sx={{ p: 3 }}>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Utilisation CPU et Mémoire (7 derniers jours)
            </ModuleSubtitle>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={sampleData}>
                <defs>
                  <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={colors.turquoise} stopOpacity={0.8} />
                    <stop offset="100%" stopColor={colors.turquoise} stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="memoryGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={colors.violet} stopOpacity={0.8} />
                    <stop offset="100%" stopColor={colors.violet} stopOpacity={0} />
                  </linearGradient>
                  <filter id="glow">
                    <feGaussianBlur stdDeviation="3" result="coloredBlur" />
                    <feMerge>
                      <feMergeNode in="coloredBlur" />
                      <feMergeNode in="SourceGraphic" />
                    </feMerge>
                  </filter>
                </defs>
                <CartesianGrid
                  strokeDasharray="2 2"
                  stroke={colors.grayMedium}
                  strokeOpacity={0.3}
                  horizontal={true}
                  vertical={true}
                />
                <XAxis
                  dataKey="name"
                  stroke={colors.grayLight}
                  style={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}
                  tick={{ fill: colors.grayLight }}
                />
                <YAxis
                  stroke={colors.grayLight}
                  domain={[0, 80]}
                  style={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}
                  tick={{ fill: colors.grayLight }}
                />
                <Tooltip
                  contentStyle={{
                    background: 'rgba(10, 10, 14, 0.95)',
                    border: `1px solid ${colors.turquoise}40`,
                    borderRadius: '8px',
                    color: '#FFFFFF',
                    fontFamily: '"JetBrains Mono", monospace',
                    boxShadow: `0 0 20px ${colors.turquoise}30`,
                  }}
                  labelStyle={{ color: colors.turquoise, fontFamily: '"Inter", sans-serif' }}
                  itemStyle={{ fontFamily: '"JetBrains Mono", monospace' }}
                />
                <Legend
                  wrapperStyle={{
                    fontFamily: '"JetBrains Mono", monospace',
                    fontSize: '0.875rem',
                    color: colors.grayLight,
                  }}
                  iconType="line"
                />
                <Line
                  type="monotone"
                  dataKey="cpu"
                  stroke={colors.turquoise}
                  strokeWidth={3}
                  name="CPU %"
                  dot={{
                    fill: colors.turquoise,
                    strokeWidth: 2,
                    r: 4,
                    filter: 'url(#glow)',
                  }}
                  activeDot={{
                    r: 6,
                    fill: colors.turquoise,
                    filter: 'url(#glow)',
                  }}
                  style={{ filter: `drop-shadow(0 0 8px ${colors.turquoise})` }}
                />
                <Line
                  type="monotone"
                  dataKey="memory"
                  stroke={colors.violet}
                  strokeWidth={3}
                  name="Mémoire %"
                  dot={{
                    fill: colors.violet,
                    strokeWidth: 2,
                    r: 4,
                    filter: 'url(#glow)',
                  }}
                  activeDot={{
                    r: 6,
                    fill: colors.violet,
                    filter: 'url(#glow)',
                  }}
                  style={{ filter: `drop-shadow(0 0 8px ${colors.violet})` }}
                />
              </LineChart>
            </ResponsiveContainer>
          </ModuleCard>
        </Grid>

        <Grid item xs={12}>
          <ModuleCard sx={{ p: 3 }}>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Requêtes par jour
            </ModuleSubtitle>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={sampleData}>
                <defs>
                  <linearGradient id="barGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={colors.turquoise} stopOpacity={1} />
                    <stop offset="100%" stopColor={colors.violet} stopOpacity={1} />
                  </linearGradient>
                  <filter id="barGlow">
                    <feGaussianBlur stdDeviation="4" result="coloredBlur" />
                    <feMerge>
                      <feMergeNode in="coloredBlur" />
                      <feMergeNode in="SourceGraphic" />
                    </feMerge>
                  </filter>
                </defs>
                <CartesianGrid
                  strokeDasharray="2 2"
                  stroke={colors.grayMedium}
                  strokeOpacity={0.3}
                  horizontal={true}
                  vertical={true}
                />
                <XAxis
                  dataKey="name"
                  stroke={colors.grayLight}
                  style={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}
                  tick={{ fill: colors.grayLight }}
                />
                <YAxis
                  stroke={colors.grayLight}
                  style={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }}
                  tick={{ fill: colors.grayLight }}
                />
                <Tooltip
                  contentStyle={{
                    background: 'rgba(10, 10, 14, 0.95)',
                    border: `1px solid ${colors.violet}40`,
                    borderRadius: '8px',
                    color: '#FFFFFF',
                    fontFamily: '"JetBrains Mono", monospace',
                    boxShadow: `0 0 20px ${colors.violet}30`,
                  }}
                  labelStyle={{ color: colors.violet, fontFamily: '"Inter", sans-serif' }}
                  itemStyle={{ fontFamily: '"JetBrains Mono", monospace' }}
                />
                <Legend
                  wrapperStyle={{
                    fontFamily: '"JetBrains Mono", monospace',
                    fontSize: '0.875rem',
                    color: colors.grayLight,
                  }}
                />
                <Bar
                  dataKey="requests"
                  fill="url(#barGradient)"
                  name="Requêtes"
                  radius={[8, 8, 0, 0]}
                  style={{ filter: 'url(#barGlow)' }}
                />
              </BarChart>
            </ResponsiveContainer>
          </ModuleCard>
        </Grid>

        <Grid item xs={12}>
          <ModuleCard sx={{ p: 2 }}>
            <Alert
              severity="info"
              sx={{
                background: 'rgba(10, 10, 14, 0.8)',
                border: `1px solid ${colors.turquoise}40`,
                borderRadius: '12px',
                color: colors.grayLight,
                '& .MuiAlert-icon': {
                  color: colors.turquoise,
                  filter: `drop-shadow(0 0 8px ${colors.turquoise})`,
                },
                fontFamily: '"JetBrains Mono", monospace',
              }}
            >
              Les métriques en temps réel seront disponibles une fois le service Metrics
              implémenté. Les données affichées sont des exemples.
            </Alert>
          </ModuleCard>
        </Grid>
      </Grid>
    </Box>
  )
}
