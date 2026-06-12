import { Box, SxProps, Theme } from '@mui/material'
import grafanaLogo from '../../assets/grafana_logo.png'

interface GrafanaIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function GrafanaIcon({ sx, active = false }: GrafanaIconProps) {
  return (
    <Box
      component="img"
      src={grafanaLogo}
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
