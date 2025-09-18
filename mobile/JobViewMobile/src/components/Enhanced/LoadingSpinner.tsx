import React from 'react';
import { View, Text, ActivityIndicator, StyleSheet, Animated } from 'react-native';

interface LoadingSpinnerProps {
  message?: string;
  size?: 'small' | 'large';
  fullScreen?: boolean;
  overlay?: boolean;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  message = '加载中...',
  size = 'large',
  fullScreen = false,
  overlay = false
}) => {
  const fadeAnim = React.useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    Animated.timing(fadeAnim, {
      toValue: 1,
      duration: 300,
      useNativeDriver: true,
    }).start();
  }, [fadeAnim]);

  const containerStyle = [
    styles.container,
    fullScreen && styles.fullScreen,
    overlay && styles.overlay,
  ];

  return (
    <Animated.View style={[containerStyle, { opacity: fadeAnim }]}>
      <View style={styles.content}>
        <ActivityIndicator
          size={size}
          color="#6750A4"
          style={styles.spinner}
        />
        {message && (
          <Text style={[
            styles.message,
            size === 'small' && styles.smallMessage
          ]}>
            {message}
          </Text>
        )}
      </View>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  fullScreen: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  overlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    zIndex: 1000,
  },
  content: {
    alignItems: 'center',
    backgroundColor: '#FFFFFF',
    padding: 24,
    borderRadius: 16,
    elevation: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
  },
  spinner: {
    marginBottom: 16,
  },
  message: {
    fontSize: 16,
    color: '#49454F',
    textAlign: 'center',
    fontWeight: '500',
  },
  smallMessage: {
    fontSize: 14,
    marginTop: 8,
  },
});

export default LoadingSpinner;