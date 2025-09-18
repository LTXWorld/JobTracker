import { useState, useEffect } from 'react';
import { useColorScheme } from 'react-native';
import NetInfo from '@react-native-community/netinfo';

// Theme hook
export const useTheme = () => {
  const systemTheme = useColorScheme();
  return { systemTheme };
};

// Network status hook
export const useNetworkStatus = () => {
  const [isOnline, setIsOnline] = useState(true);
  const [networkType, setNetworkType] = useState<string>('unknown');

  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener(state => {
      setIsOnline(state.isConnected ?? false);
      setNetworkType(state.type);
    });

    return unsubscribe;
  }, []);

  return { isOnline, networkType };
};

// Debounced value hook
export const useDebounce = <T>(value: T, delay: number): T => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
};

// Animation hooks
export {
  useFadeIn,
  useScale,
  useSlideInFromBottom,
  useSlideInFromRight,
  useBounce,
  useShake,
  usePulse,
  useStaggeredAnimation,
  useEntranceAnimation,
} from './useAnimations';

// Form validation hooks
export { useForm, validationRules } from './useForm';
