import { Typography, TypographyProps, SxProps, Theme } from '@mui/material'

interface ModuleTextProps extends TypographyProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

// Texte principal (body1) - style standard Terraform
export function ModuleBodyText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="body2"
      {...props}
      sx={{
        fontFamily: '"Inter", sans-serif',
        color: 'rgba(255, 255, 255, 0.95)',
        fontSize: '0.9375rem',
        fontWeight: 500,
        lineHeight: 1.6,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}

// Texte secondaire (body2) - style standard Terraform
export function ModuleSecondaryText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="body2"
      {...props}
      sx={{
        fontFamily: '"Inter", sans-serif',
        color: 'rgba(255, 255, 255, 0.7)',
        fontSize: '0.9375rem',
        fontWeight: 400,
        lineHeight: 1.6,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}

// Texte de sous-titre (h6) - style standard Terraform
export function ModuleSubtitle({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="h6"
      {...props}
      sx={{
        fontFamily: '"Inter", sans-serif',
        color: 'rgba(255, 255, 255, 0.95)',
        fontWeight: 600,
        fontSize: '1.125rem',
        letterSpacing: '0.02em',
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}

// Texte de caption - style standard Terraform
export function ModuleCaption({ children, sx, ...props }: ModuleTextProps) {
  return (
    <Typography
      variant="caption"
      {...props}
      sx={{
        fontFamily: '"Inter", sans-serif',
        color: 'rgba(255, 255, 255, 0.7)',
        fontSize: '0.8125rem',
        fontWeight: 400,
        ...sx,
      }}
    >
      {children}
    </Typography>
  )
}
