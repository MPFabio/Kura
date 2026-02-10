import { styled } from '@mui/material/styles'
import { Typography, TypographyProps, SxProps, Theme } from '@mui/material'

interface ModuleTextProps extends TypographyProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

// Texte très lisible sur fond sombre - contraste inversé (clair sur sombre)
const textPrimary = '#f0f0f0'
const textSecondary = '#b8b8b8'

const StyledBodyText = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: textPrimary,
  fontSize: '1rem',
  fontWeight: 500,
  lineHeight: 1.6,
  '&.MuiTypography-body2': {
    fontFamily: '"Inter", sans-serif !important',
    color: `${textPrimary} !important`,
    fontSize: '1rem !important',
    fontWeight: '500 !important',
    lineHeight: '1.6 !important',
  },
})

const StyledSecondaryText = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: textSecondary,
  fontSize: '1rem',
  fontWeight: 500,
  lineHeight: 1.6,
  '&.MuiTypography-body2': {
    fontFamily: '"Inter", sans-serif !important',
    color: `${textSecondary} !important`,
    fontSize: '1rem !important',
    fontWeight: '500 !important',
    lineHeight: '1.6 !important',
  },
})

const StyledSubtitle = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: textPrimary,
  fontWeight: 600,
  fontSize: '1.125rem',
  letterSpacing: '0.02em',
  '&.MuiTypography-h6': {
    fontFamily: '"Inter", sans-serif !important',
    color: `${textPrimary} !important`,
    fontWeight: '600 !important',
    fontSize: '1.125rem !important',
    letterSpacing: '0.02em !important',
  },
})

const StyledCaption = styled(Typography)({
  fontFamily: '"Inter", sans-serif',
  color: textSecondary,
  fontSize: '0.875rem',
  fontWeight: 500,
  '&.MuiTypography-caption': {
    fontFamily: '"Inter", sans-serif !important',
    color: `${textSecondary} !important`,
    fontSize: '0.875rem !important',
    fontWeight: '500 !important',
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
