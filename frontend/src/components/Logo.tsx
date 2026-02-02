import { Box, Typography, SxProps, Theme } from '@mui/material'
import jellyfishLogo from '../assets/jellyfish_logo.png'

export const kuraWordmarkSx = {
  fontFamily: '"Inter", sans-serif',
  fontWeight: 600,
  letterSpacing: '0.15em',
  background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
  backgroundClip: 'text',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  textShadow: 'none',
  position: 'relative' as const,
  '&::before': {
    content: '""',
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'linear-gradient(135deg, rgba(0, 229, 255, 0.4), rgba(236, 64, 122, 0.4))',
    filter: 'blur(12px)',
    zIndex: -1,
    borderRadius: '4px',
  },
}

interface LogoProps {
  variant?: 'full' | 'icon'
  size?: 'small' | 'medium' | 'large'
  sx?: SxProps<Theme>
}

export default function Logo({ variant = 'full', size = 'medium', sx }: LogoProps) {
  const sizes = {
    small: { jellyfish: 40, text: '1rem', spacing: 1 },
    medium: { jellyfish: 60, text: '1.5rem', spacing: 1.5 },
    large: { jellyfish: 100, text: '2.5rem', spacing: 2 },
  }

  const currentSize = sizes[size]

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        minWidth: 'fit-content',
        ...sx,
      }}
    >
      {/* Particules en arrière-plan */}
      {variant === 'full' && (
        <>
          {[...Array(12)].map((_, i) => (
            <Box
              key={i}
              sx={{
                position: 'absolute',
                width: 2,
                height: 2,
                borderRadius: '50%',
                background: i % 2 === 0 ? 'rgba(0, 229, 255, 0.7)' : 'rgba(236, 64, 122, 0.7)',
                left: `${Math.random() * 100}%`,
                top: `${Math.random() * 60}%`,
                animation: 'particleFloat 8s ease-in-out infinite',
                animationDelay: `${i * 0.4}s`,
                boxShadow: `0 0 8px ${i % 2 === 0 ? 'rgba(0, 229, 255, 0.9)' : 'rgba(236, 64, 122, 0.9)'}`,
              }}
            />
          ))}
        </>
      )}

      {/* Cercles orbitaux autour de la méduse */}
      <Box
        sx={{
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          mb: variant === 'full' ? currentSize.spacing : 0,
        }}
      >
        {/* Cercles orbitaux */}
        {variant === 'full' && (
          <>
            <Box
              sx={{
                position: 'absolute',
                width: currentSize.jellyfish + 30,
                height: currentSize.jellyfish + 30,
                border: '1px solid rgba(0, 229, 255, 0.4)',
                borderRadius: '50%',
                animation: 'constructAnimation 10s linear infinite',
              }}
            />
            <Box
              sx={{
                position: 'absolute',
                width: currentSize.jellyfish + 40,
                height: currentSize.jellyfish + 40,
                border: '1px solid transparent',
                borderTop: '1px solid rgba(0, 229, 255, 0.5)',
                borderRight: '1px solid rgba(179, 136, 255, 0.4)',
                borderRadius: '50%',
                animation: 'constructAnimation 8s linear infinite reverse',
              }}
            />
          </>
        )}

        {/* Logo méduse PNG avec effets de lueur et flottement naturel */}
        <Box
          component="img"
          src={jellyfishLogo}
          alt="KURA Logo"
          sx={{
            width: currentSize.jellyfish,
            height: 'auto',
            maxHeight: currentSize.jellyfish * 1.25,
            objectFit: 'contain',
            filter: 'drop-shadow(0 0 20px rgba(0, 229, 255, 0.9)) drop-shadow(0 0 40px rgba(179, 136, 255, 0.6))',
            animation: 'jellyfishFloat 8s ease-in-out infinite, breathingGlow 4s ease-in-out infinite',
            position: 'relative',
            zIndex: 1,
          }}
        />
      </Box>

      {/* Texte "KURA" et "DevOps Treasury" */}
      {variant === 'full' && (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 0.5,
          }}
        >
          <Typography
            variant="h4"
            sx={{
              ...kuraWordmarkSx,
              fontSize: currentSize.text,
            }}
          >
            KURA
          </Typography>
        </Box>
      )}
    </Box>
  )
}
