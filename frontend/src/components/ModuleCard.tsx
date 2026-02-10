import { ReactNode } from 'react'
import { Card, CardContent, Box, SxProps, Theme } from '@mui/material'
import { jellyfishColors } from '../theme'

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
  onClick 
}: ModuleCardProps) {
  return (
    <Card
      onClick={onClick}
      sx={{
        position: 'relative',
        overflow: 'hidden',
        cursor: onClick ? 'pointer' : 'default',
        minHeight: '300px',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        transition: 'all 0.4s ease',
        ...(inactive && {
          opacity: 0.6,
          '&:hover': { opacity: 0.75 },
        }),
        ...(active && {
          borderLeftColor: jellyfishColors.cyanSoft,
          borderLeftWidth: '4px',
          '&:hover': { borderLeftWidth: '5px' },
        }),
        ...(deploying && {
          borderLeftColor: jellyfishColors.cyanSoft,
          borderLeftWidth: '4px',
        }),
        ...sx,
      }}
    >
      <CardContent 
        sx={{ 
          position: 'relative', 
          zIndex: 1, 
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          height: '100%',
          '&:last-child': { pb: 2 } 
        }}
      >
        {children}
      </CardContent>
    </Card>
  )
}
