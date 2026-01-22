import { Box, SxProps, Theme } from '@mui/material'

interface PipelinesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function PipelinesIcon({ sx, active = false }: PipelinesIconProps) {
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
      {/* Pipeline/workflow style */}
      <circle cx="6" cy="6" r="2" fill={color} opacity={active ? 0.8 : 0.5} />
      <circle cx="12" cy="12" r="2" fill={color} opacity={active ? 0.8 : 0.5} />
      <circle cx="18" cy="18" r="2" fill={color} opacity={active ? 0.8 : 0.5} />
      <path d="M8 6L10 12M14 12L16 18" stroke={color} strokeWidth="1.5" />
    </Box>
  )
}
