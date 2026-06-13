import { Box, SxProps, Theme } from '@mui/material'
import zotLogo from '../../assets/zot_logo.png'

interface ZotIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function ZotIcon({ sx, active = false }: ZotIconProps) {
  return (
    <Box
      component="img"
      src={zotLogo}
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
