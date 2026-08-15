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
        opacity: active ? 1 : 0.5,
        transition: 'opacity 0.15s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
