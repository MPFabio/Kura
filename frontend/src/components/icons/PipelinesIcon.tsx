import { AllInclusive } from '@mui/icons-material'
import { SxProps, Theme } from '@mui/material'

interface PipelinesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function PipelinesIcon({ sx, active = false }: PipelinesIconProps) {
  return (
    <AllInclusive
      sx={{
        width: 24,
        height: 24,
        color: active ? '#4F8EF7' : '#6B7385',
        transition: 'color 0.15s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
