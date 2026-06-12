import { Box, SxProps, Theme } from '@mui/material'
import prometheusLogo from '../../assets/prometheus_logo.png'
import grafanaLogo from '../../assets/grafana_logo.png'

interface ObservabilityIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function ObservabilityIcon({ sx, active = false }: ObservabilityIconProps) {
  return (
    <Box
      sx={{
        position: 'relative',
        width: 80,
        height: 80,
        flexShrink: 0,
        opacity: active ? 1 : 0.5,
        filter: active ? 'none' : 'grayscale(1)',
        transition: 'opacity 0.15s ease, filter 0.15s ease',
        ...sx,
      }}
    >
      <Box
        component="img"
        src={grafanaLogo}
        alt=""
        aria-hidden
        sx={{ position: 'absolute', top: 0, left: 0, width: '62%', height: '62%', objectFit: 'contain' }}
      />
      <Box
        component="img"
        src={prometheusLogo}
        alt=""
        aria-hidden
        sx={{ position: 'absolute', bottom: 0, right: 0, width: '50%', height: '50%', objectFit: 'contain' }}
      />
    </Box>
  )
}
