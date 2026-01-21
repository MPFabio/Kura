import { Box, SxProps, Theme } from '@mui/material'

interface TerraformIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function TerraformIcon({ sx, active = false }: TerraformIconProps) {
  return (
    <Box
      component="svg"
      viewBox="0 0 24 24"
      sx={{
        width: 24,
        height: 24,
        fill: active ? '#00FFFF' : '#A0A0A0',
        filter: active ? 'drop-shadow(0 0 8px rgba(0, 255, 255, 0.8))' : 'none',
        transition: 'all 0.3s ease',
        ...sx,
      }}
    >
      {/* Terraform "T" stylisé */}
      <path d="M12 2L2 7L2 17L12 22L22 17L22 7L12 2Z" fill="currentColor" opacity="0.1" />
      <path
        d="M12 4L5 7.5L5 16.5L12 20L19 16.5L19 7.5L12 4Z"
        fill="currentColor"
        opacity="0.2"
      />
      {/* Lettre T centrale */}
      <path
        d="M10 8H14V10H12V16H10V10H10V8Z"
        fill="currentColor"
      />
      <path
        d="M10 8H14V9H10V8Z"
        fill="currentColor"
        opacity="0.8"
      />
    </Box>
  )
}
