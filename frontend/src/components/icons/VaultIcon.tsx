import { Box, SxProps, Theme } from '@mui/material'
import vaultLogo from '../../assets/vault_logo.png'

interface VaultIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function VaultIcon({ sx, active = false }: VaultIconProps) {
  return (
    <Box
      component="img"
      src={vaultLogo}
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
