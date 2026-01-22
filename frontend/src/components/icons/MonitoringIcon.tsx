import { Box, SxProps, Theme } from '@mui/material'

interface MonitoringIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function MonitoringIcon({ sx, active = false }: MonitoringIconProps) {
  const color = active ? '#00FFFF' : '#A0A0A0'
  
  return (
    <Box
      component="svg"
      viewBox="0 0 24 24"
      sx={{
        width: 24,
        height: 24,
        fill: 'none',
        stroke: color,
        strokeWidth: 1.5,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        filter: active ? 'drop-shadow(0 0 8px rgba(0, 255, 255, 0.8))' : 'none',
        transition: 'all 0.3s ease',
        ...sx,
      }}
    >
      {/* Graphique style métriques */}
      <polyline points="3,20 7,14 11,16 15,10 19,12 21,8" stroke={color} />
      <line x1="3" y1="20" x2="21" y2="20" stroke={color} opacity="0.3" />
      <line x1="3" y1="12" x2="21" y2="12" stroke={color} opacity="0.2" />
    </Box>
  )
}
