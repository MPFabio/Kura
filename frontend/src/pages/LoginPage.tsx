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
import AnimatedBackground from '../components/AnimatedBackground'

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
      console.error('Erreur de connexion:', err)
      let errorMessage = 'Erreur lors de la connexion'
      
      if (err.response) {
        // Erreur avec réponse du serveur
        const status = err.response.status
        if (status === 502 || status === 503 || status === 504) {
          errorMessage =
            'Service d’authentification temporairement indisponible (passerelle). Réessayez dans quelques secondes ou vérifiez que auth-service et Kong sont démarrés (docker-compose up -d).'
        } else {
          errorMessage = err.response.data?.error || err.response.data?.message || errorMessage
        }
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
    <Box sx={{ minHeight: '100vh', position: 'relative' }}>
      <AnimatedBackground />
      
      {/* Header avec retour à l'accueil */}
      <Box
        component="header"
        sx={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 1100,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'flex-end',
          px: { xs: 2, md: 4 },
          py: 2,
          background: '#2c2f3f',
          borderBottom: '2px solid rgba(0, 229, 255, 0.15)',
        }}
      >
        <Button
          component={Link}
          to="/"
          variant="outlined"
          sx={{
            borderColor: '#f0f0f0',
            borderWidth: 2,
            color: '#f0f0f0',
            fontWeight: 600,
            '&:hover': {
              background: 'rgba(255, 255, 255, 0.08)',
            },
          }}
        >
          Accueil
        </Button>
      </Box>

      {/* Contenu principal centré */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          px: { xs: 2, md: 4 },
          pt: 10,
        }}
      >
        <Container maxWidth="sm">
          <Box
            sx={{
              display: 'flex',
              flexDirection: { xs: 'column', md: 'row' },
              gap: 6,
              alignItems: 'center',
            }}
          >
            {/* Section gauche - Logo et titre */}
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                alignItems: { xs: 'center', md: 'flex-start' },
                textAlign: { xs: 'center', md: 'left' },
              }}
            >
              <Box
                sx={{
                  width: 180,
                  height: 180,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  border: '4px solid',
                  borderColor: '#00E5FF',
                  borderLeftColor: '#EC407A',
                  borderBottomColor: '#AB47BC',
                  mb: 3,
                  overflow: 'hidden',
                }}
              >
                <Logo variant="icon" size="large" />
              </Box>
              <Typography
                component="h1"
                sx={{
                  fontSize: { xs: '2rem', md: '2.5rem' },
                  fontWeight: 700,
                  color: '#f0f0f0',
                  lineHeight: 1.1,
                  letterSpacing: '-0.04em',
                }}
              >
                Bienvenue sur KURA
              </Typography>
            </Box>

            {/* Section droite - Formulaire */}
            <Box
              sx={{
                flex: 1,
                width: '100%',
              }}
            >
              <Paper
                elevation={0}
                sx={{
                  p: 4,
                  background: '#32364a',
                  border: '1px solid rgba(255, 255, 255, 0.08)',
                  borderLeft: '4px solid #00E5FF',
                }}
              >
                <Typography
                  component="h2"
                  sx={{
                    mb: 3,
                    fontSize: '1.5rem',
                    fontWeight: 700,
                    color: '#f0f0f0',
                    letterSpacing: '-0.02em',
                  }}
                >
                  Connexion
                </Typography>

                {error && (
                  <Alert severity="error" sx={{ mb: 3 }}>
                    {error}
                  </Alert>
                )}

                <Box component="form" onSubmit={handleSubmit}>
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
                    sx={{ mb: 2 }}
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
                    sx={{ mb: 3 }}
                  />
                  <Button
                    type="submit"
                    fullWidth
                    variant="contained"
                    disabled={loading}
                    sx={{
                      background: '#00E5FF',
                      color: '#0d0e12',
                      fontWeight: 700,
                      py: 1.5,
                      fontSize: '1rem',
                      '&:hover': {
                        background: '#26C6DA',
                      },
                    }}
                  >
                    {loading ? <CircularProgress size={24} sx={{ color: '#0d0e12' }} /> : 'Se connecter'}
                  </Button>

                  <Box
                    sx={{
                      textAlign: 'center',
                      mt: 3,
                      pt: 3,
                      borderTop: '1px solid rgba(255, 255, 255, 0.08)',
                    }}
                  >
                    <Link to="/register" style={{ textDecoration: 'none' }}>
                      <Typography
                        variant="body2"
                        sx={{
                          color: '#00E5FF',
                          fontWeight: 500,
                          '&:hover': {
                            textDecoration: 'underline',
                          },
                        }}
                      >
                        Pas encore de compte ? S'inscrire
                      </Typography>
                    </Link>
                  </Box>
                </Box>
              </Paper>
            </Box>
          </Box>
        </Container>
      </Box>
    </Box>
  )
}
