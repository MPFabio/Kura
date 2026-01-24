import { Box, Alert } from '@mui/material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'

export default function PipelinePage() {
  return (
    <Box>
      <ModuleTitle>Pipelines CI/CD</ModuleTitle>

      <ModuleCard>
        <Alert 
          severity="info"
          sx={{
            borderRadius: 3,
            animation: 'jellyfishFloat 14s ease-in-out infinite',
          }}
        >
          Le service Pipelines sera bientôt disponible. Cette page affichera les pipelines
          GitHub Actions, GitLab CI et Jenkins avec leurs statuts et historiques.
        </Alert>
      </ModuleCard>
    </Box>
  )
}
