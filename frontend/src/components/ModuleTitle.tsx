import { Typography, SxProps, Theme } from '@mui/material'
import { kuraColors } from '../theme'

interface ModuleTitleProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

export default function ModuleTitle({ children, sx }: ModuleTitleProps) {
  return (
    <Typography
      component="h1"
      sx={{
        fontSize: '1.5rem',
        fontWeight: 600,
        letterSpacing: '-0.02em',
        lineHeight: 1.3,
        color: kuraColors.text0,
        mb: 3,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}
