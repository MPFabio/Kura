import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
} from '@mui/material'
import ModuleButton from './ModuleButton'

interface ScaleDeploymentDialogProps {
  open: boolean
  onClose: () => void
  namespace: string
  name: string
  currentReplicas: number
  onConfirm: (replicas: number) => void
  isPending?: boolean
}

export default function ScaleDeploymentDialog({
  open,
  onClose,
  namespace,
  name,
  currentReplicas,
  onConfirm,
  isPending = false,
}: ScaleDeploymentDialogProps) {
  const [replicas, setReplicas] = useState(currentReplicas)
  useEffect(() => {
    if (open) {
      setReplicas(currentReplicas)
    }
  }, [open, currentReplicas])

  const handleSubmit = () => {
    onConfirm(replicas)
    onClose()
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Scale Deployment</DialogTitle>
      <DialogContent>
        <p style={{ color: '#a0a0a0', marginBottom: 16, fontSize: '0.9rem' }}>
          {name} ({namespace})
        </p>
        <TextField
          fullWidth
          type="number"
          label="Répliques"
          value={replicas}
          onChange={(e) => {
            const val = parseInt(e.target.value, 10)
            if (!isNaN(val) && val >= 0) setReplicas(val)
          }}
          inputProps={{ min: 0, step: 1 }}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Annuler</Button>
        <ModuleButton onClick={handleSubmit} disabled={isPending}>
          {isPending ? 'En cours...' : 'Appliquer'}
        </ModuleButton>
      </DialogActions>
    </Dialog>
  )
}
