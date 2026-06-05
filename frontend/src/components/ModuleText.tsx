import { Typography, TypographyProps, SxProps, Theme } from '@mui/material'
import { kuraColors } from '../theme'

interface ModuleTextProps extends TypographyProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

export function ModuleBodyText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="body1"
      {...props}
      sx={{ color: kuraColors.text0, lineHeight: 1.65, ...sx }}
    >
      {children}
    </Typography>
  )
}

export function ModuleSecondaryText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="body2"
      {...props}
      sx={{ color: kuraColors.text1, lineHeight: 1.6, ...sx }}
    >
      {children}
    </Typography>
  )
}

export function ModuleSubtitle({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="h6"
      {...props}
      sx={{
        fontSize: '0.9375rem',
        fontWeight: 600,
        color: kuraColors.text0,
        letterSpacing: 0,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}

export function ModuleCaption({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="caption"
      {...props}
      sx={{ color: kuraColors.text2, fontSize: '0.75rem', ...sx }}
    >
      {children}
    </Typography>
  )
}
