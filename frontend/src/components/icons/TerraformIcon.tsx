import { Box, SxProps, Theme } from '@mui/material'
import opentofuLogo from '../../assets/opentofu_logo.webp'

interface TerraformIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function TerraformIcon({ sx, active = false }: TerraformIconProps) {
  return (
    <Box
      component="img"
      src={opentofuLogo}
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
