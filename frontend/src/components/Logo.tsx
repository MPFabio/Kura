import { Box, SxProps, Theme } from '@mui/material'
import jellyfishLogo from '../assets/jellyfish_logo.png'

interface LogoProps {
  variant?: 'full' | 'icon'
  size?: 'small' | 'medium' | 'large'
  sx?: SxProps<Theme>
}

const sizeMap = {
  small:  96,
  medium: 186,
  large:  282,
}

export default function Logo({ variant = 'full', size = 'medium', sx }: LogoProps) {
  const px = sizeMap[size]

  if (variant === 'icon') {
    return (
      <Box
        component="img"
        src={jellyfishLogo}
        alt="Kura"
        sx={{ width: px, height: px, objectFit: 'contain', display: 'block', ...sx }}
      />
    )
  }

  return (
    <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', ...sx }}>
      <Box
        component="img"
        src={jellyfishLogo}
        alt="Kura"
        sx={{ width: px, height: px, objectFit: 'contain', display: 'block' }}
      />
    </Box>
  )
}

export const kuraWordmarkSx = {}
