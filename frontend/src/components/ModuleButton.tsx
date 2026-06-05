import { Button, ButtonProps, SxProps, Theme } from '@mui/material'

interface ModuleButtonProps extends ButtonProps {
  children: React.ReactNode
  sx?: SxProps<Theme>
}

export default function ModuleButton({ children, sx, ...props }: ModuleButtonProps) {
  return (
    <Button
      variant="contained"
      {...props}
      sx={{
        fontFamily: '"Inter", sans-serif',
        fontWeight: 500,
        textTransform: 'none',
        px: 2,
        py: 0.875,
        boxShadow: 'none',
        ...sx,
      }}
    >
      {children}
    </Button>
  )
}
