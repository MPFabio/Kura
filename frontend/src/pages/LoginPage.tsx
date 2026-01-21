import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Container,
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
      navigate('/')
    } catch (err: any) {
      console.error('Erreur de connexion:', err)
      let errorMessage = 'Erreur lors de la connexion'
      
      if (err.response) {
        // Erreur avec réponse du serveur
        errorMessage = err.response.data?.error || err.response.data?.message || errorMessage
      } else if (err.request) {
        // Pas de réponse du serveur (service non démarré, problème réseau)
        // Vérifier si c'est un problème de connexion réseau
        if (err.code === 'ECONNREFUSED' || err.code === 'ERR_NETWORK') {
          errorMessage = 'Impossible de se connecter au serveur. Vérifiez que tous les services sont démarrés (docker-compose up -d).'
        } else if (err.code === 'ETIMEDOUT') {
          errorMessage = 'Le serveur met trop de temps à répondre. Vérifiez que Kong et auth-service sont démarrés.'
        } else {
          errorMessage = `Erreur réseau: ${err.message || 'Impossible de contacter le serveur'}. Vérifiez que les services sont démarrés.`
        }
      } else {
        // Erreur lors de la configuration de la requête
        errorMessage = err.message || errorMessage
      }
      
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Container component="main" maxWidth="xs">
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <Logo variant="full" size="medium" sx={{ mb: 2 }} />
        <Typography component="h2" variant="h6" sx={{ mb: 4, color: 'text.secondary' }}>
          Connexion
        </Typography>
        <Paper elevation={3} sx={{ p: 4, width: '100%' }}>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
            <TextField
              margin="normal"
              required
              fullWidth
              id="email"
              label="Adresse email"
              name="email"
              autoComplete="email"
              autoFocus
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="password"
              label="Mot de passe"
              type="password"
              id="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : 'Se connecter'}
            </Button>
            <Box sx={{ textAlign: 'center', mt: 2 }}>
              <Link to="/register" style={{ textDecoration: 'none', color: 'inherit' }}>
                <Typography variant="body2" color="primary">
                  Pas encore de compte ? S'inscrire
                </Typography>
              </Link>
            </Box>
          </Box>
        </Paper>
      </Box>
    </Container>
  )
}
