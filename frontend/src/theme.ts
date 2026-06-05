import { createTheme } from '@mui/material/styles'

// ── Palette Obsidian ──────────────────────────────────────────────────────────
// Inspirée de Linear, Vercel, Grafana.
// Un seul accent (bleu #4F8EF7), fonds profonds, typographie lisible.
export const kuraColors = {
  // Fonds
  bg0: '#0C0E14',   // arrière-plan global
  bg1: '#12141C',   // surfaces primaires (sidebar, paper)
  bg2: '#1A1D28',   // cartes, panneaux
  bg3: '#21253A',   // hover, selected

  // Accent principal — bleu confiant
  accent:      '#4F8EF7',
  accentHover: '#6AA0F8',
  accentMuted: 'rgba(79,142,247,0.12)',
  accentSubtle:'rgba(79,142,247,0.06)',

  // Bordures
  border0: 'rgba(255,255,255,0.05)',  // très subtile
  border1: 'rgba(255,255,255,0.09)',  // défaut
  border2: 'rgba(255,255,255,0.16)',  // emphasis

  // Texte
  text0: '#F1F3F9',   // primaire
  text1: '#B8BFCC',   // secondaire
  text2: '#6B7385',   // muted

  // Statuts sémantiques
  success: '#34D399',
  successBg: 'rgba(52,211,153,0.10)',
  warning: '#FBBF24',
  warningBg: 'rgba(251,191,36,0.10)',
  error:   '#F87171',
  errorBg: 'rgba(248,113,113,0.10)',
  info:    '#60A5FA',
  infoBg:  'rgba(96,165,250,0.10)',
}

