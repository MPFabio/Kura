import { styled } from '@mui/material/styles'
import { Typography, SxProps, Theme } from '@mui/material'

interface ModuleTitleProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

/** Style commun pour tous les en-têtes de page (Modules, Terraform, Kubernetes, Ansible, etc.) */
const TITLE_COLOR = '#f0f0f0'
const TITLE_FONT = '"Inter", -apple-system, BlinkMacSystemFont, sans-serif'
const TITLE_FONT_SIZE = '2.5rem'
const TITLE_FONT_WEIGHT = 700
const TITLE_LETTER_SPACING = '-0.02em'
const TITLE_LINE_HEIGHT = 1.2

const StyledModuleTitle = styled(Typography, {
  shouldForwardProp: (prop) => prop !== 'sx',
})({
  fontWeight: `${TITLE_FONT_WEIGHT} !important`,
  color: `${TITLE_COLOR} !important`,
  fontFamily: `${TITLE_FONT} !important`,
  letterSpacing: `${TITLE_LETTER_SPACING} !important`,
  fontSize: `${TITLE_FONT_SIZE} !important`,
  lineHeight: `${TITLE_LINE_HEIGHT} !important`,
  '&.MuiTypography-root': {
    fontWeight: `${TITLE_FONT_WEIGHT} !important`,
    fontSize: `${TITLE_FONT_SIZE} !important`,
    letterSpacing: `${TITLE_LETTER_SPACING} !important`,
    fontFamily: `${TITLE_FONT} !important`,
    color: `${TITLE_COLOR} !important`,
  },
}) as typeof Typography

export default function ModuleTitle({ children, sx }: ModuleTitleProps) {
  return (
    <StyledModuleTitle
      variant="h3"
      component="h1"
      className="module-title"
      sx={{
        mb: 5,
        ...sx,
        fontWeight: `${TITLE_FONT_WEIGHT} !important`,
        fontSize: `${TITLE_FONT_SIZE} !important`,
        fontFamily: `${TITLE_FONT} !important`,
        letterSpacing: `${TITLE_LETTER_SPACING} !important`,
        lineHeight: `${TITLE_LINE_HEIGHT} !important`,
        color: `${TITLE_COLOR} !important`,
      }}
    >
      {children}
    </StyledModuleTitle>
  )
}
