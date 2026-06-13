import { Box, SxProps, Theme } from '@mui/material'
import forgejoLogo from '../../assets/forgejo_logo.png'

interface ForgejoIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function ForgejoIcon({ sx, active = false }: ForgejoIconProps) {
  return (
    <Box
      component="img"
      src={forgejoLogo}
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
