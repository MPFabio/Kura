import { Box, SxProps, Theme } from '@mui/material'

interface AnsibleIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function AnsibleIcon({ sx, active = false }: AnsibleIconProps) {
  return (
    <Box
      component="svg"
      viewBox="0 0 24 24"
      sx={{
        width: 24,
        height: 24,
        fill: active ? '#BF00FF' : '#A0A0A0',
        filter: active ? 'drop-shadow(0 0 8px rgba(191, 0, 255, 0.8))' : 'none',
        transition: 'all 0.3s ease',
        ...sx,
      }}
    >
      {/* Ansible "A" stylisé */}
      <path
        d="M12 6L9 18H10.5L11.5 14H12.5L13.5 18H15L12 6Z"
        fill="currentColor"
      />
      <path
        d="M11.2 12L12 8.5L12.8 12H11.2Z"
        fill={active ? '#BF00FF' : '#A0A0A0'}
        opacity="0.3"
      />
      {/* Barre horizontale du A */}
      <line
        x1="10.5"
        y1="11"
        x2="13.5"
        y2="11"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </Box>
  )
}
