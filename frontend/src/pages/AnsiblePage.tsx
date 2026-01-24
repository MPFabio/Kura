import { Box, Alert } from '@mui/material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'

export default function AnsiblePage() {
  return (
    <Box>
      <ModuleTitle>Ansible</ModuleTitle>

      <ModuleCard>
        <Alert severity="info">
          Le service Ansible sera bientôt disponible. Cette page affichera les jobs Ansible Tower,
          les inventaires et l'historique d'exécution.
        </Alert>
      </ModuleCard>
    </Box>
  )
}
