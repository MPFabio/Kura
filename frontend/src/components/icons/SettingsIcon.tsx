import { Box, SxProps, Theme } from '@mui/material'
import parametresLogo from '../../assets/parametres_logo.png'

const ICON_COLOR_ACTIVE = '#6BA4B8'
const ICON_COLOR_INACTIVE = '#6B7385'

interface SettingsIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

/** Même rendu que Terraform/Ansible : mask + couleur unie (cyan actif, gris inactif). */
export default function SettingsIcon({ sx, active = false }: SettingsIconProps) {
  return (
    <Box
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        flexShrink: 0,
        ...sx,
        backgroundColor: active ? ICON_COLOR_ACTIVE : ICON_COLOR_INACTIVE,
        mask: `url(${parametresLogo}) center/contain no-repeat`,
        WebkitMask: `url(${parametresLogo}) center/contain no-repeat`,
        transition: 'background-color 0.15s ease',
      }}
    />
  )
}
