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
          borderColor: 'rgba(171, 71, 188, 0.25)',
          '&:hover': { borderColor: 'rgba(171, 71, 188, 0.4)', boxShadow: '0 12px 40px rgba(0, 0, 0, 0.5), 0 0 24px rgba(236, 64, 122, 0.06)' },
        }),
        ...(active && {
          borderColor: 'rgba(0, 229, 255, 0.2)',
          '&:hover': { borderColor: 'rgba(0, 229, 255, 0.35)', boxShadow: '0 12px 40px rgba(0, 0, 0, 0.5), 0 0 24px rgba(0, 229, 255, 0.1)' },
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
              radial-gradient(circle at 50% 50%, rgba(0, 229, 255, 0.15) 0%, transparent 70%)
            `,
            animation: 'breathingGlow 3s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          }}
        />
      )}
      {active && !deploying && (
        <>
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: `
                radial-gradient(circle at 30% 30%, rgba(0, 229, 255, 0.12) 0%, transparent 50%),
                radial-gradient(circle at 70% 70%, rgba(179, 136, 255, 0.1) 0%, transparent 50%)
              `,
              pointerEvents: 'none',
              zIndex: 0,
            }}
          />
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'linear-gradient(135deg, transparent, rgba(0, 0, 0, 0.2))',
              pointerEvents: 'none',
              zIndex: 0,
            }}
          />
        </>
      )}
      {inactive && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `
              radial-gradient(circle at 50% 50%, rgba(179, 136, 255, 0.05) 0%, transparent 70%)
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
