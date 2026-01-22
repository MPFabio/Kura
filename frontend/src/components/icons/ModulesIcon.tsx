import { Box, SxProps, Theme } from '@mui/material'

interface ModulesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function ModulesIcon({ sx, active = false }: ModulesIconProps) {
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
      {/* Grille de modules */}
      <rect x="3" y="3" width="7" height="7" stroke={color} fill={color} opacity="0.1" />
      <rect x="14" y="3" width="7" height="7" stroke={color} fill={color} opacity="0.1" />
      <rect x="3" y="14" width="7" height="7" stroke={color} fill={color} opacity="0.1" />
      <rect x="14" y="14" width="7" height="7" stroke={color} fill={color} opacity="0.1" />
    </Box>
  )
}
