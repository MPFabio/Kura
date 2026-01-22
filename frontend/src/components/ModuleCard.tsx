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
          background: 'linear-gradient(135deg, rgba(10, 14, 20, 0.6), rgba(5, 8, 12, 0.7))',
          backdropFilter: 'blur(30px) saturate(180%)',
          border: 'none',
          boxShadow: `
            0 8px 32px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.05)
          `,
          animation: 'jellyfishFloat 10s ease-in-out infinite',
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          '&:hover': {
            boxShadow: `
              0 12px 48px rgba(179, 136, 255, 0.15),
              0 4px 16px rgba(179, 136, 255, 0.1),
              inset 0 1px 0 rgba(255, 255, 255, 0.08)
            `,
          },
        }),
        ...(active && {
          background: 'linear-gradient(135deg, rgba(10, 14, 20, 0.75), rgba(5, 8, 12, 0.8))',
          backdropFilter: 'blur(30px) saturate(180%)',
          border: 'none',
          boxShadow: `
            0 16px 64px rgba(0, 229, 255, 0.25),
            0 8px 24px rgba(179, 136, 255, 0.2),
            0 4px 12px rgba(0, 0, 0, 0.4),
            inset 0 1px 0 rgba(255, 255, 255, 0.1)
          `,
          animation: 'jellyfishFloat 8s ease-in-out infinite',
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          '&:hover': {
            boxShadow: `
              0 16px 64px rgba(0, 229, 255, 0.25),
              0 8px 24px rgba(179, 136, 255, 0.2),
              0 4px 12px rgba(0, 0, 0, 0.4),
              inset 0 1px 0 rgba(255, 255, 255, 0.12)
            `,
          },
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
              background: 'linear-gradient(135deg, transparent, rgba(5, 8, 12, 0.3))',
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
