import { Box } from '@mui/material'

/** Fond gris-bleu CLAIR (DA KURA) */
const BACKGROUND_DARK = '#2c2f3f'

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
        zIndex: -1,
      }}
    />
  )
}
