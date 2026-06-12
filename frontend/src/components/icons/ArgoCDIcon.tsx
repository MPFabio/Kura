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
        filter: active ? 'none' : 'grayscale(1) opacity(0.4)',
        transition: 'filter 0.2s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
