import { Box, SxProps, Theme } from '@mui/material'
import argoLogo from '../../assets/argo_logo.png'

interface ArgoCDIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function ArgoCDIcon({ sx, active = false }: ArgoCDIconProps) {
  return (
    <Box
      component="img"
      src={argoLogo}
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
