import { useQuery } from '@tanstack/react-query'
import {
  Grid,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Card,
  CardContent,
} from '@mui/material'
import {
  Cloud as CloudIcon,
  Storage as StorageIcon,
  PlayArrow as PlayArrowIcon,
  Timeline as TimelineIcon,
  Notifications as NotificationsIcon,
  BarChart as BarChartIcon,
} from '@mui/icons-material'
import { k8sService } from '../services/k8sService'
import { useSocket } from '../contexts/SocketContext'
import { format } from 'date-fns'
import fr from 'date-fns/locale/fr'
import ModuleCard from '../components/ModuleCard'

export default function DashboardPage() {
  const { data: namespaces, isLoading } = useQuery({
    queryKey: ['namespaces'],
    queryFn: () => k8sService.getNamespaces(),
  })

  const { events, connected } = useSocket()

  const stats = [
    {
      title: 'Namespaces K8s',
      value: namespaces?.items?.length || 0,
      icon: <CloudIcon sx={{ fontSize: 40 }} />,
      color: '#00FFFF',
      active: true,
    },
    {
      title: 'Événements temps réel',
      value: events.length,
      icon: <NotificationsIcon sx={{ fontSize: 40 }} />,
      color: '#BF00FF',
      active: events.length > 0,
    },
    {
      title: 'WebSocket',
      value: connected ? 'Connecté' : 'Déconnecté',
      icon: <TimelineIcon sx={{ fontSize: 40 }} />,
      color: connected ? '#00FFFF' : '#FF4500',
      active: connected,
    },
  ]

  return (
    <Box>
      <Typography 
        variant="h4" 
        sx={{ 
          mb: 4, 
          fontWeight: 600,
          background: 'linear-gradient(135deg, #00FFFF, #BF00FF)',
          backgroundClip: 'text',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          textShadow: 'none',
          letterSpacing: '0.02em',
        }}
      >
        Dashboard
      </Typography>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        {stats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="body2">
                      {stat.title}
                    </Typography>
                    <Typography variant="h4" sx={{ color: stat.color }}>
                      {stat.value}
                    </Typography>
                  </Box>
                  <Box sx={{ color: stat.color }}>{stat.icon}</Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <ModuleCard sx={{ p: 3 }}>
            <Typography 
              variant="h6" 
              sx={{ 
                mb: 2,
                color: '#00FFFF',
                fontFamily: '"Inter", sans-serif',
                fontWeight: 600,
                textShadow: '0 0 10px rgba(0, 255, 255, 0.3)',
              }}
            >
              Namespaces Kubernetes
            </Typography>
              {isLoading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                  <CircularProgress />
                </Box>
              ) : namespaces?.items && namespaces.items.length > 0 ? (
                <Box>
                  {namespaces.items.map((ns) => (
                    <Box
                      key={ns.name}
                      sx={{
                        p: 2,
                        mb: 1,
                        bgcolor: 'rgba(0, 255, 255, 0.05)',
                        border: '1px solid rgba(0, 255, 255, 0.2)',
                        borderRadius: 2,
                        transition: 'all 0.3s ease',
                        '&:hover': {
                          bgcolor: 'rgba(0, 255, 255, 0.1)',
                          borderColor: 'rgba(0, 255, 255, 0.4)',
                          boxShadow: '0 0 15px rgba(0, 255, 255, 0.2)',
                        },
                      }}
                    >
                      <Typography 
                        variant="body1" 
                        sx={{ 
                          fontWeight: 500,
                          fontFamily: '"JetBrains Mono", monospace',
                          color: '#FFFFFF',
                        }}
                      >
                        {ns.name}
                      </Typography>
                      {ns.status && (
                        <Typography 
                          variant="body2" 
                          sx={{ 
                            color: '#A0A0A0',
                            fontFamily: '"JetBrains Mono", monospace',
                            fontSize: '0.75rem',
                            mt: 0.5,
                          }}
                        >
                          Statut: {ns.status}
                        </Typography>
                      )}
                    </Box>
                  ))}
                </Box>
              ) : (
                <Alert severity="info">Aucun namespace trouvé</Alert>
              )}
          </ModuleCard>
        </Grid>

        <Grid item xs={12} md={6}>
          <ModuleCard sx={{ p: 3 }}>
            <Typography 
              variant="h6" 
              sx={{ 
                mb: 2,
                color: '#BF00FF',
                fontFamily: '"Inter", sans-serif',
                fontWeight: 600,
                textShadow: '0 0 10px rgba(191, 0, 255, 0.3)',
              }}
            >
              Événements récents
            </Typography>
              {events.length > 0 ? (
                <Box>
                  {events.slice(-5).reverse().map((event, index) => (
                    <Box
                      key={index}
                      sx={{
                        p: 2,
                        mb: 1,
                        bgcolor: 'rgba(191, 0, 255, 0.05)',
                        border: '1px solid rgba(191, 0, 255, 0.2)',
                        borderRadius: 2,
                        transition: 'all 0.3s ease',
                        '&:hover': {
                          bgcolor: 'rgba(191, 0, 255, 0.1)',
                          borderColor: 'rgba(191, 0, 255, 0.4)',
                          boxShadow: '0 0 15px rgba(191, 0, 255, 0.2)',
                        },
                      }}
                    >
                      <Typography 
                        variant="body2" 
                        sx={{ 
                          fontWeight: 500,
                          fontFamily: '"JetBrains Mono", monospace',
                          color: '#FFFFFF',
                        }}
                      >
                        {event.type}
                      </Typography>
                      <Typography 
                        variant="caption" 
                        sx={{
                          color: '#A0A0A0',
                          fontFamily: '"JetBrains Mono", monospace',
                          fontSize: '0.7rem',
                          mt: 0.5,
                          display: 'block',
                        }}
                      >
                        {format(new Date(event.timestamp), 'PPpp', { locale: fr })}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              ) : (
                <Alert severity="info">Aucun événement récent</Alert>
              )}
          </ModuleCard>
        </Grid>
      </Grid>
    </Box>
  )
}
