import React, { useEffect } from 'react';
import { StatusBar } from 'react-native';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import NetInfo from '@react-native-community/netinfo';

// Store
import { store, persistor, useAppDispatch, useAppSelector, setNetworkStatus, setSystemTheme } from '@store';

// Hooks
import { useColorScheme } from 'react-native';

// Components
import AppNavigator from '@navigation/AppNavigator';
import LoadingScreen from '@components/LoadingScreen';
import ErrorBoundary from '@components/ErrorBoundary';

const AppContent: React.FC = () => {
  const dispatch = useAppDispatch();
  const { theme, systemTheme } = useAppSelector(state => state.ui);
  const colorScheme = useColorScheme();

  // Update system theme when device theme changes
  useEffect(() => {
    if (colorScheme) {
      dispatch(setSystemTheme(colorScheme));
    }
  }, [colorScheme, dispatch]);

  // Monitor network status
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener(state => {
      dispatch(setNetworkStatus(state.isConnected ? 'online' : 'offline'));
    });

    return unsubscribe;
  }, [dispatch]);

  // Determine effective theme
  const effectiveTheme = theme === 'system' ? systemTheme : theme;
  const isDarkMode = effectiveTheme === 'dark';

  return (
    <>
      <StatusBar
        barStyle={isDarkMode ? 'light-content' : 'dark-content'}
        backgroundColor={isDarkMode ? '#141218' : '#FEF7FF'}
      />
      <AppNavigator />
    </>
  );
};

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <PersistGate
          loading={<LoadingScreen message="正在初始化应用..." />}
          persistor={persistor}
        >
          <AppContent />
        </PersistGate>
      </Provider>
    </ErrorBoundary>
  );
};

export default App;
