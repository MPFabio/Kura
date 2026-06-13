import { Box, SxProps, Theme } from '@mui/material'
import tempoLogo from '../../assets/tempo_logo.png'

interface TempoIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function TempoIcon({ sx, active = false }: TempoIconProps) {
  return (
    <Box
      component="img"
      src={tempoLogo}
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
