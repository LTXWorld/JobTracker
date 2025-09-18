import React, { useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  TouchableOpacity,
  Dimensions,
  StatusBar,
  Platform,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useColors, useTypography } from '../theme/ThemeProvider';

const { width: SCREEN_WIDTH } = Dimensions.get('window');
const TOAST_HEIGHT = 60;
const STATUS_BAR_HEIGHT = Platform.OS === 'ios' ? 44 : StatusBar.currentHeight || 0;

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface ToastProps {
  message: string;
  type: ToastType;
  visible: boolean;
  onHide: () => void;
  duration?: number;
  action?: {
    label: string;
    onPress: () => void;
  };
}

const Toast: React.FC<ToastProps> = ({
  message,
  type,
  visible,
  onHide,
  duration = 4000,
  action,
}) => {
  const colors = useColors();
  const typography = useTypography();
  const translateY = useRef(new Animated.Value(-TOAST_HEIGHT - STATUS_BAR_HEIGHT)).current;
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    if (visible) {
      // Show toast
      Animated.parallel([
        Animated.timing(translateY, {
          toValue: STATUS_BAR_HEIGHT + 10,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();

      // Auto hide after duration
      const timer = setTimeout(() => {
        hideToast();
      }, duration);

      return () => clearTimeout(timer);
    } else {
      hideToast();
    }
  }, [visible, duration]);

  const hideToast = () => {
    Animated.parallel([
      Animated.timing(translateY, {
        toValue: -TOAST_HEIGHT - STATUS_BAR_HEIGHT,
        duration: 250,
        useNativeDriver: true,
      }),
      Animated.timing(opacity, {
        toValue: 0,
        duration: 250,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onHide();
    });
  };

  const getToastConfig = () => {
    switch (type) {
      case 'success':
        return {
          backgroundColor: colors.successContainer,
          textColor: colors.onSuccessContainer,
          iconName: 'check-circle',
          iconColor: colors.success,
        };
      case 'error':
        return {
          backgroundColor: colors.errorContainer,
          textColor: colors.onErrorContainer,
          iconName: 'error',
          iconColor: colors.error,
        };
      case 'warning':
        return {
          backgroundColor: colors.warningContainer,
          textColor: colors.onWarningContainer,
          iconName: 'warning',
          iconColor: colors.warning,
        };
      case 'info':
        return {
          backgroundColor: colors.infoContainer,
          textColor: colors.onInfoContainer,
          iconName: 'info',
          iconColor: colors.info,
        };
      default:
        return {
          backgroundColor: colors.surface,
          textColor: colors.onSurface,
          iconName: 'info',
          iconColor: colors.primary,
        };
    }
  };

  const config = getToastConfig();

  if (!visible) return null;

  return (
    <Animated.View
      style={[
        styles.container,
        {
          transform: [{ translateY }],
          opacity,
          backgroundColor: config.backgroundColor,
        },
      ]}
    >
      <View style={styles.content}>
        <Icon
          name={config.iconName}
          size={20}
          color={config.iconColor}
          style={styles.icon}
        />

        <Text
          style={[
            styles.message,
            typography.bodyMedium,
            { color: config.textColor },
          ]}
          numberOfLines={2}
        >
          {message}
        </Text>

        <View style={styles.actions}>
          {action && (
            <TouchableOpacity
              onPress={action.onPress}
              style={styles.actionButton}
            >
              <Text
                style={[
                  styles.actionText,
                  typography.labelMedium,
                  { color: config.iconColor },
                ]}
              >
                {action.label}
              </Text>
            </TouchableOpacity>
          )}

          <TouchableOpacity
            onPress={hideToast}
            style={styles.closeButton}
          >
            <Icon
              name="close"
              size={18}
              color={config.textColor}
            />
          </TouchableOpacity>
        </View>
      </View>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 16,
    right: 16,
    minHeight: TOAST_HEIGHT,
    borderRadius: 12,
    elevation: 6,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.27,
    shadowRadius: 4.65,
    zIndex: 9999,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    minHeight: TOAST_HEIGHT,
  },
  icon: {
    marginRight: 12,
  },
  message: {
    flex: 1,
    flexWrap: 'wrap',
  },
  actions: {
    flexDirection: 'row',
    alignItems: 'center',
    marginLeft: 8,
  },
  actionButton: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    marginRight: 8,
  },
  actionText: {
    fontWeight: '600',
  },
  closeButton: {
    padding: 4,
  },
});

export default Toast;