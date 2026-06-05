import { Box, SxProps, Theme } from '@mui/material'

interface AlertsIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function AlertsIcon({ sx, active = false }: AlertsIconProps) {
  const color = active ? '#F1F3F9' : '#6B7385'

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
        transition: 'stroke 0.15s ease',
        ...sx,
      }}
    >
      {/* Cloche d'alerte */}
      <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" stroke={color} />
      <path d="M13.73 21a2 2 0 0 1-3.46 0" stroke={color} />
    </Box>
  )
}
