import { styled } from '@mui/material/styles'
import { Typography, TypographyProps, SxProps, Theme } from '@mui/material'

interface ModuleTextProps extends TypographyProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

// Composants styled pour forcer les styles - override les styles du theme
const StyledBodyText = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: 'rgba(255, 255, 255, 0.95)',
  fontSize: '0.9375rem',
  fontWeight: 500,
  lineHeight: 1.6,
  '&.MuiTypography-body2': {
    fontFamily: '"Inter", sans-serif !important',
    color: 'rgba(255, 255, 255, 0.95) !important',
    fontSize: '0.9375rem !important',
    fontWeight: '500 !important',
    lineHeight: '1.6 !important',
  },
})

const StyledSecondaryText = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: 'rgba(255, 255, 255, 0.7)',
  fontSize: '0.9375rem',
  fontWeight: 400,
  lineHeight: 1.6,
  '&.MuiTypography-body2': {
    fontFamily: '"Inter", sans-serif !important',
    color: 'rgba(255, 255, 255, 0.7) !important',
    fontSize: '0.9375rem !important',
    fontWeight: '400 !important',
    lineHeight: '1.6 !important',
  },
})

const StyledSubtitle = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: 'rgba(255, 255, 255, 0.95)',
  fontWeight: 600,
  fontSize: '1.125rem',
  letterSpacing: '0.02em',
  '&.MuiTypography-h6': {
    fontFamily: '"Inter", sans-serif !important',
    color: 'rgba(255, 255, 255, 0.95) !important',
    fontWeight: '600 !important',
    fontSize: '1.125rem !important',
    letterSpacing: '0.02em !important',
  },
})

const StyledCaption = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: 'rgba(255, 255, 255, 0.7)',
  fontSize: '0.8125rem',
  fontWeight: 400,
  '&.MuiTypography-caption': {
    fontFamily: '"Inter", sans-serif !important',
    color: 'rgba(255, 255, 255, 0.7) !important',
    fontSize: '0.8125rem !important',
    fontWeight: '400 !important',
  },
})

// Texte principal (body1) - style standard Terraform
export function ModuleBodyText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <StyledBodyText
      variant="body2"
      className="module-body-text"
      {...props}
      sx={sx}
    >
      {children}
    </StyledBodyText>
  )
}

// Texte secondaire (body2) - style standard Terraform
export function ModuleSecondaryText({ children, sx, ...props }: ModuleTextProps) {
  return (
    <StyledSecondaryText
      variant="body2"
      className="module-secondary-text"
      {...props}
      sx={sx}
    >
      {children}
    </StyledSecondaryText>
  )
}

// Texte de sous-titre (h6) - style standard Terraform
export function ModuleSubtitle({ children, sx, ...props }: ModuleTextProps) {
  return (
    <StyledSubtitle
      variant="h6"
      className="module-subtitle"
      {...props}
      sx={sx}
    >
      {children}
    </StyledSubtitle>
  )
}

// Texte de caption - style standard Terraform
export function ModuleCaption({ children, sx, ...props }: ModuleTextProps) {
  return (
    <StyledCaption
      variant="caption"
      className="module-caption"
      {...props}
      sx={sx}
    >
      {children}
    </StyledCaption>
  )
}
