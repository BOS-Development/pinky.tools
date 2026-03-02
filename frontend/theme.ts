'use client';
import { createTheme } from '@mui/material/styles';

// Neocom Dark design tokens — three-layer depth system
const neocom = {
  cyan: '#00d4ff',
  cyanLight: '#33ddff',
  cyanDark: '#00a8cc',
  bgVoid: '#0a0a0f',
  bgPanel: '#12141a',
  surfaceElevated: '#1a1d24',
  borderCyan: 'rgba(0, 212, 255, 0.10)',
  borderCyanHover: 'rgba(0, 212, 255, 0.30)',
  glowCyanSm: '0 0 8px rgba(0, 212, 255, 0.25)',
  glowCyanMd: '0 0 12px rgba(0, 212, 255, 0.35)',
};

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: neocom.cyan,
      light: neocom.cyanLight,
      dark: neocom.cyanDark,
    },
    secondary: {
      main: '#8b5cf6',
      light: '#a78bfa',
      dark: '#7c3aed',
    },
    success: {
      main: '#2dd4bf',
      light: '#5eead4',
      dark: '#14b8a6',
    },
    error: {
      main: '#f43f5e',
      light: '#fb7185',
      dark: '#e11d48',
    },
    warning: {
      main: '#fbbf24',
      light: '#fde68a',
      dark: '#f59e0b',
    },
    background: {
      default: neocom.bgVoid,
      paper: neocom.bgPanel,
    },
    text: {
      primary: '#f1f5f9',
      secondary: '#94a3b8',
    },
    divider: 'rgba(0, 212, 255, 0.08)',
  },
  typography: {
    fontFamily: 'var(--font-geist-sans), -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    h1: {
      fontSize: '1.75rem',
      fontWeight: 700,
      letterSpacing: '-0.02em',
    },
    h2: {
      fontSize: '1.5rem',
      fontWeight: 700,
      letterSpacing: '-0.01em',
    },
    h3: {
      fontSize: '1.375rem',
      fontWeight: 600,
      letterSpacing: '-0.01em',
    },
    h4: {
      fontSize: '1.25rem',
      fontWeight: 600,
      letterSpacing: '-0.01em',
    },
    h5: {
      fontSize: '1.125rem',
      fontWeight: 600,
    },
    h6: {
      fontSize: '1rem',
      fontWeight: 600,
    },
    body1: {
      fontSize: '0.9375rem',
      lineHeight: 1.6,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.5,
    },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          scrollbarColor: `${neocom.surfaceElevated} ${neocom.bgVoid}`,
          '&::-webkit-scrollbar': {
            width: '8px',
            height: '8px',
          },
          '&::-webkit-scrollbar-track': {
            background: neocom.bgVoid,
          },
          '&::-webkit-scrollbar-thumb': {
            background: neocom.surfaceElevated,
            borderRadius: '2px',
            '&:hover': {
              background: '#334155',
            },
          },
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          textTransform: 'none',
          fontWeight: 500,
          fontSize: '0.875rem',
          padding: '8px 16px',
        },
        contained: {
          boxShadow: neocom.glowCyanSm,
          '&:hover': {
            boxShadow: neocom.glowCyanMd,
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          backgroundImage: 'none',
          border: `1px solid ${neocom.borderCyan}`,
          '&:hover': {
            borderColor: neocom.borderCyanHover,
            boxShadow: neocom.glowCyanSm,
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          borderRadius: 2,
        },
        outlined: {
          border: `1px solid ${neocom.borderCyan}`,
        },
      },
    },
    MuiTable: {
      styleOverrides: {
        root: {
          borderCollapse: 'separate',
          borderSpacing: 0,
        },
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-root': {
            backgroundColor: neocom.bgPanel,
            color: neocom.cyan,
            fontWeight: 600,
            fontSize: '0.75rem',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            borderBottom: `1px solid ${neocom.borderCyanHover}`,
            padding: '12px 16px',
          },
        },
      },
    },
    MuiTableBody: {
      styleOverrides: {
        root: {
          '& .MuiTableRow-root': {
            '&:hover': {
              backgroundColor: 'rgba(0, 212, 255, 0.04)',
            },
            '& .MuiTableCell-root': {
              borderBottom: `1px solid ${neocom.borderCyan}`,
              padding: '12px 16px',
              fontSize: '0.875rem',
            },
          },
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderColor: neocom.borderCyan,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          fontWeight: 500,
          fontSize: '0.75rem',
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
          fontSize: '0.875rem',
          minHeight: 48,
          '&.Mui-selected': {
            color: neocom.cyan,
            textShadow: '0 0 8px rgba(0, 212, 255, 0.4)',
          },
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        indicator: {
          backgroundColor: neocom.cyan,
          height: 2,
          boxShadow: '0 0 8px rgba(0, 212, 255, 0.5)',
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 2,
            '& fieldset': {
              borderColor: neocom.borderCyan,
            },
            '&:hover fieldset': {
              borderColor: neocom.borderCyanHover,
            },
            '&.Mui-focused fieldset': {
              borderColor: neocom.cyan,
              boxShadow: neocom.glowCyanSm,
            },
          },
        },
      },
    },
    MuiSelect: {
      styleOverrides: {
        root: {
          borderRadius: 2,
        },
      },
    },
  },
});

export default theme;
