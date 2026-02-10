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
        background: '#00E5FF',
        color: '#2c2f3f',
        '&:hover': {
          background: '#26C6DA',
        },
        fontFamily: '"Inter", sans-serif',
        fontWeight: 700,
        textTransform: 'none',
        px: 3,
        py: 1.5,
        boxShadow: 'none',
        ...sx,
      }}
    >
      {children}
    </Button>
  )
}
