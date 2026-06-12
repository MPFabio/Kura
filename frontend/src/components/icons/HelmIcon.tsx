import { Box, SxProps, Theme } from '@mui/material'
import helmLogo from '../../assets/helm_logo.png'

interface HelmIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function HelmIcon({ sx, active = false }: HelmIconProps) {
  return (
    <Box
      component="img"
      src={helmLogo}
      alt=""
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        flexShrink: 0,
        objectFit: 'contain',
        opacity: active ? 1 : 0.5,
        transition: 'opacity 0.15s ease',
        ...sx,
      }}
    />
  )
}
