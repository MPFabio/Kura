import { createTheme } from '@mui/material/styles'

// Palette DA KURA : charbon, bioluminescence cyan/magenta/violet, glassmorphism
export const jellyfishColors = {
  // Fonds charbon (DA)
  backgroundLight: '#0d0e12',   // charbon principal
  backgroundPaper: '#14161f',   // cartes / paper légèrement plus clair
  backgroundCard: '#1a1d28',   // cartes

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
      fontSize: '2.5rem',
      fontWeight: 500,
      letterSpacing: '-0.02em',
      color: '#f0f0f0',
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 500,
      letterSpacing: '-0.01em',
      color: '#f0f0f0',
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 500,
      letterSpacing: '-0.01em',
      // Les composants ModuleTitle override ces valeurs
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 500,
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 500,
    },
    h6: {
      fontSize: '1rem',
      fontWeight: 500,
    },
    body1: {
      fontSize: '1rem',
      lineHeight: 1.6,
      color: '#f0f0f0',
    },
    body2: {
      fontSize: '1rem',
      lineHeight: 1.5,
      color: '#b8b8b8',
    },
  },
  shape: {
    borderRadius: 8,
  },
  spacing: 8,
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          borderRadius: 8,
          fontWeight: 500,
          padding: '10px 20px',
          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
          boxShadow: 'none',
        },
        outlined: {
          border: `2px solid ${jellyfishColors.cyanSoft}60`,
          color: jellyfishColors.cyanSoft,
          background: 'transparent',
          borderRadius: 12,
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          boxShadow: `0 0 8px ${jellyfishColors.cyanSoft}30`,
          '&:hover': {
            border: `2px solid ${jellyfishColors.cyanSoft}80`,
            background: `${jellyfishColors.cyanSoft}15`,
            boxShadow: `0 0 16px ${jellyfishColors.cyanSoft}50`,
          },
        },
        contained: {
          background: `linear-gradient(135deg, ${jellyfishColors.cyanSoft}, ${jellyfishColors.magenta})`,
          color: '#fff',
          border: 'none',
          boxShadow: `0 4px 16px ${jellyfishColors.cyanSoft}30`,
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          '&:hover': {
            background: `linear-gradient(135deg, ${jellyfishColors.cyanMedium}, ${jellyfishColors.magentaDeep})`,
            boxShadow: `0 6px 20px ${jellyfishColors.cyanSoft}35, 0 0 16px ${jellyfishColors.magenta}40`,
          },
        },
        text: {
          color: jellyfishColors.cyanSoft,
          '&:hover': {
            background: jellyfishColors.cyanSubtle,
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          background: 'rgba(20, 22, 31, 0.6)',
          backdropFilter: 'blur(20px) saturate(150%)',
          border: '1px solid rgba(255, 255, 255, 0.06)',
          borderRadius: 20,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.03)',
          transition: 'all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          position: 'relative',
          overflow: 'hidden',
          animationDelay: 'var(--card-delay, 0s)',
          '&:hover': {
            borderColor: 'rgba(0, 229, 255, 0.15)',
            boxShadow: '0 12px 40px rgba(0, 0, 0, 0.5), 0 0 24px rgba(0, 229, 255, 0.06), inset 0 1px 0 rgba(255, 255, 255, 0.04)',
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          background: 'rgba(20, 22, 31, 0.65)',
          backdropFilter: 'blur(20px) saturate(150%)',
          border: '1px solid rgba(255, 255, 255, 0.06)',
          borderRadius: 20,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.03)',
          transition: 'all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          position: 'relative',
          overflow: 'hidden',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          background: 'rgba(13, 14, 18, 0.85)',
          backdropFilter: 'blur(24px) saturate(150%)',
          borderRight: '1px solid rgba(255, 255, 255, 0.06)',
          borderRadius: '0 20px 20px 0',
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.5), inset 1px 0 0 rgba(0, 229, 255, 0.08)',
          position: 'fixed',
          height: '100vh',
          overflowY: 'hidden',
          overflowX: 'hidden',
          transition: 'all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          zIndex: 1200,
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: `linear-gradient(180deg, ${jellyfishColors.backgroundPaper}E6, ${jellyfishColors.backgroundLight}E6)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          borderBottom: 'none',
          borderLeft: 'none',
          borderRadius: '16px',
          boxShadow: `
            0 4px 16px rgba(236, 64, 122, 0.2),
            0 2px 8px rgba(171, 71, 188, 0.12),
            inset 0 1px 0 rgba(255, 255, 255, 0.8),
            0 0 0 1px ${jellyfishColors.frameViolet}
          `,
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `linear-gradient(180deg, ${jellyfishColors.violetMedium}22 0%, transparent 30%, transparent 70%, ${jellyfishColors.magenta}18 100%)`,
            animation: 'etherealGlow 8s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            borderRadius: '16px',
            background: `linear-gradient(135deg, rgba(236, 64, 122, 0.15) 0%, transparent 50%, rgba(171, 71, 188, 0.12) 100%)`,
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          margin: '2px 8px',
          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
          '&:hover': {
            background: `linear-gradient(90deg, ${jellyfishColors.cyanSubtle}, ${jellyfishColors.violetSubtle})`,
          },
          '&.Mui-selected': {
            background: `linear-gradient(90deg, ${jellyfishColors.magenta}35, ${jellyfishColors.violetMedium}30)`,
            borderRadius: 12,
            boxShadow: `
              0 0 16px rgba(236, 64, 122, 0.35),
              0 0 8px rgba(171, 71, 188, 0.25),
              inset 0 0 20px rgba(236, 64, 122, 0.08)
            `,
            position: 'relative',
            '&::before': {
              content: '""',
              position: 'absolute',
              left: 0,
              top: 0,
              bottom: 0,
              width: '3px',
              background: `linear-gradient(180deg, ${jellyfishColors.magenta}, ${jellyfishColors.cyanSoft})`,
              borderRadius: '12px 0 0 12px',
              zIndex: 1,
            },
            '&:hover': {
              background: `linear-gradient(90deg, ${jellyfishColors.magenta}45, ${jellyfishColors.violetMedium}40)`,
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
          // Dégradé logo (cyan → magenta) pour les items non sélectionnés
          '.MuiListItemButton-root:not(.Mui-selected) &': {
            background: 'linear-gradient(135deg, #00E5FF, #EC407A) !important',
            backgroundClip: 'text !important',
            WebkitBackgroundClip: 'text !important',
            WebkitTextFillColor: 'transparent !important',
            color: 'transparent !important',
          },
          // Item sélectionné : texte clair pour lisibilité sur fond magenta/violet
          '.MuiListItemButton-root.Mui-selected &': {
            color: '#f0f0f0 !important',
            background: 'none !important',
            WebkitTextFillColor: '#f0f0f0 !important',
          },
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          background: `linear-gradient(135deg, ${jellyfishColors.backgroundPaper}CC, ${jellyfishColors.backgroundLight}CC)`,
          backdropFilter: 'blur(35px) saturate(200%)',
          border: `1px solid ${jellyfishColors.frameViolet}`,
          borderRadius: 24,
          boxShadow: `
            0 8px 32px rgba(236, 64, 122, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.8)
          `,
          padding: '24px 28px',
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          animation: 'jellyfishFloat 12s ease-in-out infinite',
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: '-50%',
            left: '-50%',
            width: '200%',
            height: '200%',
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}15 0%, ${jellyfishColors.violetSoft}10 50%, transparent 70%)`,
            animation: 'etherealGlow 8s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
        standardInfo: {
          background: `linear-gradient(135deg, rgba(0, 229, 255, 0.12), rgba(236, 64, 122, 0.1))`,
          '& .MuiAlert-icon': {
            color: jellyfishColors.cyanSoft,
            filter: `drop-shadow(0 0 12px ${jellyfishColors.cyanSoft})`,
            animation: 'breathingGlow 3s ease-in-out infinite',
            position: 'relative',
            zIndex: 1,
          },
        },
        standardSuccess: {
          borderColor: `${jellyfishColors.successSoft}40`,
        },
        standardWarning: {
          borderColor: `${jellyfishColors.warningSoft}40`,
        },
        standardError: {
          borderColor: `${jellyfishColors.errorSoft}40`,
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          background: `linear-gradient(135deg, ${jellyfishColors.backgroundPaper}CC, ${jellyfishColors.backgroundLight}CC)`,
          backdropFilter: 'blur(40px) saturate(180%)',
          borderRadius: 28,
          border: `1px solid ${jellyfishColors.frameViolet}`,
          boxShadow: `
            0 24px 80px rgba(236, 64, 122, 0.22),
            0 8px 24px rgba(171, 71, 188, 0.18),
            inset 0 1px 0 rgba(255, 255, 255, 0.8)
          `,
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          animation: 'depthPulse 12s ease-in-out infinite',
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: '-50%',
            left: '-50%',
            width: '200%',
            height: '200%',
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}12 0%, ${jellyfishColors.violetSoft}08 50%, transparent 100%)`,
            animation: 'etherealGlow 10s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
      },
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          background: 'transparent',
          borderBottom: 'none',
          padding: '24px 28px 20px',
          color: '#f0f0f0',
        },
      },
    },
    MuiDialogContent: {
      styleOverrides: {
        root: {
          background: 'transparent',
          padding: '0 28px 24px',
        },
      },
    },
    MuiDialogActions: {
      styleOverrides: {
        root: {
          background: 'transparent',
          borderTop: 'none',
          padding: '20px 28px 24px',
          gap: 12,
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 16,
            background: `linear-gradient(135deg, ${jellyfishColors.backgroundPaper}CC, ${jellyfishColors.backgroundLight}CC)`,
            backdropFilter: 'blur(20px) saturate(180%)',
            border: `1px solid ${jellyfishColors.frameViolet}`,
            boxShadow: `
              0 4px 16px rgba(236, 64, 122, 0.12),
              inset 0 1px 0 rgba(255, 255, 255, 0.8)
            `,
            transition: 'all 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
            '& fieldset': {
              border: 'none',
            },
            '&:hover': {
              boxShadow: `
                0 6px 20px rgba(236, 64, 122, 0.2),
                0 0 0 1px rgba(171, 71, 188, 0.3),
                inset 0 1px 0 rgba(255, 255, 255, 0.9)
              `,
            },
            '&.Mui-focused': {
              boxShadow: `
                0 8px 24px rgba(236, 64, 122, 0.25),
                0 0 0 2px rgba(171, 71, 188, 0.35),
                inset 0 1px 0 rgba(255, 255, 255, 0.9)
              `,
            },
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 12,
          fontWeight: 500,
          fontSize: '0.8125rem',
          backdropFilter: 'blur(10px)',
          border: 'none',
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.2)',
        },
        colorPrimary: {
          background: `linear-gradient(135deg, ${jellyfishColors.cyanSubtle}, ${jellyfishColors.violetSubtle})`,
          color: jellyfishColors.cyanSoft,
          boxShadow: '0 4px 12px rgba(0, 229, 255, 0.2)',
        },
        colorSecondary: {
          backgroundColor: jellyfishColors.violetSubtle,
          color: jellyfishColors.violetSoft,
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        root: {
          borderBottom: 'none',
          position: 'relative',
          '&::after': {
            content: '""',
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: '1px',
            background: `linear-gradient(90deg, transparent, ${jellyfishColors.cyanSoft}30, ${jellyfishColors.magenta}30, transparent)`,
            opacity: 0.5,
          },
        },
        indicator: {
          background: `linear-gradient(90deg, ${jellyfishColors.cyanSoft}, ${jellyfishColors.magenta})`,
          height: 3,
          borderRadius: '3px 3px 0 0',
          boxShadow: `
            0 0 12px ${jellyfishColors.cyanSoft}70,
            0 0 8px ${jellyfishColors.magenta}60,
            0 0 4px ${jellyfishColors.cyanSoft}40
          `,
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
          transition: 'all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          '&:hover': {
            color: jellyfishColors.cyanSoft,
          },
          '&.Mui-selected': {
            background: `linear-gradient(135deg, ${jellyfishColors.cyanSoft}25, ${jellyfishColors.magenta}22)`,
            color: '#f0f0f0',
            fontWeight: 600,
            textShadow: 'none',
            borderRadius: '8px 8px 0 0',
          },
        },
      },
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: 28,
          border: `1px solid ${jellyfishColors.frameViolet}`,
          background: `linear-gradient(135deg, ${jellyfishColors.backgroundPaper}CC, ${jellyfishColors.backgroundLight}CC)`,
          backdropFilter: 'blur(35px) saturate(200%)',
          boxShadow: `
            0 8px 32px rgba(236, 64, 122, 0.18),
            inset 0 1px 0 rgba(255, 255, 255, 0.8)
          `,
          overflow: 'hidden',
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          animation: 'depthPulse 12s ease-in-out infinite',
          position: 'relative',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: '-50%',
            left: '-50%',
            width: '200%',
            height: '200%',
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}12 0%, ${jellyfishColors.violetSoft}08 50%, transparent 70%)`,
            animation: 'etherealGlow 10s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-head': {
            background: `linear-gradient(135deg, rgba(0, 229, 255, 0.15), rgba(171, 71, 188, 0.12))`,
            color: '#e8e8e8',
            fontWeight: 600,
            fontSize: '0.875rem',
            borderBottom: 'none',
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
          transition: 'all 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          position: 'relative',
          zIndex: 1,
          '&:hover': {
            background: `linear-gradient(135deg, ${jellyfishColors.cyanSubtle}, ${jellyfishColors.violetSubtle})`,
            boxShadow: '0 2px 8px rgba(0, 229, 255, 0.08)',
          },
          '& .MuiTableCell-body': {
            borderBottom: 'none',
            padding: '20px 24px',
            transition: 'all 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          },
        },
      },
    },
  },
})
