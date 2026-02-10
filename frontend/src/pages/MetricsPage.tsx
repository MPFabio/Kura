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
                    background: '#2c2f3f',
                    border: `1px solid ${colors.turquoise}`,
                    borderRadius: '0',
                    color: '#FFFFFF',
                    fontFamily: '"JetBrains Mono", monospace',
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
                  strokeWidth={2}
                  name="CPU %"
                  dot={{
                    fill: colors.turquoise,
                    strokeWidth: 0,
                    r: 3,
                  }}
                  activeDot={{
                    r: 5,
                    fill: colors.turquoise,
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="memory"
                  stroke={colors.magenta}
                  strokeWidth={2}
                  name="Mémoire %"
                  dot={{
                    fill: colors.magenta,
                    strokeWidth: 0,
                    r: 3,
                  }}
                  activeDot={{
                    r: 5,
                    fill: colors.magenta,
                  }}
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
                    background: '#2c2f3f',
                    border: `1px solid ${colors.turquoise}`,
                    borderRadius: '0',
                    color: '#FFFFFF',
                    fontFamily: '"JetBrains Mono", monospace',
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
                />
                <Bar
                  dataKey="requests"
                  fill={colors.turquoise}
                  name="Requêtes"
                  radius={[0, 0, 0, 0]}
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
                background: '#2c2f3f',
                border: `1px solid ${colors.turquoise}`,
                borderRadius: '0',
                color: colors.grayLight,
                '& .MuiAlert-icon': {
                  color: colors.turquoise,
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
