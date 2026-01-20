import { createTheme } from '@mui/material/styles'

// Palette "Abyssal Glow"
export const abyssalColors = {
  // Fonds
  abyssalBlack: '#050508',
  abyssalDark: '#0A0A0E',
  
  // Bioluminescence principale
  turquoiseElectric: '#00FFFF',
  turquoiseLight: '#00E0E0',
  cyanElectric: '#00FFFF',
  
  // Bioluminescence secondaire
  violetNeon: '#BF00FF',
  magentaNeon: '#FF00BF',
  
  // Gris froids
  grayLight: '#A0A0A0',
  grayMedium: '#808080',
  grayDark: '#606060',
  
  // Alertes
  alertRed: '#FF4500',
  alertYellow: '#FFD700',
}

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: abyssalColors.turquoiseElectric,
      light: abyssalColors.turquoiseLight,
      dark: abyssalColors.cyanElectric,
    },
    secondary: {
      main: abyssalColors.violetNeon,
      light: abyssalColors.magentaNeon,
      dark: abyssalColors.violetNeon,
    },
    background: {
      default: abyssalColors.abyssalBlack,
      paper: abyssalColors.abyssalDark,
    },
    text: {
      primary: '#FFFFFF',
      secondary: abyssalColors.grayLight,
    },
    error: {
      main: abyssalColors.alertRed,
    },
    warning: {
      main: abyssalColors.alertYellow,
    },
  },
  typography: {
    fontFamily: [
      'Inter',
      '-apple-system',
      'BlinkMacSystemFont',
      '"Segoe UI"',
      'Roboto',
      '"Helvetica Neue"',
      'Arial',
      'sans-serif',
    ].join(','),
    h1: {
      fontSize: '2.5rem',
      fontWeight: 600,
      letterSpacing: '0.02em',
      color: abyssalColors.turquoiseElectric,
      textShadow: `0 0 20px ${abyssalColors.turquoiseElectric}40`,
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 600,
      letterSpacing: '0.02em',
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 600,
      letterSpacing: '0.01em',
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 600,
      letterSpacing: '0.01em',
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 600,
    },
    h6: {
      fontSize: '1rem',
      fontWeight: 600,
    },
    body1: {
      fontFamily: '"JetBrains Mono", "Space Mono", "IBM Plex Mono", monospace',
    },
    body2: {
      fontFamily: '"JetBrains Mono", "Space Mono", "IBM Plex Mono", monospace',
    },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          borderRadius: 12,
          border: `1px solid ${abyssalColors.turquoiseElectric}`,
          background: 'transparent',
          color: '#FFFFFF',
          padding: '10px 24px',
          transition: 'all 0.3s ease',
          '&:hover': {
            borderWidth: '2px',
            boxShadow: `0 0 20px ${abyssalColors.turquoiseElectric}60, 0 0 40px ${abyssalColors.turquoiseElectric}40`,
            background: `${abyssalColors.turquoiseElectric}10`,
          },
        },
        contained: {
          background: `linear-gradient(135deg, ${abyssalColors.turquoiseElectric}, ${abyssalColors.violetNeon})`,
          border: 'none',
          '&:hover': {
            background: `linear-gradient(135deg, ${abyssalColors.turquoiseLight}, ${abyssalColors.magentaNeon})`,
            boxShadow: `0 0 30px ${abyssalColors.turquoiseElectric}80, 0 0 60px ${abyssalColors.violetNeon}60`,
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          background: `linear-gradient(135deg, ${abyssalColors.abyssalDark}CC, ${abyssalColors.abyssalBlack}DD)`,
          backdropFilter: 'blur(10px)',
          border: `1px solid ${abyssalColors.turquoiseElectric}30`,
          borderRadius: 16,
          boxShadow: `0 4px 20px ${abyssalColors.abyssalBlack}80, inset 0 1px 0 ${abyssalColors.turquoiseElectric}20`,
          transition: 'all 0.3s ease',
          '&:hover': {
            borderColor: `${abyssalColors.turquoiseElectric}60`,
            boxShadow: `0 8px 40px ${abyssalColors.turquoiseElectric}40, inset 0 1px 0 ${abyssalColors.turquoiseElectric}40`,
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: abyssalColors.abyssalDark,
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          background: `${abyssalColors.abyssalDark}E6`,
          backdropFilter: 'blur(20px)',
          borderRight: `1px solid ${abyssalColors.turquoiseElectric}20`,
          boxShadow: `2px 0 20px ${abyssalColors.abyssalBlack}80`,
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: `${abyssalColors.abyssalDark}E6`,
          backdropFilter: 'blur(20px)',
          borderBottom: `1px solid ${abyssalColors.turquoiseElectric}20`,
          boxShadow: `0 2px 20px ${abyssalColors.abyssalBlack}80`,
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          margin: '4px 8px',
          transition: 'all 0.3s ease',
          '&:hover': {
            background: `${abyssalColors.turquoiseElectric}15`,
            boxShadow: `0 0 15px ${abyssalColors.turquoiseElectric}30`,
          },
          '&.Mui-selected': {
            background: `${abyssalColors.turquoiseElectric}25`,
            borderLeft: `3px solid ${abyssalColors.turquoiseElectric}`,
            boxShadow: `0 0 20px ${abyssalColors.turquoiseElectric}50`,
            '&:hover': {
              background: `${abyssalColors.turquoiseElectric}35`,
            },
          },
        },
      },
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: abyssalColors.grayLight,
          transition: 'color 0.3s ease',
          '.Mui-selected &': {
            color: abyssalColors.turquoiseElectric,
            filter: `drop-shadow(0 0 8px ${abyssalColors.turquoiseElectric})`,
          },
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          background: `${abyssalColors.abyssalDark}DD`,
          backdropFilter: 'blur(10px)',
          border: `1px solid ${abyssalColors.turquoiseElectric}40`,
          borderRadius: 12,
        },
        standardInfo: {
          borderColor: `${abyssalColors.turquoiseElectric}40`,
          '& .MuiAlert-icon': {
            color: abyssalColors.turquoiseElectric,
            filter: `drop-shadow(0 0 8px ${abyssalColors.turquoiseElectric})`,
          },
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          background: `linear-gradient(135deg, ${abyssalColors.abyssalDark}EE, ${abyssalColors.abyssalBlack}FF)`,
          backdropFilter: 'blur(20px)',
          borderRadius: 16,
          border: 'none',
          position: 'relative',
          isolation: 'isolate',
          boxShadow: `
            0 8px 32px ${abyssalColors.abyssalBlack}CC,
            inset 0 1px 0 rgba(255, 255, 255, 0.1)
          `,
          '&::before': {
            content: '""',
            position: 'absolute',
            inset: '-3px',
            borderRadius: '19px',
            background: 'linear-gradient(135deg, #BF00FF 0%, #FF00BF 50%, #00FFFF 100%)',
            zIndex: -1,
            opacity: 1,
            boxShadow: `
              0 0 20px rgba(191, 0, 255, 0.8),
              0 0 40px rgba(0, 255, 255, 0.6),
              0 0 60px rgba(191, 0, 255, 0.4)
            `,
            animation: 'glowPulse 2s ease-in-out infinite',
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            inset: '3px',
            borderRadius: '13px',
            background: `linear-gradient(135deg, ${abyssalColors.abyssalDark}EE, ${abyssalColors.abyssalBlack}FF)`,
            backdropFilter: 'blur(20px)',
            zIndex: -1,
          },
        },
      },
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          background: 'transparent',
          borderBottom: `1px solid rgba(0, 255, 255, 0.2)`,
          padding: '16px 24px',
        },
      },
    },
    MuiDialogContent: {
      styleOverrides: {
        root: {
          background: 'transparent',
          padding: '24px',
        },
      },
    },
    MuiDialogActions: {
      styleOverrides: {
        root: {
          background: 'transparent',
          borderTop: `1px solid rgba(191, 0, 255, 0.2)`,
          padding: '16px 24px',
        },
      },
    },
  },
})
