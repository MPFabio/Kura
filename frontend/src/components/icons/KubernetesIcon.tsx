import { Box, SxProps, Theme } from '@mui/material'

interface KubernetesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function KubernetesIcon({ sx, active = false }: KubernetesIconProps) {
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
      {/* Kubernetes "steering wheel" logo */}
      <circle cx="12" cy="12" r="8" opacity="0.2" />
      {/* Central hub */}
      <circle cx="12" cy="12" r="2.5" fill={color} opacity="0.8" />
      {/* 7 spokes */}
      {[0, 51.4, 102.9, 154.3, 205.7, 257.1, 308.6].map((angle, i) => {
        const rad = (angle * Math.PI) / 180
        const x1 = 12 + 2.5 * Math.cos(rad)
        const y1 = 12 + 2.5 * Math.sin(rad)
        const x2 = 12 + 8 * Math.cos(rad)
        const y2 = 12 + 8 * Math.sin(rad)
        return (
          <line
            key={i}
            x1={x1}
            y1={y1}
            x2={x2}
            y2={y2}
            stroke={color}
            strokeWidth="1.5"
          />
        )
      })}
    </Box>
  )
}
