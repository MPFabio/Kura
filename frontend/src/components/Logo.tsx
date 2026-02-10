import { Box, Typography, SxProps, Theme } from '@mui/material'
import jellyfishLogo from '../assets/jellyfish_logo.png'

export const kuraWordmarkSx = {
  fontFamily: '"Inter", sans-serif',
  fontWeight: 700,
  letterSpacing: '0.15em',
  color: '#00E5FF',
  textShadow: 'none',
  position: 'relative' as const,
}

interface LogoProps {
  variant?: 'full' | 'icon'
  size?: 'small' | 'medium' | 'large'
  sx?: SxProps<Theme>
}

export default function Logo({ variant = 'full', size = 'medium', sx }: LogoProps) {
  const sizes = {
    small: { jellyfish: 40, text: '1rem', spacing: 0.5 },
    medium: { jellyfish: 60, text: '1.5rem', spacing: 0.75 },
    large: { jellyfish: 100, text: '2.5rem', spacing: 1 },
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
        margin: '0 auto',
        background: 'transparent !important',
        ...sx,
      }}
    >
      {/* Bloc logo : zone fixe pour contenir méduse + orbites et tout centrer */}
      <Box
        sx={{
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: currentSize.jellyfish + 50,
          height: currentSize.jellyfish + 50,
          margin: '0 auto',
          mb: variant === 'full' ? currentSize.spacing : 0,
          flexShrink: 0,
        }}
      >
        {/* Particules qui tournent autour de la méduse */}
        {variant === 'full' && (
          <Box
            sx={{
              position: 'absolute',
              left: '50%',
              top: '50%',
              width: 0,
              height: 0,
              animation: 'constructAnimation 15s linear infinite',
              pointerEvents: 'none',
            }}
          >
            {[...Array(12)].map((_, i) => {
              const angle = (i / 12) * 2 * Math.PI
              const radius = Math.min(25, (currentSize.jellyfish + 50) / 2 - 10)
              const x = Math.cos(angle) * radius
              const y = Math.sin(angle) * radius
              return (
                <Box
                  key={i}
                  sx={{
                    position: 'absolute',
                    width: 3,
                    height: 3,
                    borderRadius: '50%',
                    background: i % 2 === 0 ? 'rgba(0, 229, 255, 0.9)' : 'rgba(236, 64, 122, 0.9)',
                    left: 0,
                    top: 0,
                    transform: `translate(calc(-50% + ${x}px), calc(-50% + ${y}px))`,
                    boxShadow: `0 0 10px ${i % 2 === 0 ? 'rgba(0, 229, 255, 1)' : 'rgba(236, 64, 122, 1)'}`,
                  }}
                />
              )
            })}
          </Box>
        )}

        {/* Cercles orbitaux - cercles complets centrés autour de la méduse (left/right/top/bottom + margin auto) */}
        {variant === 'full' && (
          <>
            <Box
              sx={{
                position: 'absolute',
                left: 0,
                right: 0,
                top: 0,
                bottom: 0,
                margin: 'auto',
                width: currentSize.jellyfish + 20,
                height: currentSize.jellyfish + 20,
                border: '1px solid rgba(0, 229, 255, 0.4)',
                borderRadius: '50%',
                animation: 'constructAnimation 10s linear infinite',
              }}
            />
            <Box
              sx={{
                position: 'absolute',
                left: 0,
                right: 0,
                top: 0,
                bottom: 0,
                margin: 'auto',
                width: currentSize.jellyfish + 28,
                height: currentSize.jellyfish + 28,
                border: '1px solid rgba(179, 136, 255, 0.35)',
                borderRadius: '50%',
                animation: 'constructAnimation 8s linear infinite reverse',
              }}
            />
          </>
        )}

        {/* Logo méduse PNG - centré */}
        <Box
          component="img"
          src={jellyfishLogo}
          alt="KURA Logo"
          sx={{
            width: currentSize.jellyfish,
            height: 'auto',
            maxHeight: currentSize.jellyfish * 1.25,
            objectFit: 'contain',
            objectPosition: 'center',
            position: 'relative',
            zIndex: 1,
            display: 'block',
            margin: '0 auto',
          }}
        />
      </Box>

      {/* Texte "KURA" - sans décalage pour équilibrer l'espace au-dessus / en-dessous */}
      {variant === 'full' && (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            textAlign: 'center',
            width: '100%',
            gap: 0.5,
          }}
        >
          <Typography
            variant="h4"
            sx={{
              ...kuraWordmarkSx,
              fontSize: currentSize.text,
              margin: 0,
              lineHeight: 1.2,
            }}
          >
            KURA
          </Typography>
        </Box>
      )}
    </Box>
  )
}
