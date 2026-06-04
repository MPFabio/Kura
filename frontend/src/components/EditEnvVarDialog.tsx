import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  IconButton,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material'
import { Add as AddIcon, Delete as DeleteIcon } from '@mui/icons-material'
import ModuleButton from './ModuleButton'

export interface EnvVar {
  name: string
  value: string
}

interface EditEnvVarDialogProps {
  open: boolean
  onClose: () => void
  namespace: string
  deploymentName: string
  containers: { name: string }[]
  envVars: EnvVar[]
  containerName: string
  onConfirm: (container: string, env: EnvVar[]) => void
  isPending?: boolean
}

export default function EditEnvVarDialog({
  open,
  onClose,
  namespace,
  deploymentName,
  containers,
  envVars,
  containerName,
  onConfirm,
  isPending = false,
}: EditEnvVarDialogProps) {
  const [selectedContainer, setSelectedContainer] = useState(containerName)
  const [vars, setVars] = useState<EnvVar[]>([])

  useEffect(() => {
    if (open) {
      setSelectedContainer(containerName)
      setVars(envVars.length > 0 ? envVars.map((e) => ({ ...e })) : [{ name: '', value: '' }])
    }
  }, [open, containerName, envVars])

  const handleAdd = () => {
    setVars([...vars, { name: '', value: '' }])
  }

  const handleRemove = (idx: number) => {
    setVars(vars.filter((_, i) => i !== idx))
  }

  const handleChange = (idx: number, field: 'name' | 'value', value: string) => {
    const next = [...vars]
    next[idx] = { ...next[idx], [field]: value }
    setVars(next)
  }

  const handleSubmit = () => {
    const valid = vars.filter((v) => v.name.trim())
    onConfirm(selectedContainer, valid)
    onClose()
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Modifier les variables d&apos;environnement</DialogTitle>
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
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
          {vars.map((v, idx) => (
            <Box key={idx} sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
              <TextField
                size="small"
                placeholder="Nom"
                value={v.name}
                onChange={(e) => handleChange(idx, 'name', e.target.value)}
                sx={{ flex: 1 }}
              />
              <TextField
                size="small"
                placeholder="Valeur"
                value={v.value}
                onChange={(e) => handleChange(idx, 'value', e.target.value)}
                sx={{ flex: 1 }}
              />
              <IconButton size="small" color="error" onClick={() => handleRemove(idx)}>
                <DeleteIcon />
              </IconButton>
            </Box>
          ))}
          <Button startIcon={<AddIcon />} onClick={handleAdd} sx={{ alignSelf: 'flex-start' }}>
            Ajouter
          </Button>
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
