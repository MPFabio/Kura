import { Box, Alert } from '@mui/material'
import { useSocket } from '../contexts/SocketContext'
import { format } from 'date-fns'
import fr from 'date-fns/locale/fr'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ModuleCaption } from '../components/ModuleText'

export default function AlertsPage() {
  const { events } = useSocket()
  const alerts = events.filter((e) => e.type === 'alert')

  return (
    <Box>
      <ModuleTitle>Alertes</ModuleTitle>

      <ModuleCard>
        {alerts.length > 0 ? (
          <Box>
            {alerts.slice().reverse().map((alert, index) => (
              <Alert
                key={index}
                severity="warning"
                sx={{ mb: 2 }}
                action={
                  <ModuleCaption>
                    {format(new Date(alert.timestamp), 'PPpp', { locale: fr })}
                  </ModuleCaption>
                }
              >
                {JSON.stringify(alert.data, null, 2)}
              </Alert>
            ))}
          </Box>
        ) : (
          <Alert severity="info">Aucune alerte récente</Alert>
        )}
      </ModuleCard>
    </Box>
  )
}
