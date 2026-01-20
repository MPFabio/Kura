import { ReactNode } from 'react'
import { Card, CardContent, Box, SxProps, Theme } from '@mui/material'

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
          background: 'rgba(10, 10, 14, 0.8)',
          border: '1px solid rgba(191, 0, 255, 0.1)',
          boxShadow: 'none',
          '&:hover': {
            borderColor: 'rgba(191, 0, 255, 0.2)',
            boxShadow: '0 0 10px rgba(191, 0, 255, 0.1)',
          },
        }),
        ...(active && {
          background: 'linear-gradient(135deg, rgba(10, 10, 14, 0.95), rgba(5, 5, 8, 0.98))',
          border: '2px solid rgba(0, 255, 255, 0.6)',
          boxShadow: `
            0 0 40px rgba(0, 255, 255, 0.5),
            inset 0 1px 0 rgba(0, 255, 255, 0.6),
            0 0 80px rgba(0, 255, 255, 0.2)
          `,
          animation: 'glowPulse 2s ease-in-out infinite',
        }),
        ...(deploying && {
          animation: 'glowPulse 2s ease-in-out infinite',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: '-100%',
            width: '100%',
            height: '100%',
            background: 'linear-gradient(90deg, transparent, rgba(0, 255, 255, 0.2), transparent)',
            animation: 'shimmer 2s infinite',
            zIndex: 0,
          },
        }),
        ...sx,
      }}
    >
      {deploying && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `
              radial-gradient(circle at 50% 50%, rgba(0, 255, 255, 0.15) 0%, transparent 70%)
            `,
            animation: 'breathingGlow 3s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          }}
        />
      )}
      {active && !deploying && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `
              radial-gradient(circle at 50% 50%, rgba(0, 255, 255, 0.08) 0%, transparent 70%)
            `,
            pointerEvents: 'none',
            zIndex: 0,
          }}
        />
      )}
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
