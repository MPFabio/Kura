import { useQuery } from '@tanstack/react-query'
import {
  Grid,
  Box,
  CircularProgress,
  Alert,
  Card,
  CardContent,
} from '@mui/material'
import {
  Cloud as CloudIcon,
  Timeline as TimelineIcon,
  Notifications as NotificationsIcon,
} from '@mui/icons-material'
import { k8sService } from '../services/k8sService'
import { useSocket } from '../contexts/SocketContext'
import { format } from 'date-fns'
import fr from 'date-fns/locale/fr'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import { ModuleSubtitle, ModuleBodyText, ModuleSecondaryText, ModuleCaption } from '../components/ModuleText'

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
      color: '#4F8EF7',
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
      color: connected ? '#4F8EF7' : '#FF4500',
      active: connected,
    },
  ]

  return (
    <Box>
      <ModuleTitle>Dashboard</ModuleTitle>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        {stats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Box>
                    <ModuleSecondaryText sx={{ mb: 1 }}>
                      {stat.title}
                    </ModuleSecondaryText>
                    <ModuleSubtitle sx={{ color: stat.color, fontSize: '2rem', fontWeight: 700 }}>
                      {stat.value}
                    </ModuleSubtitle>
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
            <ModuleSubtitle sx={{ mb: 2 }}>
              Namespaces Kubernetes
            </ModuleSubtitle>
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
                      <ModuleBodyText>
                        {ns.name}
                      </ModuleBodyText>
                      {ns.status && (
                        <ModuleSecondaryText sx={{ mt: 0.5, fontSize: '0.875rem' }}>
                          Statut: {ns.status}
                        </ModuleSecondaryText>
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
            <ModuleSubtitle sx={{ mb: 2 }}>
              Événements récents
            </ModuleSubtitle>
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
                      <ModuleBodyText>
                        {event.type}
                      </ModuleBodyText>
                      <ModuleCaption sx={{ mt: 0.5, display: 'block' }}>
                        {format(new Date(event.timestamp), 'PPpp', { locale: fr })}
                      </ModuleCaption>
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
