import { Box, SxProps, Theme } from '@mui/material'
import parametresLogo from '../../assets/parametres_logo.png'

const ICON_COLOR_ACTIVE = '#00E5FF'
const ICON_COLOR_INACTIVE = '#b8b8b8'

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
        backgroundColor: active ? ICON_COLOR_ACTIVE : ICON_COLOR_INACTIVE,
        mask: `url(${parametresLogo}) center/contain no-repeat`,
        WebkitMask: `url(${parametresLogo}) center/contain no-repeat`,
        maskSize: 'contain',
        WebkitMaskSize: 'contain',
        maskRepeat: 'no-repeat',
        WebkitMaskRepeat: 'no-repeat',
        maskPosition: 'center',
        WebkitMaskPosition: 'center',
        filter: 'none',
        transition: 'all 0.3s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
