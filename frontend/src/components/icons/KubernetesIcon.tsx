import { Box, SxProps, Theme } from '@mui/material'
import k8sLogo from '../../assets/k8s_logo.png'

interface KubernetesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

/**
 * Logo Kubernetes deux tons (cyan + fond) comme avant.
 * Uniforme avec les autres : inactif = gris (filtre), actif = logo cyan + fond.
 */
export default function KubernetesIcon({ sx, active = false }: KubernetesIconProps) {
  return (
    <Box
      component="img"
      src={k8sLogo}
      alt=""
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        objectFit: 'contain',
        filter: active ? 'none' : 'grayscale(1) brightness(0.85) contrast(0.9)',
        opacity: active ? 1 : 0.9,
        transition: 'all 0.3s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
