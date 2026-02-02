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
        background: 'linear-gradient(135deg, #00E5FF, #EC407A)',
        '&:hover': {
          background: 'linear-gradient(135deg, #00B8D4, #9C6ADE)',
        },
        fontFamily: '"Inter", sans-serif',
        fontWeight: 500,
        textTransform: 'none',
        px: 3,
        py: 1.5,
        boxShadow: '0 0 12px rgba(0, 229, 255, 0.3)',
        ...sx,
      }}
    >
      {children}
    </Button>
  )
}
