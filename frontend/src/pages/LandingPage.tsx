import { Link, useNavigate } from 'react-router-dom'
import {
  Box,
  Button,
  Typography,
  Card,
  CardContent,
  Grid,
} from '@mui/material'
import { KeyboardArrowUp as ChevronUpIcon } from '@mui/icons-material'
import AnimatedBackground from '../components/AnimatedBackground'
import Logo from '../components/Logo'
import jellyfishLogo from '../assets/jellyfish_logo.png'

const kuraYamlSnippet = `version: "1.0"
environment: production
services:
  api:
    build: ./api
    depends_on: [db, cache]
  worker:
    store: redis
    process: celery
`

export default function LandingPage() {
  const navigate = useNavigate()

  return (
    <Box sx={{ minHeight: '100vh', position: 'relative' }}>
      <AnimatedBackground />

      {/* Header */}
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
          justifyContent: 'space-between',
          px: { xs: 2, md: 4 },
          py: 2,
          background: 'rgba(13, 14, 18, 0.7)',
          backdropFilter: 'blur(20px)',
          borderBottom: '1px solid rgba(255, 255, 255, 0.06)',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Logo variant="full" size="small" />
          <Box sx={{ display: 'flex', gap: 3, color: '#b8b8b8', fontSize: '0.9375rem' }}>
            <Link to="/login" style={{ color: 'inherit', textDecoration: 'none' }}>
              ./deploy
            </Link>
            <Link to="/login" style={{ color: 'inherit', textDecoration: 'none' }}>
              //observe
            </Link>
          </Box>
        </Box>
        <Button
          variant="outlined"
          onClick={() => navigate('/login')}
          sx={{
            borderColor: 'rgba(0, 229, 255, 0.6)',
            color: '#f0f0f0',
            '&:hover': {
              borderColor: '#00E5FF',
              background: 'rgba(0, 229, 255, 0.08)',
              boxShadow: '0 0 20px rgba(0, 229, 255, 0.2)',
            },
          }}
        >
          Login
        </Button>
      </Box>

      {/* Hero */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          px: { xs: 2, md: 6 },
          pt: 14,
          pb: 6,
        }}
      >
        <Grid container spacing={6} alignItems="center">
          <Grid item xs={12} md={7}>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: '2rem', md: '3rem' },
                fontWeight: 700,
                color: '#f0f0f0',
                lineHeight: 1.2,
                mb: 4,
                letterSpacing: '-0.02em',
              }}
            >
              L'orchestration aussi fluide qu'une méduse.
            </Typography>
            <Box
              sx={{
                p: 2.5,
                borderRadius: 2,
                background: 'rgba(20, 22, 31, 0.8)',
                border: '1px solid rgba(255, 255, 255, 0.06)',
                mb: 4,
                fontFamily: '"JetBrains Mono", monospace',
                fontSize: '0.875rem',
              }}
            >
              <Box component="span" sx={{ color: '#66BB6A' }}>
                $ kura deploy -environment production
              </Box>
              <Box component="div" sx={{ color: '#b8b8b8', mt: 1 }}>
                ✓ Infrastructure synchronisée en 42ms.
              </Box>
            </Box>
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              <Button
                variant="contained"
                onClick={() => navigate('/login')}
                sx={{
                  background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
                  color: '#fff',
                  fontWeight: 600,
                  px: 3,
                  py: 1.5,
                  borderRadius: 2,
                  boxShadow: '0 4px 24px rgba(0, 229, 255, 0.25)',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #26C6DA, #E91E63)',
                    boxShadow: '0 6px 28px rgba(0, 229, 255, 0.35)',
                  },
                }}
              >
                Démarrer le cluster
              </Button>
              <Button
                variant="outlined"
                sx={{
                  borderColor: 'rgba(255, 255, 255, 0.2)',
                  color: '#f0f0f0',
                  '&:hover': {
                    borderColor: 'rgba(255, 255, 255, 0.35)',
                    background: 'rgba(255, 255, 255, 0.04)',
                  },
                }}
              >
                Documentation
              </Button>
            </Box>
          </Grid>
          <Grid item xs={12} md={5} sx={{ display: { xs: 'none', md: 'flex' }, justifyContent: 'center' }}>
            <Box
              sx={{
                width: 320,
                height: 320,
                background: 'radial-gradient(circle at 50% 30%, rgba(0, 229, 255, 0.15) 0%, transparent 50%), radial-gradient(circle at 50% 80%, rgba(236, 64, 122, 0.12) 0%, transparent 45%)',
                borderRadius: '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                border: '1px solid rgba(0, 229, 255, 0.2)',
                boxShadow: '0 0 80px rgba(0, 229, 255, 0.15), 0 0 120px rgba(236, 64, 122, 0.1)',
              }}
            >
              <Box
                component="img"
                src={jellyfishLogo}
                alt="KURA"
                sx={{
                  width: 200,
                  height: 'auto',
                  filter: 'drop-shadow(0 0 30px rgba(0, 229, 255, 0.5)) drop-shadow(0 0 50px rgba(236, 64, 122, 0.3))',
                }}
              />
            </Box>
          </Grid>
        </Grid>
        <Box
          sx={{
            position: 'absolute',
            bottom: 24,
            left: '50%',
            transform: 'translateX(-50%)',
            color: 'rgba(255, 255, 255, 0.5)',
            animation: 'breathingGlow 2s ease-in-out infinite',
          }}
        >
          <ChevronUpIcon sx={{ fontSize: 32, transform: 'rotate(180deg)' }} />
        </Box>
      </Box>

      {/* Dashboard Preview */}
      <Box sx={{ position: 'relative', zIndex: 1, px: { xs: 2, md: 6 }, py: 8, pb: 10 }}>
        <Typography
          component="h2"
          sx={{
            fontSize: '1.5rem',
            fontWeight: 700,
            color: '#f0f0f0',
            mb: 4,
          }}
        >
          Dashboard Preview
        </Typography>
        <Grid container spacing={3}>
          {/* System Health */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  System Health
                </Typography>
                <Box sx={{ display: 'flex', justifyContent: 'space-around', gap: 1 }}>
                  {[0.7, 0.9, 0.6].map((value, i) => (
                    <Box
                      key={i}
                      sx={{
                        width: 64,
                        height: 64,
                        borderRadius: '50%',
                        border: '3px solid',
                        borderColor: i === 0 ? 'rgba(236, 64, 122, 0.6)' : i === 1 ? 'rgba(0, 229, 255, 0.6)' : 'rgba(171, 71, 188, 0.6)',
                        background: `conic-gradient(${i === 0 ? '#EC407A' : i === 1 ? '#00E5FF' : '#AB47BC'} ${value * 360}deg, rgba(255,255,255,0.06) 0deg)`,
                        boxShadow: `0 0 16px ${i === 0 ? 'rgba(236, 64, 122, 0.3)' : i === 1 ? 'rgba(0, 229, 255, 0.3)' : 'rgba(171, 71, 188, 0.3)'}`,
                      }}
                    />
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
          {/* Deployment Timeline */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  Deployment Timeline
                </Typography>
                <Box
                  sx={{
                    height: 120,
                    background: 'linear-gradient(180deg, transparent 0%, rgba(0, 229, 255, 0.05) 50%, transparent 100%)',
                    borderRadius: 1,
                    border: '1px solid rgba(0, 229, 255, 0.15)',
                    position: 'relative',
                    overflow: 'hidden',
                  }}
                >
                  <Box
                    sx={{
                      position: 'absolute',
                      bottom: 0,
                      left: 0,
                      right: 0,
                      height: '60%',
                      background: 'linear-gradient(90deg, transparent 0%, rgba(0, 229, 255, 0.4) 20%, rgba(0, 229, 255, 0.6) 50%, rgba(236, 64, 122, 0.4) 80%, transparent 100%)',
                      clipPath: 'polygon(0% 100%, 10% 60%, 25% 80%, 40% 40%, 55% 70%, 70% 30%, 85% 50%, 100% 20%, 100% 100%)',
                    }}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          {/* kura.yaml */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  kura.yaml
                </Typography>
                <Box sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem', color: '#b8b8b8', whiteSpace: 'pre-wrap' }}>
                  {kuraYamlSnippet.split('\n').map((line, i) => (
                    <Box key={i}>
                      {line.split(/(version|environment|services|api|worker|build|depends_on|store|process|redis|celery|production|"1.0"|\/\w+)/g).map((part, j) => {
                        if (['version', 'environment', 'services', 'api', 'worker', 'build', 'depends_on', 'store', 'process'].includes(part)) return <span key={j} style={{ color: '#00E5FF' }}>{part}</span>
                        if (part?.match(/^["'].*["']$/)) return <span key={j} style={{ color: '#66BB6A' }}>{part}</span>
                        if (part === 'production' || part === 'redis' || part === 'celery') return <span key={j} style={{ color: '#EC407A' }}>{part}</span>
                        return <span key={j}>{part}</span>
                      })}
                      {'\n'}
                    </Box>
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
          {/* Second row: repeat pattern */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  Deployment Timeline
                </Typography>
                <Box sx={{ height: 120, background: 'rgba(0,0,0,0.2)', borderRadius: 1, border: '1px solid rgba(255,255,255,0.06)' }} />
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  kura.yaml
                </Typography>
                <Box sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.7rem', color: '#b8b8b8' }}>
                  <span style={{ color: '#66BB6A' }}>username</span>: encode
                  <br />
                  <span style={{ color: '#66BB6A' }}>test</span>: <span style={{ color: '#EC407A' }}>build</span>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 600, mb: 2, fontSize: '1rem' }}>
                  kura.yaml
                </Typography>
                <Box sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.7rem', color: '#b8b8b8' }}>
                  <span style={{ color: '#00E5FF' }}>environment</span>: store
                  <br />
                  <span style={{ color: '#00E5FF' }}>process</span>: <span style={{ color: '#EC407A' }}>depends_on</span>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Box>
    </Box>
  )
}
