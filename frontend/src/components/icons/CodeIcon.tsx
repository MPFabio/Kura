import { Box, SxProps, Theme } from '@mui/material'
// import githubLogo from '../../assets/github_logo.png' // conservé mais désactivé en prod (remplacé par forgejo_logo)
import forgejoLogo from '../../assets/forgejo_logo.png'

interface CodeIconProps {
  sx?: SxProps<Theme>
  active?: boolean
}

export default function CodeIcon({ sx, active = false }: CodeIconProps) {
  return (
    <Box
      component="img"
      src={forgejoLogo}
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
