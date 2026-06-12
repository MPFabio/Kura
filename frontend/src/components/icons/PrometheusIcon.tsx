import { Box, SxProps, Theme } from '@mui/material'
import prometheusLogo from '../../assets/prometheus_logo.png'

interface PrometheusIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function PrometheusIcon({ sx, active = false }: PrometheusIconProps) {
  return (
    <Box
      component="img"
      src={prometheusLogo}
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
