import React, { useEffect } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

// Import navigation types
import { RootStackParamList } from '@types';

// Import screens
import SplashScreen from '@screens/SplashScreen';
import AuthNavigator from './AuthNavigator';
import MainNavigator from './MainNavigator';

// Import hooks and actions
import { useAppSelector, useAppDispatch } from '@store';
import { loadStoredAuth } from '@store/slices/authSlice';

const Stack = createNativeStackNavigator<RootStackParamList>();

const AppNavigator: React.FC = () => {
  const dispatch = useAppDispatch();
  const { isAuthenticated } = useAppSelector(state => state.auth);
  const [isInitializing, setIsInitializing] = React.useState(true);

  // Load stored authentication on app start
  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Load stored auth data
        await dispatch(loadStoredAuth());
      } catch (error) {
        console.error('Failed to load stored auth:', error);
      } finally {
        // Wait minimum time for splash screen
        setTimeout(() => {
          setIsInitializing(false);
        }, 2000);
      }
    };

    initializeApp();
  }, [dispatch]);

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {isInitializing ? (
          <Stack.Screen name="Splash" component={SplashScreen} />
        ) : isAuthenticated ? (
          <Stack.Screen name="Main" component={MainNavigator} />
        ) : (
          <Stack.Screen name="Auth" component={AuthNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default AppNavigator;