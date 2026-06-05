import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Box,
  Paper,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
} from '@mui/material'
import { useAuth } from '../contexts/AuthContext'
import Logo from '../components/Logo'
import AnimatedBackground from '../components/AnimatedBackground'
import { kuraColors } from '../theme'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      navigate('/projects')
    } catch (err: any) {
      let msg = 'Identifiants incorrects'
      if (err.response) {
        const s = err.response.status
        if (s === 502 || s === 503 || s === 504) {
          msg = 'Service temporairement indisponible. Réessayez dans quelques secondes.'
        } else {
          msg = err.response.data?.error || err.response.data?.message || msg
        }
      } else if (err.request) {
        msg = 'Impossible de joindre le serveur. Vérifiez votre connexion.'
      }
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ minHeight: '100vh', position: 'relative', bgcolor: kuraColors.bg0 }}>
      <AnimatedBackground />

      {/* Header — fond logo pour éviter le carré visible */}
      <Box
        component="header"
        sx={{
          position: 'fixed',
          top: 0, left: 0, right: 0,
          zIndex: 1100,
          height: 72,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 3,
          bgcolor: kuraColors.bg1,
          borderBottom: `1px solid ${kuraColors.border1}`,
        }}
      >
        <Logo variant="icon" size="small" />
        <Button component={Link} to="/" variant="outlined" size="small">
          Accueil
        </Button>
      </Box>

      {/* Main */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          px: 2,
          pt: '88px',
          pb: '32px',
        }}
      >
        <Box sx={{ width: '100%', maxWidth: 400 }}>

          {/* Logo + titre */}
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mb: 1 }}>
            <Logo variant="icon" size="large" sx={{ mt: '-30px', mb: '-40px' }} />
            <Typography sx={{ fontSize: '1.125rem', fontWeight: 600, color: kuraColors.text0 }}>
              Connexion à Kura
            </Typography>
            <Typography sx={{ fontSize: '0.875rem', color: kuraColors.text2, mt: 0.25 }}>
              Gérez votre infrastructure DevOps
            </Typography>
          </Box>

          <Paper
            elevation={0}
            sx={{
              px: 3, pb: 2.5, pt: '36px',
              bgcolor: kuraColors.bg2,
              border: `1px solid ${kuraColors.border1}`,
              borderRadius: '10px',
              overflow: 'visible',
            }}
          >
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            <Box component="form" onSubmit={handleSubmit}>
              <TextField
                required fullWidth
                label="Adresse email"
                type="email"
                autoComplete="email"
                autoFocus
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                InputLabelProps={{ shrink: true }}
                sx={{ mb: 1.5 }}
              />
              <TextField
                required fullWidth
                label="Mot de passe"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                InputLabelProps={{ shrink: true }}
                sx={{ mb: 2 }}
              />
              <Button type="submit" fullWidth variant="contained" disabled={loading} sx={{ py: 1.25 }}>
                {loading ? <CircularProgress size={20} sx={{ color: '#fff' }} /> : 'Se connecter'}
              </Button>

              <Box sx={{ textAlign: 'center', mt: 2, pt: 2, borderTop: `1px solid ${kuraColors.border0}` }}>
                <Link to="/register" style={{ textDecoration: 'none' }}>
                  <Typography sx={{ fontSize: '0.875rem', color: kuraColors.accent, '&:hover': { textDecoration: 'underline' } }}>
                    Pas encore de compte ? S'inscrire
                  </Typography>
                </Link>
              </Box>
            </Box>
          </Paper>
        </Box>
      </Box>
    </Box>
  )
}
