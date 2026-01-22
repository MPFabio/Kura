import { Box, Typography, Card, CardContent, Alert } from '@mui/material'

export default function AnsiblePage() {
  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 600 }}>
        Ansible
      </Typography>

      <Card>
        <CardContent>
          <Alert severity="info">
            Le service Ansible sera bientôt disponible. Cette page affichera les jobs Ansible Tower,
            les inventaires et l'historique d'exécution.
          </Alert>
        </CardContent>
      </Card>
    </Box>
  )
}
