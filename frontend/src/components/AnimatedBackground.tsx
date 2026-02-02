import { Box } from '@mui/material'

/** Fond charbon fixe (DA KURA) */
const BACKGROUND_DARK = '#0d0e12'

/**
 * Fond sombre fixe, cohérent avec le thème inversé
 * Utilisé sur toutes les pages pour un style cohérent
 */
export default function AnimatedBackground() {
  return (
    <Box
      sx={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        background: BACKGROUND_DARK,
        pointerEvents: 'none',
        zIndex: 0,
      }}
    />
  )
}
