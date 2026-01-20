import { Box, Typography, Card, CardContent, Alert } from '@mui/material'

export default function PipelinePage() {
  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 600 }}>
        Pipelines CI/CD
      </Typography>

      <Card>
        <CardContent>
          <Alert severity="info">
            Le service Pipelines sera bientôt disponible. Cette page affichera les pipelines
            GitHub Actions, GitLab CI et Jenkins avec leurs statuts et historiques.
          </Alert>
        </CardContent>
      </Card>
    </Box>
  )
}
