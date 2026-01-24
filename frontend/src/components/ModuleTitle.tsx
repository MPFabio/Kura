import { Typography, SxProps, Theme } from '@mui/material'

interface ModuleTitleProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

export default function ModuleTitle({ children, sx }: ModuleTitleProps) {
  return (
    <Typography
      variant="h3"
      component="h1"
      sx={{
        fontWeight: 600,
        background: 'linear-gradient(135deg, #00E5FF, #B388FF)',
        backgroundClip: 'text',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
        fontFamily: '"Inter", sans-serif',
        letterSpacing: '0.02em',
        mb: 5,
        fontSize: '2.5rem',
        lineHeight: 1.2,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}
