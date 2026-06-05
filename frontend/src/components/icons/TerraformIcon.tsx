import { Box, SxProps, Theme } from '@mui/material'
import terraformLogo from '../../assets/terraform_logo.png'

const ICON_COLOR_ACTIVE = '#7B42BC'
const ICON_COLOR_INACTIVE = '#6B7385'

interface TerraformIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function TerraformIcon({ sx, active = false }: TerraformIconProps) {
  return (
    <Box
      aria-hidden
      sx={{
        width: 24,
        height: 24,
        flexShrink: 0,
        ...sx,
        backgroundColor: active ? ICON_COLOR_ACTIVE : ICON_COLOR_INACTIVE,
        mask: `url(${terraformLogo}) center/contain no-repeat`,
        WebkitMask: `url(${terraformLogo}) center/contain no-repeat`,
        transition: 'background-color 0.15s ease',
      }}
    />
  )
}
