import { Box, SxProps, Theme } from '@mui/material'

interface SettingsIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function SettingsIcon({ sx, active = false }: SettingsIconProps) {
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
      {/* Engrenage/Paramètres */}
      <circle cx="12" cy="12" r="3" stroke={color} />
      <path
        d="M12 1v6m0 6v6m9-9h-6m-6 0H3m15.364 6.364l-4.243-4.243m-4.242 0l-4.243 4.243m4.242-4.242l-4.243-4.243m4.242 0l4.243 4.243"
        stroke={color}
      />
    </Box>
  )
}
