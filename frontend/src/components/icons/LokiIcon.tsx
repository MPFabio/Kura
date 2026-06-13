import { Box, SxProps, Theme } from '@mui/material'
import lokiLogo from '../../assets/loki_logo.png'

interface LokiIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function LokiIcon({ sx, active = false }: LokiIconProps) {
  return (
    <Box
      component="img"
      src={lokiLogo}
      alt=""
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        flexShrink: 0,
        objectFit: 'contain',
        opacity: active ? 1 : 0.5,
        transition: 'opacity 0.15s ease',
        ...sx,
      }}
    />
  )
}
