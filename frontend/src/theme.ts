import { createTheme } from '@mui/material/styles'

// Palette DA KURA : fond gris-bleu clair, accents cyan/magenta/violet
export const jellyfishColors = {
  // Fonds gris-bleu CLAIRS (lisibles et agréables)
  backgroundLight: '#2c2f3f',   // fond principal gris-bleu clair
  backgroundPaper: '#32364a',   // cartes / paper légèrement plus clair
  backgroundCard: '#383c50',   // cartes

  // Cloche méduse - cyan / bleu
  cyanSoft: '#00E5FF',      // Cyan vif (haut cloche)
  cyanMedium: '#26C6DA',    // #26C6DA cyan 400 - bleu-cyan
  cyanLight: '#4DD0E1',     // #4DD0E1 cyan 300
  cyanDeep: '#00ACC1',      // #00ACC1 cyan 600 - bleu plus profond
  cyanSubtle: 'rgba(0, 229, 255, 0.2)',

  // Bas cloche + tentacules - violet-rose / magenta / fuchsia (logo)
  violetSoft: '#CE93D8',    // #CE93D8 violet 200 - violet clair
  violetMedium: '#AB47BC',   // #AB47BC violet 400 - violet moyen
  violetDeep: '#7B1FA2',    // #7B1FA2 violet 700 - violet foncé
  violetLight: '#E1BEE7',   // violet 100
  // Violet-rose / magenta / fuchsia (tentacules)
  magenta: '#EC407A',       // #EC407A Pink 400 - rose vif
  magentaDeep: '#E91E63',   // #E91E63 Pink 600 - magenta
  fuchsia: '#E040FB',       // #E040FB Accent A200 - fuchsia (tentacules)
  violetRed: '#D500F9',     // #D500F9 Accent A400 - violet-rouge
  violetSubtle: 'rgba(206, 147, 216, 0.25)',
  frameViolet: 'rgba(171, 71, 188, 0.65)',     // #AB47BC - cadre visible comme la méduse
  frameVioletRed: 'rgba(236, 64, 122, 0.55)',  // #EC407A magenta - bordure violet-rose
  magentaSoft: 'rgba(236, 64, 122, 0.4)',      // #EC407A

  // Gris (texte sur fond clair)
  grayLight: '#546E7A',
  grayMedium: '#455A64',
  grayDark: '#37474F',
  graySubtle: 'rgba(55, 71, 79, 0.08)',

  // Alertes
  successSoft: '#66BB6A',
  warningSoft: '#FFA726',
  errorSoft: '#EF5350',
  infoSoft: '#42A5F5',
}

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: jellyfishColors.cyanSoft,
      light: jellyfishColors.cyanLight,
      dark: jellyfishColors.cyanMedium,
    },
    secondary: {
      main: jellyfishColors.violetSoft,
      light: jellyfishColors.violetLight,
      dark: jellyfishColors.violetMedium,
    },
    background: {
      default: jellyfishColors.backgroundLight,
      paper: jellyfishColors.backgroundPaper,
    },
    text: {
      primary: '#f0f0f0',
      secondary: '#b8b8b8',
    },
    error: {
      main: jellyfishColors.errorSoft,
    },
    warning: {
      main: jellyfishColors.warningSoft,
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
      fontSize: '3rem',
      fontWeight: 700,
      letterSpacing: '-0.04em',
      color: '#f0f0f0',
    },
    h2: {
      fontSize: '2.25rem',
      fontWeight: 700,
      letterSpacing: '-0.03em',
      color: '#f0f0f0',
    },
    h3: {
      fontSize: '1.875rem',
      fontWeight: 700,
      letterSpacing: '-0.02em',
      color: '#f0f0f0',
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 700,
      letterSpacing: '-0.02em',
      color: '#f0f0f0',
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 700,
      color: '#f0f0f0',
    },
    h6: {
      fontSize: '1.125rem',
      fontWeight: 700,
      color: '#f0f0f0',
    },
    body1: {
      fontSize: '0.9375rem',
      lineHeight: 1.6,
      fontWeight: 400,
      color: '#f0f0f0',
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.5,
      fontWeight: 400,
      color: '#b8b8b8',
    },
  },
  shape: {
    borderRadius: 0,
  },
  spacing: 8,
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          borderRadius: 0,
          fontWeight: 700,
          padding: '12px 24px',
          fontSize: '0.9375rem',
          transition: 'all 0.15s ease',
          boxShadow: 'none',
          '&:hover': {
            boxShadow: 'none',
          },
        },
        outlined: {
          border: `2px solid ${jellyfishColors.cyanSoft}`,
          color: jellyfishColors.cyanSoft,
          background: 'transparent',
          '&:hover': {
            background: 'rgba(0, 229, 255, 0.08)',
            borderColor: jellyfishColors.cyanSoft,
          },
        },
        contained: {
          background: jellyfishColors.cyanSoft,
          color: '#0d0e12',
          fontWeight: 700,
          '&:hover': {
            background: jellyfishColors.cyanMedium,
          },
        },
        text: {
          color: jellyfishColors.cyanSoft,
          fontWeight: 600,
          '&:hover': {
            background: 'rgba(0, 229, 255, 0.08)',
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          background: jellyfishColors.backgroundCard,
          border: '1px solid rgba(0, 229, 255, 0.15)',
          borderRadius: 0,
          boxShadow: 'none',
          transition: 'border-color 0.15s ease',
          position: 'relative',
          overflow: 'hidden',
          '&:hover': {
            borderColor: 'rgba(0, 229, 255, 0.25)',
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          background: '#14161f',
          border: '1px solid rgba(255, 255, 255, 0.08)',
          borderRadius: 0,
          boxShadow: 'none',
          transition: 'all 0.15s ease',
          position: 'relative',
          overflow: 'hidden',
          '&.MuiDrawer-paper': {
            background: 'transparent !important',
            backgroundColor: 'transparent !important',
            backgroundImage: 'none !important',
            border: 'none !important',
          },
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          background: jellyfishColors.backgroundLight,
          borderRight: '2px solid rgba(0, 229, 255, 0.2)',
          borderRadius: 0,
          boxShadow: 'none',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: jellyfishColors.backgroundLight,
          borderBottom: '2px solid rgba(0, 229, 255, 0.2)',
          borderRadius: 0,
          boxShadow: 'none',
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 0,
          margin: '0',
          padding: '12px 16px',
          transition: 'all 0.2s ease',
          '&:hover': {
            background: 'rgba(0, 229, 255, 0.08)',
          },
          '&.Mui-selected': {
            background: 'rgba(0, 229, 255, 0.15)',
            borderLeft: '4px solid',
            borderLeftColor: jellyfishColors.cyanSoft,
            paddingLeft: '12px',
            '&:hover': {
              background: 'rgba(0, 229, 255, 0.2)',
            },
          },
        },
      },
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: jellyfishColors.grayLight,
          transition: 'color 0.2s ease',
          minWidth: 40,
          '.Mui-selected &': {
            color: jellyfishColors.cyanSoft,
          },
        },
      },
    },
    MuiListItemText: {
      styleOverrides: {
        primary: {
          fontWeight: 500,
          // Items non sélectionnés : texte gris
          '.MuiListItemButton-root:not(.Mui-selected) &': {
            color: '#b8b8b8 !important',
          },
          // Item sélectionné : texte clair
          '.MuiListItemButton-root.Mui-selected &': {
            color: '#f0f0f0 !important',
          },
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          background: jellyfishColors.backgroundPaper,
          border: `1px solid rgba(255, 255, 255, 0.08)`,
          borderRadius: 0,
          boxShadow: 'none',
          padding: '16px 20px',
        },
        standardInfo: {
          borderLeftColor: jellyfishColors.cyanSoft,
          borderLeftWidth: '4px',
          '& .MuiAlert-icon': {
            color: jellyfishColors.cyanSoft,
          },
        },
        standardSuccess: {
          borderLeftColor: jellyfishColors.successSoft,
          borderLeftWidth: '4px',
        },
        standardWarning: {
          borderLeftColor: jellyfishColors.warningSoft,
          borderLeftWidth: '4px',
        },
        standardError: {
          borderLeftColor: jellyfishColors.errorSoft,
          borderLeftWidth: '4px',
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          background: jellyfishColors.backgroundPaper,
          borderRadius: 0,
          border: `1px solid rgba(255, 255, 255, 0.08)`,
          boxShadow: 'none',
        },
      },
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          background: 'transparent',
          borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
          padding: '20px 24px',
          color: '#f0f0f0',
          fontWeight: 700,
          fontSize: '1.25rem',
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
          borderTop: '1px solid rgba(255, 255, 255, 0.08)',
          padding: '20px 24px',
          gap: 12,
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 0,
            background: jellyfishColors.backgroundPaper,
            '& fieldset': {
              borderColor: 'rgba(255, 255, 255, 0.08)',
              borderWidth: '1px',
            },
            '&:hover fieldset': {
              borderColor: 'rgba(0, 229, 255, 0.3)',
            },
            '&.Mui-focused fieldset': {
              borderColor: jellyfishColors.cyanSoft,
              borderWidth: '2px',
            },
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 0,
          fontWeight: 600,
          fontSize: '0.75rem',
          border: '1px solid',
        },
        colorPrimary: {
          backgroundColor: 'transparent',
          borderColor: jellyfishColors.cyanSoft,
          color: jellyfishColors.cyanSoft,
        },
        colorSecondary: {
          backgroundColor: 'transparent',
          borderColor: jellyfishColors.violetSoft,
          color: jellyfishColors.violetSoft,
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        root: {
          borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
        },
        indicator: {
          background: jellyfishColors.cyanSoft,
          height: 3,
        },
      },
    },
    MuiTypography: {
      styleOverrides: {
        root: {
          // Les composants styled (ModuleTitle, ModuleText) override ces valeurs
          // Ne pas définir de styles par défaut qui pourraient override
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
          fontSize: '0.9375rem',
          color: '#b8b8b8',
          padding: '12px 20px',
          '&:hover': {
            color: jellyfishColors.cyanSoft,
          },
          '&.Mui-selected': {
            color: '#f0f0f0',
            fontWeight: 700,
          },
        },
      },
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: 0,
          border: `1px solid rgba(255, 255, 255, 0.08)`,
          background: jellyfishColors.backgroundPaper,
          boxShadow: 'none',
          overflow: 'hidden',
        },
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-head': {
            background: 'rgba(0, 229, 255, 0.08)',
            color: '#e8e8e8',
            fontWeight: 700,
            fontSize: '0.875rem',
            borderBottom: '2px solid rgba(0, 229, 255, 0.2)',
            padding: '16px 20px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          },
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:hover': {
            background: 'rgba(0, 229, 255, 0.05)',
          },
          '& .MuiTableCell-body': {
            borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
            padding: '16px 20px',
          },
        },
      },
    },
  },
})
