import { ReactNode } from 'react'
import { Box, SxProps, Theme } from '@mui/material'
import { kuraColors } from '../theme'

interface ModuleCardProps {
  children: ReactNode
  sx?: SxProps<Theme>
  active?: boolean
  inactive?: boolean
  deploying?: boolean
  onClick?: () => void
}

export default function ModuleCard({
  children,
  sx,
  active = false,
  inactive = false,
  deploying = false,
  onClick,
}: ModuleCardProps) {
  return (
    <Box
      onClick={onClick}
      sx={{
        background: kuraColors.bg2,
        border: `1px solid ${kuraColors.border1}`,
        borderRadius: '8px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
        transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
        overflow: 'hidden',
        cursor: onClick ? 'pointer' : 'default',
        display: 'flex',
        flexDirection: 'column',
        ...(inactive && { opacity: 0.45 }),
        ...(onClick && {
          '&:hover': {
            borderColor: kuraColors.border2,
            boxShadow: '0 4px 12px rgba(0,0,0,0.25)',
          },
        }),
        ...sx,
      }}
    >
      {children}
    </Box>
  )
}
