import { Link, useNavigate } from 'react-router-dom'
import { Box, Button, Typography, Grid } from '@mui/material'
import AnimatedBackground from '../components/AnimatedBackground'
import Logo from '../components/Logo'
import { kuraColors } from '../theme'

export default function LandingPage() {
  const navigate = useNavigate()

  return (
    <Box sx={{ minHeight: '100vh', position: 'relative', bgcolor: kuraColors.bg0 }}>
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
          height: 72,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 3,
          bgcolor: kuraColors.bg1,
          borderBottom: `1px solid ${kuraColors.border1}`,
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 5 }}>
          <Logo variant="icon" size="small" />
          <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 3 }}>
            {['Kubernetes', 'OpenTofu', 'Semaphore', 'Pipelines'].map((item) => (
              <Link key={item} to="/login" style={{ color: kuraColors.text2, textDecoration: 'none', fontSize: '0.875rem' }}>
                {item}
              </Link>
            ))}
          </Box>
        </Box>
        <Box sx={{ display: 'flex', gap: 1.5 }}>
          <Button variant="outlined" size="small" onClick={() => navigate('/login')}>
            Se connecter
          </Button>
          <Button variant="contained" size="small" onClick={() => navigate('/register')}>
            Créer un compte
          </Button>
        </Box>
      </Box>

      {/* Hero */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          px: { xs: 3, md: 8 },
          pt: '90px',
          pb: 6,
        }}
      >
        <Grid container spacing={8} alignItems="center">
          <Grid item xs={12} md={6}>
            {/* Badge */}
            <Box
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 1,
                px: 1.5,
                py: 0.5,
                borderRadius: '20px',
                border: `1px solid ${kuraColors.border2}`,
                bgcolor: kuraColors.bg2,
                mb: 3,
              }}
            >
              <Box sx={{ width: 6, height: 6, borderRadius: '50%', bgcolor: kuraColors.success }} />
              <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, fontWeight: 500 }}>
                Plateforme DevOps unifiée
              </Typography>
            </Box>

            <Typography
              component="h1"
              sx={{
                fontSize: { xs: '2.25rem', md: '3.25rem' },
                fontWeight: 700,
                color: kuraColors.text0,
                lineHeight: 1.15,
                mb: 2.5,
                letterSpacing: '-0.035em',
              }}
            >
              Un seul outil pour piloter toute votre infrastructure.
            </Typography>

            <Typography sx={{ fontSize: '1.0625rem', color: kuraColors.text1, lineHeight: 1.7, mb: 4, maxWidth: 500 }}>
              Kubernetes, OpenTofu, Semaphore, pipelines CI/CD et monitoring — centralisés dans une interface unique.
            </Typography>

            {/* Terminal */}
            <Box
              sx={{
                p: 2.5,
                bgcolor: '#0A0C11',
                border: `1px solid ${kuraColors.border2}`,
                borderRadius: '8px',
                mb: 4,
                fontFamily: '"JetBrains Mono", monospace',
                fontSize: '0.875rem',
              }}
            >
              <Box sx={{ display: 'flex', gap: 1, mb: 1.5 }}>
                {['#FF5F57', '#FEBC2E', '#28C840'].map((c) => (
                  <Box key={c} sx={{ width: 10, height: 10, borderRadius: '50%', bgcolor: c }} />
                ))}
              </Box>
              <Box sx={{ color: kuraColors.text2 }}>
                <Box component="span" sx={{ color: kuraColors.success }}>$</Box>
                {' '}
                <Box component="span" sx={{ color: kuraColors.text0 }}>kura deploy --env production</Box>
              </Box>
              <Box sx={{ color: kuraColors.text2, mt: 1, fontSize: '0.8125rem' }}>
                <Box component="span" sx={{ color: kuraColors.success }}>✓</Box>
                {' '}Infrastructure synchronisée en 42ms
              </Box>
            </Box>

            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button variant="contained" size="large" onClick={() => navigate('/register')} sx={{ px: 3 }}>
                Créer un compte
              </Button>
              <Button variant="outlined" size="large" onClick={() => navigate('/login')} sx={{ px: 3 }}>
                Se connecter
              </Button>
            </Box>
          </Grid>

          {/* Preview panel */}
          <Grid item xs={12} md={6} sx={{ display: { xs: 'none', md: 'block' } }}>
            <Box
              sx={{
                bgcolor: kuraColors.bg1,
                border: `1px solid ${kuraColors.border1}`,
                borderRadius: '12px',
                overflow: 'hidden',
                boxShadow: '0 25px 60px rgba(0,0,0,0.5)',
              }}
            >
              {/* Window chrome */}
              <Box sx={{ px: 2, py: 1.5, bgcolor: kuraColors.bg2, borderBottom: `1px solid ${kuraColors.border0}`, display: 'flex', gap: 1 }}>
                {['#FF5F57', '#FEBC2E', '#28C840'].map((c) => (
                  <Box key={c} sx={{ width: 10, height: 10, borderRadius: '50%', bgcolor: c }} />
                ))}
              </Box>
              {/* Content */}
              <Box sx={{ p: 3, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2 }}>
                {[
                  { title: 'Services actifs', value: '6', sub: 'Tous opérationnels', color: kuraColors.success },
                  { title: 'Déploiements', value: '14', sub: 'Cette semaine', color: kuraColors.accent },
                  { title: 'Clusters K8s', value: '3', sub: 'GKE / AKS / Local', color: kuraColors.info },
                  { title: 'Alertes', value: '0', sub: 'Aucun incident', color: kuraColors.success },
                ].map((item, i) => (
                  <Box
                    key={i}
                    sx={{
                      p: 2,
                      bgcolor: kuraColors.bg2,
                      border: `1px solid ${kuraColors.border0}`,
                      borderRadius: '8px',
                    }}
                  >
                    <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, mb: 0.5 }}>{item.title}</Typography>
                    <Typography sx={{ fontSize: '1.75rem', fontWeight: 700, color: item.color, fontFamily: '"JetBrains Mono", monospace', lineHeight: 1 }}>
                      {item.value}
                    </Typography>
                    <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, mt: 0.5 }}>{item.sub}</Typography>
                  </Box>
                ))}
              </Box>
              {/* Status bar */}
              <Box sx={{ px: 3, py: 1.5, borderTop: `1px solid ${kuraColors.border0}`, display: 'flex', gap: 3 }}>
                {['auth-service', 'k8s-service', 'terraform-service'].map((svc) => (
                  <Box key={svc} sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                    <Box sx={{ width: 6, height: 6, borderRadius: '50%', bgcolor: kuraColors.success }} />
                    <Typography sx={{ fontSize: '0.7rem', color: kuraColors.text2, fontFamily: '"JetBrains Mono", monospace' }}>
                      {svc}
                    </Typography>
                  </Box>
                ))}
              </Box>
            </Box>
          </Grid>
        </Grid>
      </Box>

      {/* Features strip */}
      <Box
        sx={{
          position: 'relative',
          zIndex: 1,
          px: { xs: 3, md: 8 },
          py: 8,
          borderTop: `1px solid ${kuraColors.border0}`,
        }}
      >
        <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.1em', textAlign: 'center', mb: 5 }}>
          Intégrations natives
        </Typography>
        <Box sx={{ display: 'flex', justifyContent: 'center', flexWrap: 'wrap', gap: 4 }}>
          {['Kubernetes', 'OpenTofu', 'Semaphore', 'GitHub Actions', 'VictoriaMetrics', 'Grafana'].map((tool) => (
            <Typography key={tool} sx={{ fontSize: '0.9375rem', color: kuraColors.text2, fontWeight: 500 }}>
              {tool}
            </Typography>
          ))}
        </Box>
      </Box>
    </Box>
  )
}
