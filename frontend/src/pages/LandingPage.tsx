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
          background: '#2c2f3f',
          borderBottom: '2px solid rgba(0, 229, 255, 0.15)',
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
            borderColor: '#00E5FF',
            borderWidth: 2,
            color: '#00E5FF',
            fontWeight: 600,
            '&:hover': {
              background: 'rgba(0, 229, 255, 0.08)',
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
                fontSize: { xs: '2.5rem', md: '4rem' },
                fontWeight: 700,
                color: '#f0f0f0',
                lineHeight: 1.1,
                mb: 5,
                letterSpacing: '-0.04em',
              }}
            >
              L'orchestration aussi fluide qu'une méduse.
            </Typography>
            <Box
              sx={{
                p: 2.5,
                borderLeft: '4px solid #66BB6A',
                background: '#1f2235',
                border: '1px solid rgba(255, 255, 255, 0.08)',
                borderLeftWidth: 4,
                mb: 5,
                fontFamily: '"JetBrains Mono", monospace',
                fontSize: '0.9375rem',
              }}
            >
              <Box component="span" sx={{ color: '#66BB6A', fontWeight: 600 }}>
                $ kura deploy --environment production
              </Box>
              <Box component="div" sx={{ color: '#a0a0a0', mt: 1.5, fontSize: '0.875rem' }}>
                ✓ Infrastructure synchronisée en 42ms.
              </Box>
            </Box>
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              <Button
                variant="contained"
                onClick={() => navigate('/login')}
                sx={{
                  background: '#00E5FF',
                  color: '#0d0e12',
                  fontWeight: 700,
                  px: 4,
                  py: 1.5,
                  fontSize: '1rem',
                  '&:hover': {
                    background: '#26C6DA',
                  },
                }}
              >
                Démarrer le cluster
              </Button>
              <Button
                variant="outlined"
                sx={{
                  borderColor: '#f0f0f0',
                  borderWidth: 2,
                  color: '#f0f0f0',
                  fontWeight: 600,
                  px: 4,
                  py: 1.5,
                  '&:hover': {
                    background: 'rgba(255, 255, 255, 0.08)',
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
                width: 280,
                height: 280,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                border: '4px solid',
                borderColor: '#00E5FF',
                borderLeftColor: '#EC407A',
                borderBottomColor: '#AB47BC',
              }}
            >
              <Box
                component="img"
                src={jellyfishLogo}
                alt="KURA"
                sx={{
                  width: 180,
                  height: 'auto',
                  opacity: 0.95,
                  objectFit: 'contain',
                  objectPosition: 'center',
                }}
              />
            </Box>
          </Grid>
        </Grid>
        <Box
          sx={{
            position: 'absolute',
            bottom: 32,
            left: '50%',
            transform: 'translateX(-50%)',
            color: 'rgba(255, 255, 255, 0.3)',
          }}
        >
          <ChevronUpIcon sx={{ fontSize: 24, transform: 'rotate(180deg)' }} />
        </Box>
      </Box>

      {/* Dashboard Preview */}
      <Box sx={{ position: 'relative', zIndex: 1, px: { xs: 2, md: 6 }, py: 10, pb: 12, borderTop: '1px solid rgba(255, 255, 255, 0.08)' }}>
        <Typography
          component="h2"
          sx={{
            fontSize: '1.125rem',
            fontWeight: 700,
            color: '#808080',
            mb: 6,
            textTransform: 'uppercase',
            letterSpacing: '0.15em',
          }}
        >
          Dashboard Preview
        </Typography>
        <Grid container spacing={3}>
          {/* System Health */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #EC407A' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  System Health
                </Typography>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', gap: 2 }}>
                  {[
                    { value: 70, color: '#EC407A', label: 'CPU' },
                    { value: 90, color: '#00E5FF', label: 'MEM' },
                    { value: 60, color: '#AB47BC', label: 'DISK' }
                  ].map((item, i) => (
                    <Box key={i} sx={{ flex: 1, textAlign: 'center' }}>
                      <Typography sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '1.5rem', fontWeight: 700, color: item.color, mb: 0.5 }}>
                        {item.value}
                      </Typography>
                      <Typography sx={{ fontSize: '0.6875rem', color: '#606060', fontWeight: 700, letterSpacing: '0.1em' }}>
                        {item.label}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
          {/* Deployment Timeline */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #00E5FF' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  Deployment Timeline
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, alignItems: 'flex-end', height: 80 }}>
                  {[40, 70, 50, 90, 60, 80, 45].map((h, i) => (
                    <Box
                      key={i}
                      sx={{
                        flex: 1,
                        height: `${h}%`,
                        background: i % 2 === 0 ? '#00E5FF' : '#EC407A',
                        opacity: 0.8,
                        transition: 'all 0.2s ease',
                        '&:hover': { opacity: 1 },
                      }}
                    />
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
          {/* kura.yaml */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #AB47BC' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  kura.yaml
                </Typography>
                <Box sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8125rem', color: '#b8b8b8', whiteSpace: 'pre-wrap', lineHeight: 1.7 }}>
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
          {/* Second row */}
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #AB47BC' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  Metrics
                </Typography>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                  {['CPU', 'Memory', 'Network'].map((label, i) => (
                    <Box key={i} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography sx={{ fontSize: '0.875rem', color: '#a0a0a0' }}>{label}</Typography>
                      <Box sx={{ width: [60, 80, 50][i], height: 3, background: [60, 80, 50][i] > 70 ? '#EC407A' : '#00E5FF' }} />
                    </Box>
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #00E5FF' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  Logs
                </Typography>
                <Box sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem', color: '#606060', lineHeight: 1.6 }}>
                  <Box sx={{ color: '#66BB6A' }}>[OK] Service started</Box>
                  <Box sx={{ color: '#00E5FF' }}>[INFO] Port 8080</Box>
                  <Box sx={{ color: '#606060' }}>[DEBUG] Connected</Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ height: '100%', borderLeft: '4px solid #EC407A' }}>
              <CardContent sx={{ p: 3 }}>
                <Typography sx={{ color: '#f0f0f0', fontWeight: 700, mb: 3, fontSize: '1rem', letterSpacing: '0.02em' }}>
                  Status
                </Typography>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                  {['API', 'DB', 'Cache'].map((svc, i) => (
                    <Box key={i} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography sx={{ fontSize: '0.875rem', color: '#a0a0a0', fontFamily: '"JetBrains Mono", monospace' }}>{svc}</Typography>
                      <Box sx={{ width: 8, height: 8, background: '#66BB6A' }} />
                    </Box>
                  ))}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Box>
    </Box>
  )
}
