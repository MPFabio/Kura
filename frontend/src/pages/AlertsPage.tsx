import { Box, Typography, Card, CardContent, Alert } from '@mui/material'
import { useSocket } from '../contexts/SocketContext'
import { format } from 'date-fns'
import fr from 'date-fns/locale/fr'

export default function AlertsPage() {
  const { events } = useSocket()
  const alerts = events.filter((e) => e.type === 'alert')

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 600 }}>
        Alertes
      </Typography>

      <Card>
        <CardContent>
          {alerts.length > 0 ? (
            <Box>
              {alerts.slice().reverse().map((alert, index) => (
                <Alert
                  key={index}
                  severity="warning"
                  sx={{ mb: 2 }}
                  action={
                    <Typography variant="caption" color="textSecondary">
                      {format(new Date(alert.timestamp), 'PPpp', { locale: fr })}
                    </Typography>
                  }
                >
                  {JSON.stringify(alert.data, null, 2)}
                </Alert>
              ))}
            </Box>
          ) : (
            <Alert severity="info">Aucune alerte récente</Alert>
          )}
        </CardContent>
      </Card>
    </Box>
  )
}
