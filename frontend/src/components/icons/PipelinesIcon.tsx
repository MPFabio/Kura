import { Box, SxProps, Theme } from '@mui/material'
import pipelineLogo from '../../assets/pipeline_logo.png'

interface PipelinesIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

/**
 * Logo Pipelines en image (comme Kubernetes) : le mask donne un carré avec ce PNG.
 * Actif = logo tel quel, inactif = gris (grayscale) pour uniformité sidebar.
 */
export default function PipelinesIcon({ sx, active = false }: PipelinesIconProps) {
  return (
    <Box
      component="img"
      src={pipelineLogo}
      alt=""
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        objectFit: 'contain',
        filter: active ? 'none' : 'grayscale(1) brightness(0.85) contrast(0.9)',
        opacity: active ? 1 : 0.9,
        transition: 'all 0.3s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
