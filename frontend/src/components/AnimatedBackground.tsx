import { Box, keyframes } from '@mui/material'

const gentleFlow = keyframes`
  0%, 100% {
    background-position: 0% 0%, 100% 100%, 50% 50%, 30% 80%;
  }
  50% {
    background-position: 100% 100%, 0% 0%, 50% 50%, 70% 20%;
  }
`

/**
 * Composant de fond animé avec particules/étoiles
 * Utilisé sur toutes les pages pour un style cohérent
 */
export default function AnimatedBackground() {
  return (
    <Box
      sx={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        background: 'linear-gradient(135deg, #0a0e27 0%, #1a1f3a 50%, #2d1b3d 100%)',
        backgroundImage: [
          'radial-gradient(ellipse at 20% 30%, rgba(0, 229, 255, 0.12) 0%, transparent 60%)',
          'radial-gradient(ellipse at 80% 70%, rgba(179, 136, 255, 0.1) 0%, transparent 60%)',
          'radial-gradient(ellipse at 50% 50%, rgba(0, 229, 255, 0.05) 0%, transparent 80%)',
          'radial-gradient(ellipse at 30% 80%, rgba(179, 136, 255, 0.06) 0%, transparent 70%)',
        ].join(', '),
        backgroundSize: '300% 300%, 250% 250%, 400% 400%, 350% 350%',
        animation: `${gentleFlow} 40s ease-in-out infinite`,
        pointerEvents: 'none',
        zIndex: 0,
      }}
    />
  )
}
