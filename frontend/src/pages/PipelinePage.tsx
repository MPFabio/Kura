import { Box, Typography, Alert } from '@mui/material'

export default function PipelinePage() {
  return (
    <Box>
      <Typography 
        variant="h3" 
        component="h1"
        sx={{ 
          mb: 5,
          fontWeight: 600,
          background: 'linear-gradient(135deg, #00E5FF, #B388FF)',
          backgroundClip: 'text',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          fontFamily: '"Inter", sans-serif',
          letterSpacing: '0.02em',
        }}
      >
        Pipelines CI/CD
      </Typography>

      <Alert 
        severity="info"
        sx={{
          borderRadius: 3,
          maxWidth: '900px',
          animation: 'jellyfishFloat 14s ease-in-out infinite',
        }}
      >
        Le service Pipelines sera bientôt disponible. Cette page affichera les pipelines
        GitHub Actions, GitLab CI et Jenkins avec leurs statuts et historiques.
      </Alert>
    </Box>
  )
}
