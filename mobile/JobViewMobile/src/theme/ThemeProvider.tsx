import React, { createContext, useContext, ReactNode } from 'react';
import { useColorScheme } from 'react-native';
import { theme, darkTheme, Theme } from './index';
import { useAppSelector } from '../store';

interface ThemeContextType {
  theme: Theme;
  isDark: boolean;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

interface ThemeProviderProps {
  children: ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const systemColorScheme = useColorScheme();
  const { theme: userTheme } = useAppSelector(state => state.ui);

  // Determine the actual theme to use
  const isDark = userTheme === 'dark' || (userTheme === 'system' && systemColorScheme === 'dark');
  const currentTheme = isDark ? darkTheme : theme;

  return (
    <ThemeContext.Provider value={{ theme: currentTheme, isDark }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};

// Utility hooks for commonly used theme values
export const useColors = () => {
  const { theme } = useTheme();
  return theme.colors;
};

export const useTypography = () => {
  const { theme } = useTheme();
  return theme.typography;
};

export const useSpacing = () => {
  const { theme } = useTheme();
  return theme.spacing;
};

export const useBorderRadius = () => {
  const { theme } = useTheme();
  return theme.borderRadius;
};

export const useElevation = () => {
  const { theme } = useTheme();
  return theme.elevation;
};