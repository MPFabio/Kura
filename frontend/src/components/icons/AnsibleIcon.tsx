import { Box, SxProps, Theme } from '@mui/material'
import ansibleLogo from '../../assets/ansible_logo.png'

const ICON_COLOR_ACTIVE = '#00E5FF'
const ICON_COLOR_INACTIVE = '#b8b8b8'

interface AnsibleIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function AnsibleIcon({ sx, active = false }: AnsibleIconProps) {
  return (
    <Box
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        backgroundColor: active ? ICON_COLOR_ACTIVE : ICON_COLOR_INACTIVE,
        mask: `url(${ansibleLogo}) center/contain no-repeat`,
        WebkitMask: `url(${ansibleLogo}) center/contain no-repeat`,
        maskSize: 'contain',
        WebkitMaskSize: 'contain',
        maskRepeat: 'no-repeat',
        WebkitMaskRepeat: 'no-repeat',
        maskPosition: 'center',
        WebkitMaskPosition: 'center',
        filter: active ? 'drop-shadow(0 0 6px rgba(0, 229, 255, 0.6))' : 'none',
        transition: 'all 0.3s ease',
        flexShrink: 0,
        ...sx,
      }}
    />
  )
}
