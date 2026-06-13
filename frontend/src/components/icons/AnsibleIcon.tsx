import { Box, SxProps, Theme } from '@mui/material'
import ansibleLogo from '../../assets/semaphore_logo.png'

interface AnsibleIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function AnsibleIcon({ sx, active = false }: AnsibleIconProps) {
  const size = (sx as any)?.width ?? (sx as any)?.fontSize ?? 24

  if (!active) {
    return (
      <Box
        aria-hidden
        sx={{
          width: size,
          height: size,
          backgroundColor: '#6B7385',
          mask: `url(${ansibleLogo}) center/contain no-repeat`,
          WebkitMask: `url(${ansibleLogo}) center/contain no-repeat`,
          flexShrink: 0,
          opacity: 0.5,
          ...sx,
        }}
      />
    )
  }

  // Actif : cercle noir + points blancs (double couche)
  // Couche 1 (derrière) : fond blanc visible à travers les points transparents du stencil
  // Couche 2 (devant) : masque noir = le cercle
  return (
    <Box sx={{ position: 'relative', width: size, height: size, flexShrink: 0, ...sx }}>
      {/* Fond blanc — visible seulement à travers les points (zone transparente du stencil) */}
      <Box sx={{ position: 'absolute', inset: 0, bgcolor: 'white', borderRadius: '50%' }} />
      {/* Masque noir au-dessus = cercle noir, points = transparents → fond blanc visible */}
      <Box
        aria-hidden
        sx={{
          position: 'absolute',
          inset: 0,
          backgroundColor: '#1A1A1A',
          mask: `url(${ansibleLogo}) center/contain no-repeat`,
          WebkitMask: `url(${ansibleLogo}) center/contain no-repeat`,
        }}
      />
    </Box>
  )
}
