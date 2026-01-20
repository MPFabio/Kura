import { Box, Typography } from '@mui/material'
import { Storage as StorageIcon } from '@mui/icons-material'
import ModuleCard from '../components/ModuleCard'

export default function TerraformPage() {
  return (
    <Box
      sx={{
        minHeight: '60vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
      }}
    >
      {/* Particules autour de la carte */}
      <Box
        sx={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          pointerEvents: 'none',
          overflow: 'hidden',
        }}
      >
        {[...Array(20)].map((_, i) => (
          <Box
            key={i}
            sx={{
              position: 'absolute',
              width: 2,
              height: 2,
              borderRadius: '50%',
              background: i % 2 === 0 ? 'rgba(0, 255, 255, 0.6)' : 'rgba(191, 0, 255, 0.6)',
              left: `${Math.random() * 100}%`,
              top: `${Math.random() * 100}%`,
              animation: 'particleFloat 8s ease-in-out infinite',
              animationDelay: `${i * 0.3}s`,
              boxShadow: `0 0 10px ${i % 2 === 0 ? 'rgba(0, 255, 255, 0.8)' : 'rgba(191, 0, 255, 0.8)'}`,
            }}
          />
        ))}
      </Box>

      <ModuleCard
        deploying={true}
        sx={{
          maxWidth: 600,
          width: '100%',
          textAlign: 'center',
          py: 6,
          px: 4,
        }}
      >
        {/* Icône Terraform avec effet de lueur et animation */}
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            mb: 4,
            position: 'relative',
          }}
        >
          <Box
            sx={{
              position: 'relative',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {/* Cercles concentriques animés */}
            <Box
              sx={{
                position: 'absolute',
                width: 120,
                height: 120,
                border: '2px solid rgba(0, 255, 255, 0.3)',
                borderRadius: '50%',
                animation: 'constructAnimation 4s linear infinite',
              }}
            />
            <Box
              sx={{
                position: 'absolute',
                width: 140,
                height: 140,
                border: '1px solid rgba(191, 0, 255, 0.2)',
                borderRadius: '50%',
                animation: 'constructAnimation 6s linear infinite reverse',
              }}
            />
            {/* Lignes courbes autour de l'icône */}
            <Box
              sx={{
                position: 'absolute',
                width: 160,
                height: 160,
                border: '1px solid transparent',
                borderTop: '1px solid rgba(0, 255, 255, 0.4)',
                borderRight: '1px solid rgba(191, 0, 255, 0.4)',
                borderRadius: '50%',
                animation: 'constructAnimation 8s linear infinite',
              }}
            />
            {/* Icône principale */}
            <StorageIcon
              sx={{
                fontSize: 80,
                color: '#00FFFF',
                filter: 'drop-shadow(0 0 20px rgba(0, 255, 255, 0.8)) drop-shadow(0 0 40px rgba(0, 255, 255, 0.5))',
                animation: 'breathingGlow 3s ease-in-out infinite',
                position: 'relative',
                zIndex: 1,
              }}
            />
          </Box>
        </Box>

        {/* Message "Bientôt disponible" */}
        <Typography
          variant="h5"
          sx={{
            fontFamily: '"JetBrains Mono", monospace',
            color: '#A0A0A0',
            mb: 2,
            textShadow: '0 0 10px rgba(160, 160, 160, 0.5)',
            letterSpacing: '0.05em',
          }}
        >
          Bientôt disponible
        </Typography>

        <Typography
          variant="body2"
          sx={{
            fontFamily: '"JetBrains Mono", monospace',
            color: '#808080',
            maxWidth: 500,
            mx: 'auto',
            lineHeight: 1.8,
          }}
        >
          Le service Terraform sera bientôt disponible. Cette page affichera l'état des ressources
          Terraform, la détection de drift et les visualisations.
        </Typography>

        {/* Petits points lumineux autour du texte */}
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            gap: 1,
            mt: 3,
          }}
        >
          {[...Array(3)].map((_, i) => (
            <Box
              key={i}
              sx={{
                width: 6,
                height: 6,
                borderRadius: '50%',
                background: i === 0 ? '#00FFFF' : i === 1 ? '#BF00FF' : '#FF00BF',
                boxShadow: `0 0 10px ${i === 0 ? '#00FFFF' : i === 1 ? '#BF00FF' : '#FF00BF'}`,
                animation: 'breathingGlow 2s ease-in-out infinite',
                animationDelay: `${i * 0.3}s`,
              }}
            />
          ))}
        </Box>
      </ModuleCard>

      {/* Injection des animations CSS */}
      <style>{`
        @keyframes particleFloat {
          0%, 100% {
            transform: translateY(0) translateX(0);
            opacity: 0;
          }
          50% {
            transform: translateY(-20px) translateX(10px);
            opacity: 1;
          }
        }
      `}</style>
    </Box>
  )
}
