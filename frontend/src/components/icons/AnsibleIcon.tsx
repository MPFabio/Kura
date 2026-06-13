import { Box, SxProps, Theme } from '@mui/material'
import ansibleLogo from '../../assets/semaphore_logo.png'

interface AnsibleIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function AnsibleIcon({ sx, active = false }: AnsibleIconProps) {
  return (
    <Box
      component="img"
      src={ansibleLogo}
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
