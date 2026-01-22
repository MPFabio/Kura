import { createTheme } from '@mui/material/styles'

// Palette "Jellyfish Serenity" - Minimaliste et calme
export const jellyfishColors = {
  // Fonds doux et apaisants
  deepOcean: '#0F1419',
  oceanDark: '#1A1F2E',
  oceanMedium: '#252B3A',
  
  // Cyan de la méduse - versions plus vives mais harmonieuses
  cyanSoft: '#00E5FF',      // Cyan vif principal (plus lumineux)
  cyanMedium: '#00B8D4',    // Cyan moyen
  cyanLight: '#4DD0E1',     // Cyan clair pour accents
  cyanSubtle: 'rgba(0, 229, 255, 0.2)', // Cyan subtil pour backgrounds
  
  // Violet de la méduse - versions plus vives mais harmonieuses
  violetSoft: '#B388FF',    // Violet vif principal (plus lumineux)
  violetMedium: '#9C27B0',  // Violet moyen
  violetLight: '#CE93D8',   // Violet clair
  violetSubtle: 'rgba(179, 136, 255, 0.2)', // Violet subtil
  
  // Gris neutres et doux
  grayLight: '#B0BEC5',     // Gris clair pour texte secondaire
  grayMedium: '#78909C',    // Gris moyen
  grayDark: '#546E7A',      // Gris foncé
  graySubtle: 'rgba(176, 190, 197, 0.1)', // Gris très subtil
  
  // Alertes douces
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
      default: jellyfishColors.deepOcean,
      paper: jellyfishColors.oceanDark,
    },
    text: {
      primary: 'rgba(255, 255, 255, 0.95)',
      secondary: jellyfishColors.grayLight,
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
      color: 'rgba(255, 255, 255, 0.95)',
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 500,
      letterSpacing: '-0.01em',
      color: 'rgba(255, 255, 255, 0.95)',
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 500,
      letterSpacing: '-0.01em',
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
      fontSize: '0.9375rem',
      lineHeight: 1.6,
      color: 'rgba(255, 255, 255, 0.85)',
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.5,
      color: 'rgba(255, 255, 255, 0.75)',
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
          border: `1px solid ${jellyfishColors.cyanSoft}40`,
          color: jellyfishColors.cyanSoft,
          background: 'transparent',
          '&:hover': {
            borderColor: `${jellyfishColors.cyanSoft}80`,
            background: jellyfishColors.cyanSubtle,
            boxShadow: `0 2px 8px ${jellyfishColors.cyanSoft}20`,
          },
        },
        contained: {
          background: `linear-gradient(135deg, ${jellyfishColors.cyanSoft}, ${jellyfishColors.violetSoft})`,
          color: '#FFFFFF',
          border: 'none',
          boxShadow: `0 4px 16px ${jellyfishColors.cyanSoft}30`,
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          '&:hover': {
            background: `linear-gradient(135deg, ${jellyfishColors.cyanMedium}, ${jellyfishColors.violetMedium})`,
            boxShadow: `0 6px 20px ${jellyfishColors.cyanSoft}35, 0 0 16px ${jellyfishColors.violetSoft}25`,
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
          background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          border: 'none',
          borderRadius: 24,
          boxShadow: `
            0 8px 32px rgba(0, 0, 0, 0.3),
            0 2px 8px rgba(0, 0, 0, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.08)
          `,
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          position: 'relative',
          overflow: 'hidden',
          animation: 'jellyfishFloat 8s ease-in-out infinite',
          animationDelay: 'var(--card-delay, 0s)',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: '-50%',
            left: '-50%',
            width: '200%',
            height: '200%',
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}15 0%, transparent 70%)`,
            animation: 'etherealGlow 6s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: '2px',
            background: `linear-gradient(90deg, transparent, ${jellyfishColors.cyanSoft}50, ${jellyfishColors.violetSoft}50, transparent)`,
            opacity: 0,
            transition: 'opacity 0.6s ease',
            pointerEvents: 'none',
            zIndex: 1,
          },
          '&:hover': {
            boxShadow: `
              0 12px 48px rgba(0, 229, 255, 0.2),
              0 4px 16px rgba(179, 136, 255, 0.15),
              0 2px 8px rgba(0, 0, 0, 0.3),
              inset 0 1px 0 rgba(255, 255, 255, 0.1)
            `,
            '&::after': {
              opacity: 1,
            },
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          border: 'none',
          borderRadius: 24,
          boxShadow: `
            0 8px 32px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.08)
          `,
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          animation: 'depthPulse 10s ease-in-out infinite',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          background: `linear-gradient(180deg, ${jellyfishColors.deepOcean}CC, ${jellyfishColors.oceanDark}CC)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          borderRight: 'none',
          boxShadow: `
            4px 0 24px rgba(0, 0, 0, 0.4),
            inset -1px 0 0 rgba(255, 255, 255, 0.05)
          `,
          position: 'relative',
          transition: 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            right: 0,
            width: '2px',
            height: '100%',
            background: `linear-gradient(180deg, transparent, ${jellyfishColors.cyanSoft}30, ${jellyfishColors.violetSoft}30, transparent)`,
            pointerEvents: 'none',
            opacity: 0.6,
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            top: '-50%',
            right: '-50%',
            width: '200%',
            height: '200%',
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}08 0%, ${jellyfishColors.violetSoft}06 50%, transparent 100%)`,
            animation: 'etherealGlow 12s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: `${jellyfishColors.oceanDark}E6`,
          backdropFilter: 'blur(20px)',
          borderBottom: `1px solid ${jellyfishColors.graySubtle}`,
          boxShadow: '0 1px 4px rgba(0, 0, 0, 0.1)',
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
            background: `linear-gradient(90deg, ${jellyfishColors.cyanSubtle}, ${jellyfishColors.violetSubtle})`,
            borderLeft: `3px solid ${jellyfishColors.cyanSoft}`,
            boxShadow: `
              0 0 8px ${jellyfishColors.cyanSoft}30,
              0 0 6px ${jellyfishColors.violetSoft}20,
              inset 3px 0 0 ${jellyfishColors.violetSoft}40
            `,
            position: 'relative',
            '&::before': {
              content: '""',
              position: 'absolute',
              left: 0,
              top: 0,
              bottom: 0,
              width: '3px',
              background: `linear-gradient(180deg, ${jellyfishColors.cyanSoft}, ${jellyfishColors.violetSoft})`,
              zIndex: 1,
            },
            '&:hover': {
              background: `linear-gradient(90deg, ${jellyfishColors.cyanSoft}30, ${jellyfishColors.violetSoft}25)`,
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
    MuiAlert: {
      styleOverrides: {
        root: {
          background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          border: 'none',
          borderRadius: 20,
          boxShadow: `
            0 8px 32px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.08)
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
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}10 0%, transparent 70%)`,
            animation: 'etherealGlow 8s ease-in-out infinite',
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
        standardInfo: {
          background: `linear-gradient(135deg, rgba(0, 229, 255, 0.12), rgba(179, 136, 255, 0.08))`,
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
          background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
          backdropFilter: 'blur(40px) saturate(180%)',
          borderRadius: 28,
          border: 'none',
          boxShadow: `
            0 24px 80px rgba(0, 0, 0, 0.6),
            inset 0 1px 0 rgba(255, 255, 255, 0.1)
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
          color: 'rgba(255, 255, 255, 0.95)',
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
            background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
            backdropFilter: 'blur(20px) saturate(180%)',
            border: 'none',
            boxShadow: `
              0 4px 16px rgba(0, 0, 0, 0.2),
              inset 0 1px 0 rgba(255, 255, 255, 0.08)
            `,
            transition: 'all 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
            '& fieldset': {
              border: 'none',
            },
            '&:hover': {
              boxShadow: `
                0 6px 20px rgba(0, 229, 255, 0.15),
                inset 0 1px 0 rgba(255, 255, 255, 0.1)
              `,
            },
            '&.Mui-focused': {
              boxShadow: `
                0 8px 24px rgba(0, 229, 255, 0.25),
                0 0 0 2px rgba(0, 229, 255, 0.15),
                inset 0 1px 0 rgba(255, 255, 255, 0.12)
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
            background: `linear-gradient(90deg, transparent, ${jellyfishColors.cyanSoft}30, ${jellyfishColors.violetSoft}30, transparent)`,
            opacity: 0.5,
          },
        },
        indicator: {
          background: `linear-gradient(90deg, ${jellyfishColors.cyanSoft}, ${jellyfishColors.violetSoft})`,
          height: 3,
          borderRadius: '3px 3px 0 0',
          boxShadow: `0 0 8px ${jellyfishColors.cyanSoft}50`,
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
          fontSize: '0.9375rem',
          color: jellyfishColors.grayLight,
          transition: 'all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94)',
          '&:hover': {
            color: jellyfishColors.cyanSoft,
          },
          '&.Mui-selected': {
            color: jellyfishColors.cyanSoft,
            textShadow: `0 0 8px ${jellyfishColors.cyanSoft}50`,
          },
        },
      },
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: 24,
          border: 'none',
          background: `linear-gradient(135deg, ${jellyfishColors.oceanDark}CC, ${jellyfishColors.deepOcean}CC)`,
          backdropFilter: 'blur(30px) saturate(180%)',
          boxShadow: `
            0 8px 32px rgba(0, 0, 0, 0.3),
            inset 0 1px 0 rgba(255, 255, 255, 0.08)
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
            background: `radial-gradient(circle, ${jellyfishColors.cyanSoft}08 0%, transparent 70%)`,
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
            background: `linear-gradient(135deg, ${jellyfishColors.oceanMedium}80, ${jellyfishColors.oceanDark}80)`,
            color: jellyfishColors.grayLight,
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