// Rétrocompatibilité — les pages qui importent encore jellyfishColors
export const jellyfishColors = {
  backgroundLight:   kuraColors.bg1,
  backgroundPaper:   kuraColors.bg2,
  backgroundCard:    kuraColors.bg2,
  cyanSoft:          kuraColors.accent,
  cyanMedium:        kuraColors.accentHover,
  cyanLight:         kuraColors.accentHover,
  cyanDeep:          kuraColors.accent,
  cyanSubtle:        kuraColors.accentMuted,
  violetSoft:        '#A78BFA',
  violetMedium:      '#7C3AED',
  violetDeep:        '#5B21B6',
  violetLight:       '#DDD6FE',
  magenta:           kuraColors.error,
  magentaDeep:       kuraColors.error,
  fuchsia:           '#C084FC',
  violetRed:         '#A855F7',
  violetSubtle:      'rgba(167,139,250,0.15)',
  frameViolet:       'rgba(124,58,237,0.4)',
  frameVioletRed:    'rgba(248,113,113,0.3)',
  magentaSoft:       'rgba(248,113,113,0.2)',
  grayLight:         kuraColors.text2,
  grayMedium:        kuraColors.text2,
  grayDark:          kuraColors.text2,
  graySubtle:        kuraColors.border0,
  successSoft:       kuraColors.success,
  warningSoft:       kuraColors.warning,
  errorSoft:         kuraColors.error,
  infoSoft:          kuraColors.info,
}

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main:  kuraColors.accent,
      light: kuraColors.accentHover,
      dark:  kuraColors.accent,
    },
    secondary: {
      main: '#A78BFA',
    },
    background: {
      default: kuraColors.bg0,
      paper:   kuraColors.bg1,
    },
    text: {
      primary:   kuraColors.text0,
      secondary: kuraColors.text1,
    },
    error:   { main: kuraColors.error },
    warning: { main: kuraColors.warning },
    success: { main: kuraColors.success },
    info:    { main: kuraColors.info },
    divider: kuraColors.border1,
  },

  typography: {
    fontFamily: [
      '"Inter"',
      '-apple-system',
      'BlinkMacSystemFont',
      '"Segoe UI"',
      'sans-serif',
    ].join(','),
    h1: { fontSize: '2.25rem', fontWeight: 600, letterSpacing: '-0.03em', lineHeight: 1.2 },
    h2: { fontSize: '1.875rem', fontWeight: 600, letterSpacing: '-0.025em', lineHeight: 1.25 },
    h3: { fontSize: '1.5rem',   fontWeight: 600, letterSpacing: '-0.02em',  lineHeight: 1.3 },
    h4: { fontSize: '1.25rem',  fontWeight: 600, letterSpacing: '-0.015em', lineHeight: 1.35 },
    h5: { fontSize: '1.125rem', fontWeight: 500, lineHeight: 1.4 },
    h6: { fontSize: '1rem',     fontWeight: 500, lineHeight: 1.5 },
    body1: { fontSize: '0.9375rem', lineHeight: 1.65, fontWeight: 400 },
    body2: { fontSize: '0.875rem',  lineHeight: 1.6,  fontWeight: 400, color: kuraColors.text1 },
    caption: { fontSize: '0.75rem', lineHeight: 1.5, color: kuraColors.text2 },
    button: { textTransform: 'none', fontWeight: 500, fontSize: '0.875rem' },
  },

  shape: { borderRadius: 6 },
  spacing: 8,

  components: {
    // ── Boutons ─────────────────────────────────────────────────────────────
    MuiButton: {
      defaultProps: { disableElevation: true },
      styleOverrides: {
        root: {
          borderRadius: 6,
          fontWeight: 500,
          fontSize: '0.875rem',
          padding: '7px 16px',
          transition: 'background 0.15s ease, border-color 0.15s ease, color 0.15s ease',
          boxShadow: 'none',
          '&:hover': { boxShadow: 'none' },
        },
        contained: {
          background: kuraColors.accent,
          color: '#fff',
          '&:hover': { background: kuraColors.accentHover },
        },
        outlined: {
          borderColor: kuraColors.border2,
          color: kuraColors.text0,
          '&:hover': {
            borderColor: kuraColors.accent,
            background: kuraColors.accentSubtle,
            color: kuraColors.accent,
          },
        },
        text: {
          color: kuraColors.text1,
          '&:hover': { background: kuraColors.bg3, color: kuraColors.text0 },
        },
      },
    },

    // ── Cartes ───────────────────────────────────────────────────────────────
    MuiCard: {
      styleOverrides: {
        root: {
          background: kuraColors.bg2,
          border: `1px solid ${kuraColors.border1}`,
          borderRadius: 8,
          boxShadow: '0 1px 3px rgba(0,0,0,0.3), 0 1px 2px rgba(0,0,0,0.2)',
          transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
          '&:hover': {
            borderColor: kuraColors.border2,
          },
        },
      },
    },

    // ── Paper ────────────────────────────────────────────────────────────────
    MuiPaper: {
      styleOverrides: {
        root: {
          background: kuraColors.bg1,
          border: `1px solid ${kuraColors.border1}`,
          borderRadius: 8,
          boxShadow: 'none',
          '&.MuiDrawer-paper': {
            background: `${kuraColors.bg1} !important`,
            backgroundImage: 'none !important',
            border: 'none !important',
          },
        },
        elevation1: {
          boxShadow: '0 1px 3px rgba(0,0,0,0.3)',
        },
      },
    },

    // ── Drawer (sidebar) ─────────────────────────────────────────────────────
    MuiDrawer: {
      styleOverrides: {
        paper: {
          background: kuraColors.bg1,
          borderRight: `1px solid ${kuraColors.border1}`,
          borderRadius: 0,
          boxShadow: 'none',
        },
      },
    },

    // ── AppBar ───────────────────────────────────────────────────────────────
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: kuraColors.bg1,
          borderBottom: `1px solid ${kuraColors.border1}`,
          borderRadius: 0,
          boxShadow: 'none',
        },
      },
    },

    // ── Navigation items ─────────────────────────────────────────────────────
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          margin: '1px 8px',
          padding: '8px 12px',
          transition: 'background 0.12s ease',
          '&:hover': {
            background: kuraColors.bg3,
          },
          '&.Mui-selected': {
            background: kuraColors.accentMuted,
            '& .MuiListItemIcon-root': { color: kuraColors.accent },
            '& .MuiListItemText-primary': { color: kuraColors.text0, fontWeight: 500 },
            '&:hover': { background: kuraColors.accentMuted },
          },
        },
      },
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: kuraColors.text2,
          minWidth: 36,
          transition: 'color 0.12s ease',
        },
      },
    },
    MuiListItemText: {
      styleOverrides: {
        primary: {
          fontSize: '0.875rem',
          fontWeight: 400,
          color: kuraColors.text1,
        },
      },
    },

    // ── Alertes ──────────────────────────────────────────────────────────────
    MuiAlert: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          border: `1px solid`,
          fontSize: '0.875rem',
          padding: '10px 16px',
          boxShadow: 'none',
        },
        standardInfo: {
          background: kuraColors.infoBg,
          borderColor: `rgba(96,165,250,0.25)`,
          color: kuraColors.text0,
          '& .MuiAlert-icon': { color: kuraColors.info },
        },
        standardSuccess: {
          background: kuraColors.successBg,
          borderColor: `rgba(52,211,153,0.25)`,
          '& .MuiAlert-icon': { color: kuraColors.success },
        },
        standardWarning: {
          background: kuraColors.warningBg,
          borderColor: `rgba(251,191,36,0.25)`,
          '& .MuiAlert-icon': { color: kuraColors.warning },
        },
        standardError: {
          background: kuraColors.errorBg,
          borderColor: `rgba(248,113,113,0.25)`,
          '& .MuiAlert-icon': { color: kuraColors.error },
        },
      },
    },

    // ── Dialogs ──────────────────────────────────────────────────────────────
    MuiDialog: {
      styleOverrides: {
        paper: {
          background: kuraColors.bg2,
          borderRadius: 10,
          border: `1px solid ${kuraColors.border2}`,
          boxShadow: '0 25px 50px rgba(0,0,0,0.5)',
        },
      },
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          padding: '20px 24px 16px',
          fontSize: '1.0625rem',
          fontWeight: 600,
          color: kuraColors.text0,
          borderBottom: `1px solid ${kuraColors.border1}`,
        },
      },
    },
    MuiDialogContent: {
      styleOverrides: {
        root: { padding: '20px 24px' },
      },
    },
    MuiDialogActions: {
      styleOverrides: {
        root: {
          padding: '16px 24px',
          gap: 8,
          borderTop: `1px solid ${kuraColors.border1}`,
        },
      },
    },

    // ── Inputs ───────────────────────────────────────────────────────────────
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 6,
            fontSize: '0.875rem',
            '& fieldset': {
              borderColor: kuraColors.border2,
              transition: 'border-color 0.15s ease',
            },
            '&:hover fieldset': {
              borderColor: kuraColors.border2,
            },
            '&.Mui-focused fieldset': {
              borderColor: kuraColors.accent,
              borderWidth: 1,
            },
          },
          '& .MuiInputLabel-root': {
            fontSize: '0.875rem',
            color: kuraColors.text2,
            '&.Mui-focused': { color: kuraColors.accent },
            '&.MuiInputLabel-shrink': {
              paddingLeft: '4px',
              paddingRight: '4px',
            },
          },
        },
      },
    },
    MuiSelect: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          fontSize: '0.875rem',
        },
      },
    },

    // ── Chips / badges ───────────────────────────────────────────────────────
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 4,
          fontWeight: 500,
          fontSize: '0.75rem',
          height: 22,
          letterSpacing: '0.01em',
        },
        outlined: {
          borderColor: kuraColors.border2,
          color: kuraColors.text1,
        },
        filled: {
          background: kuraColors.bg3,
          color: kuraColors.text0,
        },
      },
    },

    // ── Tabs ─────────────────────────────────────────────────────────────────
    MuiTabs: {
      styleOverrides: {
        root: {
          borderBottom: `1px solid ${kuraColors.border1}`,
          minHeight: 42,
        },
        indicator: {
          background: kuraColors.accent,
          height: 2,
          borderRadius: 1,
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 400,
          fontSize: '0.875rem',
          color: kuraColors.text2,
          minHeight: 42,
          padding: '8px 16px',
          '&:hover': { color: kuraColors.text1 },
          '&.Mui-selected': { color: kuraColors.text0, fontWeight: 500 },
        },
      },
    },

    // ── Tables ───────────────────────────────────────────────────────────────
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          border: `1px solid ${kuraColors.border1}`,
          background: kuraColors.bg2,
          boxShadow: 'none',
        },
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-head': {
            background: kuraColors.bg1,
            color: kuraColors.text2,
            fontWeight: 500,
            fontSize: '0.75rem',
            textTransform: 'uppercase',
            letterSpacing: '0.06em',
            borderBottom: `1px solid ${kuraColors.border1}`,
            padding: '10px 16px',
          },
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:hover': { background: kuraColors.bg3 },
          '& .MuiTableCell-body': {
            borderBottom: `1px solid ${kuraColors.border0}`,
            padding: '12px 16px',
            fontSize: '0.875rem',
            color: kuraColors.text0,
          },
        },
      },
    },

    // ── Tooltip ──────────────────────────────────────────────────────────────
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          background: kuraColors.bg3,
          border: `1px solid ${kuraColors.border2}`,
          color: kuraColors.text0,
          fontSize: '0.75rem',
          borderRadius: 4,
          boxShadow: '0 4px 12px rgba(0,0,0,0.4)',
        },
      },
    },

    // ── Divider ──────────────────────────────────────────────────────────────
    MuiDivider: {
      styleOverrides: {
        root: { borderColor: kuraColors.border1 },
      },
    },

    // ── Menu ─────────────────────────────────────────────────────────────────
    MuiMenu: {
      styleOverrides: {
        paper: {
          background: kuraColors.bg2,
          border: `1px solid ${kuraColors.border2}`,
          borderRadius: 8,
          boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
        },
      },
    },
    MuiMenuItem: {
      styleOverrides: {
        root: {
          fontSize: '0.875rem',
          padding: '8px 12px',
          borderRadius: 4,
          margin: '1px 4px',
          '&:hover': { background: kuraColors.bg3 },
          '&.Mui-selected': {
            background: kuraColors.accentMuted,
            '&:hover': { background: kuraColors.accentMuted },
          },
        },
      },
    },
  },
})
