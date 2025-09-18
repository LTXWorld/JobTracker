// Material Design 3 Color System
export const theme = {
  colors: {
    // Primary colors
    primary: '#6750A4',
    onPrimary: '#FFFFFF',
    primaryContainer: '#EADDFF',
    onPrimaryContainer: '#21005D',

    // Secondary colors
    secondary: '#625B71',
    onSecondary: '#FFFFFF',
    secondaryContainer: '#E8DEF8',
    onSecondaryContainer: '#1D192B',

    // Tertiary colors
    tertiary: '#7D5260',
    onTertiary: '#FFFFFF',
    tertiaryContainer: '#FFD8E4',
    onTertiaryContainer: '#31111D',

    // Error colors
    error: '#B3261E',
    onError: '#FFFFFF',
    errorContainer: '#F9DEDC',
    onErrorContainer: '#410E0B',

    // Background colors
    background: '#FEF7FF',
    onBackground: '#1C1B1F',
    surface: '#FFFFFF',
    onSurface: '#1C1B1F',
    surfaceVariant: '#E7E0EC',
    onSurfaceVariant: '#49454F',

    // Outline colors
    outline: '#79747E',
    outlineVariant: '#CAC4D0',

    // Other colors
    shadow: '#000000',
    scrim: '#000000',
    inverseSurface: '#313033',
    inverseOnSurface: '#F4EFF4',
    inversePrimary: '#D0BCFF',

    // Success colors (custom)
    success: '#2E7D32',
    onSuccess: '#FFFFFF',
    successContainer: '#E8F5E8',
    onSuccessContainer: '#1B5E20',

    // Warning colors (custom)
    warning: '#F57C00',
    onWarning: '#FFFFFF',
    warningContainer: '#FFF3E0',
    onWarningContainer: '#E65100',

    // Info colors (custom)
    info: '#1976D2',
    onInfo: '#FFFFFF',
    infoContainer: '#E3F2FD',
    onInfoContainer: '#0D47A1',
  },

  // Typography scale
  typography: {
    displayLarge: {
      fontSize: 57,
      lineHeight: 64,
      fontWeight: '400' as const,
    },
    displayMedium: {
      fontSize: 45,
      lineHeight: 52,
      fontWeight: '400' as const,
    },
    displaySmall: {
      fontSize: 36,
      lineHeight: 44,
      fontWeight: '400' as const,
    },
    headlineLarge: {
      fontSize: 32,
      lineHeight: 40,
      fontWeight: '600' as const,
    },
    headlineMedium: {
      fontSize: 28,
      lineHeight: 36,
      fontWeight: '600' as const,
    },
    headlineSmall: {
      fontSize: 24,
      lineHeight: 32,
      fontWeight: '600' as const,
    },
    titleLarge: {
      fontSize: 22,
      lineHeight: 28,
      fontWeight: '600' as const,
    },
    titleMedium: {
      fontSize: 16,
      lineHeight: 24,
      fontWeight: '500' as const,
    },
    titleSmall: {
      fontSize: 14,
      lineHeight: 20,
      fontWeight: '500' as const,
    },
    labelLarge: {
      fontSize: 14,
      lineHeight: 20,
      fontWeight: '500' as const,
    },
    labelMedium: {
      fontSize: 12,
      lineHeight: 16,
      fontWeight: '500' as const,
    },
    labelSmall: {
      fontSize: 11,
      lineHeight: 16,
      fontWeight: '500' as const,
    },
    bodyLarge: {
      fontSize: 16,
      lineHeight: 24,
      fontWeight: '400' as const,
    },
    bodyMedium: {
      fontSize: 14,
      lineHeight: 20,
      fontWeight: '400' as const,
    },
    bodySmall: {
      fontSize: 12,
      lineHeight: 16,
      fontWeight: '400' as const,
    },
  },

  // Spacing scale
  spacing: {
    none: 0,
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
    xxxl: 32,
    xxxxl: 40,
    xxxxxl: 48,
  },

  // Border radius scale
  borderRadius: {
    none: 0,
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
    full: 1000,
  },

  // Elevation/Shadow system
  elevation: {
    none: {
      elevation: 0,
      shadowOffset: { width: 0, height: 0 },
      shadowOpacity: 0,
      shadowRadius: 0,
    },
    level1: {
      elevation: 1,
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.2,
      shadowRadius: 1.41,
    },
    level2: {
      elevation: 2,
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.22,
      shadowRadius: 2.22,
    },
    level3: {
      elevation: 3,
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.24,
      shadowRadius: 3.84,
    },
    level4: {
      elevation: 4,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.25,
      shadowRadius: 3.84,
    },
    level5: {
      elevation: 5,
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.27,
      shadowRadius: 4.65,
    },
  },

  // Component sizing
  sizing: {
    button: {
      small: { height: 32, paddingHorizontal: 16 },
      medium: { height: 40, paddingHorizontal: 24 },
      large: { height: 48, paddingHorizontal: 32 },
    },
    input: {
      small: { height: 40 },
      medium: { height: 48 },
      large: { height: 56 },
    },
    icon: {
      small: 16,
      medium: 20,
      large: 24,
      xlarge: 32,
    },
  },

  // Animation durations
  animation: {
    fast: 150,
    normal: 300,
    slow: 500,
  },
};

// Dark theme variant
export const darkTheme = {
  ...theme,
  colors: {
    ...theme.colors,
    // Primary colors
    primary: '#D0BCFF',
    onPrimary: '#381E72',
    primaryContainer: '#4F378B',
    onPrimaryContainer: '#EADDFF',

    // Secondary colors
    secondary: '#CCC2DC',
    onSecondary: '#332D41',
    secondaryContainer: '#4A4458',
    onSecondaryContainer: '#E8DEF8',

    // Background colors
    background: '#1C1B1F',
    onBackground: '#E6E1E5',
    surface: '#1C1B1F',
    onSurface: '#E6E1E5',
    surfaceVariant: '#49454F',
    onSurfaceVariant: '#CAC4D0',

    // Other adjustments for dark mode
    outline: '#938F99',
    shadow: '#000000',
    scrim: '#000000',
    inverseSurface: '#E6E1E5',
    inverseOnSurface: '#313033',
    inversePrimary: '#6750A4',
  },
};

export type Theme = typeof theme;
export type Colors = typeof theme.colors;
export type Typography = typeof theme.typography;
export type Spacing = typeof theme.spacing;