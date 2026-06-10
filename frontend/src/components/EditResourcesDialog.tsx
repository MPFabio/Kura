import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Typography,
} from '@mui/material'
import ModuleButton from './ModuleButton'

export interface ResourceValues {
  cpu_request: string
  cpu_limit: string
  mem_request: string
  mem_limit: string
}

interface EditResourcesDialogProps {
  open: boolean
  onClose: () => void
  namespace: string
  deploymentName: string
  containers: { name: string }[]
  resources: ResourceValues
  containerName: string
  onConfirm: (container: string, resources: ResourceValues) => void
  isPending?: boolean
}

export default function EditResourcesDialog({
  open,
  onClose,
  namespace,
  deploymentName,
  containers,
  resources,
  containerName,
  onConfirm,
  isPending = false,
}: EditResourcesDialogProps) {
  const [selectedContainer, setSelectedContainer] = useState(containerName)
  const [values, setValues] = useState<ResourceValues>(resources)

  useEffect(() => {
    if (open) {
      setSelectedContainer(containerName)
      setValues(resources)
    }
  }, [open, containerName, resources])

  const handleChange = (field: keyof ResourceValues, value: string) => {
    setValues({ ...values, [field]: value })
  }

  const handleSubmit = () => {
    onConfirm(selectedContainer, values)
    onClose()
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Modifier les ressources (CPU / mémoire)</DialogTitle>
      <DialogContent>
        <p style={{ color: '#a0a0a0', marginBottom: 16, fontSize: '0.9rem' }}>
          {deploymentName} ({namespace})
        </p>
        <FormControl fullWidth sx={{ mb: 2 }}>
          <InputLabel>Container</InputLabel>
          <Select
            value={selectedContainer}
            label="Container"
            onChange={(e) => setSelectedContainer(e.target.value)}
          >
            {containers.map((c) => (
              <MenuItem key={c.name} value={c.name}>
                {c.name}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <Typography variant="subtitle2" sx={{ mb: 1, color: '#a0a0a0' }}>
          Requests
        </Typography>
        <Box sx={{ display: 'flex', gap: 1.5, mb: 2 }}>
          <TextField
            size="small"
            label="CPU request"
            placeholder="ex: 100m"
            value={values.cpu_request}
            onChange={(e) => handleChange('cpu_request', e.target.value)}
            sx={{ flex: 1 }}
          />
          <TextField
            size="small"
            label="Mémoire request"
            placeholder="ex: 128Mi"
            value={values.mem_request}
            onChange={(e) => handleChange('mem_request', e.target.value)}
            sx={{ flex: 1 }}
          />
        </Box>

        <Typography variant="subtitle2" sx={{ mb: 1, color: '#a0a0a0' }}>
          Limits
        </Typography>
        <Box sx={{ display: 'flex', gap: 1.5 }}>
          <TextField
            size="small"
            label="CPU limit"
            placeholder="ex: 500m"
            value={values.cpu_limit}
            onChange={(e) => handleChange('cpu_limit', e.target.value)}
            sx={{ flex: 1 }}
          />
          <TextField
            size="small"
            label="Mémoire limit"
            placeholder="ex: 256Mi"
            value={values.mem_limit}
            onChange={(e) => handleChange('mem_limit', e.target.value)}
            sx={{ flex: 1 }}
          />
        </Box>
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
