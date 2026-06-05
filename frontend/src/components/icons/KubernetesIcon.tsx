import { Box, SxProps, Theme } from '@mui/material'
import k8sLogo from '../../assets/k8s_logo.png'

interface KubernetesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

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
        filter: active ? 'none' : 'grayscale(1) opacity(0.4)',
        transition: 'filter 0.2s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
