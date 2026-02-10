import { useState } from 'react'
import {
  Box,
  TextField,
  Grid,
  Alert,
  CircularProgress,
} from '@mui/material'
import { useAuth } from '../contexts/AuthContext'
import { authService } from '../services/authService'
import ModuleTitle from '../components/ModuleTitle'
import ModuleButton from '../components/ModuleButton'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle } from '../components/ModuleText'

export default function SettingsPage() {
  const { user, refreshUser } = useAuth()
  const [name, setName] = useState(user?.name || '')
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleUpdateProfile = async () => {
    setLoading(true)
    setMessage(null)

    try {
      await authService.updateUser({ name })
      await refreshUser()
      setMessage({ type: 'success', text: 'Profil mis à jour avec succès' })
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Erreur lors de la mise à jour' })
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async () => {
    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: 'Les mots de passe ne correspondent pas' })
      return
    }

    if (newPassword.length < 6) {
      setMessage({ type: 'error', text: 'Le mot de passe doit contenir au moins 6 caractères' })
      return
    }

    setLoading(true)
    setMessage(null)

    try {
      await authService.changePassword(currentPassword, newPassword)
      setMessage({ type: 'success', text: 'Mot de passe modifié avec succès' })
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Erreur lors du changement de mot de passe' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box>
      <ModuleTitle>Paramètres</ModuleTitle>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <ModuleCard>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Profil
            </ModuleSubtitle>
              {message && (
                <Alert severity={message.type} sx={{ mb: 2 }}>
                  {message.text}
                </Alert>
              )}
              <TextField
                fullWidth
                label="Email"
                value={user?.email || ''}
                disabled
                sx={{ mb: 2 }}
              />
              <TextField
                fullWidth
                label="Nom"
                value={name}
                onChange={(e) => setName(e.target.value)}
                sx={{ mb: 2 }}
              />
              <ModuleButton
                onClick={handleUpdateProfile}
                disabled={loading}
              >
                {loading ? <CircularProgress size={24} /> : 'Mettre à jour le profil'}
              </ModuleButton>
          </ModuleCard>
        </Grid>

        <Grid item xs={12} md={6}>
          <ModuleCard>
            <ModuleSubtitle sx={{ mb: 3 }}>
              Changer le mot de passe
            </ModuleSubtitle>
            <TextField
              fullWidth
              label="Mot de passe actuel"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Nouveau mot de passe"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Confirmer le nouveau mot de passe"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              sx={{ mb: 2 }}
            />
            <ModuleButton
              onClick={handleChangePassword}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : 'Changer le mot de passe'}
            </ModuleButton>
          </ModuleCard>
        </Grid>
      </Grid>
    </Box>
  )
}
