import { styled } from '@mui/material/styles'
import { Typography, SxProps, Theme } from '@mui/material'
import { useEffect, useRef } from 'react'

interface ModuleTitleProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

const StyledModuleTitle = styled(Typography, {
  shouldForwardProp: (prop) => prop !== 'sx',
})({
  fontWeight: '600 !important',
  background: 'linear-gradient(135deg, #00E5FF, #EC407A) !important',
  backgroundClip: 'text !important',
  WebkitBackgroundClip: 'text !important',
  WebkitTextFillColor: 'transparent !important',
  color: 'transparent !important', // Forcer color transparent pour que le dégradé soit visible
  fontFamily: '"Inter", sans-serif !important',
  letterSpacing: '0.02em !important',
  fontSize: '2.5rem !important',
  lineHeight: '1.2 !important',
  // Override complet des styles du theme
  '&.MuiTypography-root': {
    fontWeight: '600 !important',
    fontSize: '2.5rem !important',
    letterSpacing: '0.02em !important',
    fontFamily: '"Inter", sans-serif !important',
    color: 'transparent !important', // Forcer color transparent dans le sélecteur MUI aussi
  },
}) as typeof Typography

export default function ModuleTitle({ children, sx }: ModuleTitleProps) {
  const ref = useRef<HTMLHeadingElement>(null)

  useEffect(() => {
    if (ref.current) {
      const element = ref.current
      
      // Forcer color transparent directement sur l'élément DOM pour override MUI
      // Les styles inline ont la priorité la plus élevée en CSS
      element.style.setProperty('color', 'transparent', 'important')
      element.style.setProperty('-webkit-text-fill-color', 'transparent', 'important')
      
      // Vérifier aussi le background et forcer son application
      const computedStyle = window.getComputedStyle(element)
      
      // #region agent log
      fetch('http://127.0.0.1:7243/ingest/0f545d4a-2194-41e7-8131-580ff06d52fc', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          location: 'ModuleTitle.tsx:useEffect',
          message: 'ModuleTitle - detailed style inspection',
          data: {
            titleText: typeof children === 'string' ? children : 'ModuleTitle',
            computedBackground: computedStyle.background,
            computedBackgroundImage: computedStyle.backgroundImage,
            computedBackgroundClip: computedStyle.backgroundClip,
            computedWebkitBackgroundClip: computedStyle.webkitBackgroundClip,
            computedColor: computedStyle.color,
            computedWebkitTextFillColor: computedStyle.webkitTextFillColor,
            inlineColor: element.style.color,
            inlineWebkitTextFillColor: element.style.webkitTextFillColor,
            inlineBackground: element.style.background,
            allComputedStyles: {
              background: computedStyle.background,
              backgroundImage: computedStyle.backgroundImage,
              backgroundClip: computedStyle.backgroundClip,
              webkitBackgroundClip: computedStyle.webkitBackgroundClip,
              color: computedStyle.color,
              webkitTextFillColor: computedStyle.webkitTextFillColor,
            },
            className: element.className,
            sxProp: sx ? JSON.stringify(sx) : null,
          },
          timestamp: Date.now(),
          sessionId: 'debug-session',
          runId: 'detailed-inspection',
          hypothesisId: 'E',
        }),
      }).catch(() => {})
      // #endregion
      
      // Forcer aussi le background directement si nécessaire
      if (!computedStyle.backgroundImage || !computedStyle.backgroundImage.includes('gradient')) {
        element.style.setProperty('background', 'linear-gradient(135deg, #00E5FF, #EC407A)', 'important')
        element.style.setProperty('background-clip', 'text', 'important')
        element.style.setProperty('-webkit-background-clip', 'text', 'important')
      }
    }
  }, [children, sx])

  return (
    <StyledModuleTitle
      ref={ref}
      variant="h3"
      component="h1"
      className="module-title"
      sx={{
        mb: 5,
        ...sx,
        // Forcer le dégradé après sx pour s'assurer qu'il n'est jamais override
        background: 'linear-gradient(135deg, #00E5FF, #EC407A) !important',
        backgroundClip: 'text !important',
        WebkitBackgroundClip: 'text !important',
        WebkitTextFillColor: 'transparent !important',
        color: 'transparent !important', // Forcer color transparent pour que le dégradé soit visible
      }}
    >
      {children}
    </StyledModuleTitle>
  )
}
