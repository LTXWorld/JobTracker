import React from 'react';
import {
  TouchableOpacity,
  Text,
  StyleSheet,
  Animated,
  View,
  ActivityIndicator,
  Pressable,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';

interface EnhancedButtonProps {
  title: string;
  onPress: () => void;
  variant?: 'filled' | 'outlined' | 'text';
  size?: 'small' | 'medium' | 'large';
  disabled?: boolean;
  loading?: boolean;
  icon?: string;
  iconPosition?: 'left' | 'right';
  style?: any;
  textStyle?: any;
  fullWidth?: boolean;
}

const EnhancedButton: React.FC<EnhancedButtonProps> = ({
  title,
  onPress,
  variant = 'filled',
  size = 'medium',
  disabled = false,
  loading = false,
  icon,
  iconPosition = 'left',
  style,
  textStyle,
  fullWidth = false,
}) => {
  const scaleAnim = React.useRef(new Animated.Value(1)).current;
  const opacityAnim = React.useRef(new Animated.Value(1)).current;

  const handlePressIn = () => {
    Animated.parallel([
      Animated.spring(scaleAnim, {
        toValue: 0.96,
        useNativeDriver: true,
        tension: 300,
        friction: 10,
      }),
      Animated.timing(opacityAnim, {
        toValue: 0.8,
        duration: 100,
        useNativeDriver: true,
      }),
    ]).start();
  };

  const handlePressOut = () => {
    Animated.parallel([
      Animated.spring(scaleAnim, {
        toValue: 1,
        useNativeDriver: true,
        tension: 300,
        friction: 10,
      }),
      Animated.timing(opacityAnim, {
        toValue: 1,
        duration: 100,
        useNativeDriver: true,
      }),
    ]).start();
  };

  const getButtonStyle = () => {
    const baseStyle = [
      styles.button,
      styles[size],
      styles[variant],
      fullWidth && styles.fullWidth,
      disabled && styles.disabled,
    ];

    return baseStyle;
  };

  const getTextStyle = () => {
    return [
      styles.text,
      styles[`${size}Text`],
      styles[`${variant}Text`],
      disabled && styles.disabledText,
      textStyle,
    ];
  };

  const renderIcon = () => {
    if (loading) {
      return (
        <ActivityIndicator
          size={size === 'large' ? 20 : size === 'medium' ? 18 : 16}
          color={variant === 'filled' ? '#FFFFFF' : '#6750A4'}
          style={styles.loadingIcon}
        />
      );
    }

    if (icon) {
      return (
        <Icon
          name={icon}
          size={size === 'large' ? 20 : size === 'medium' ? 18 : 16}
          color={variant === 'filled' ? '#FFFFFF' : '#6750A4'}
          style={iconPosition === 'left' ? styles.iconLeft : styles.iconRight}
        />
      );
    }

    return null;
  };

  const renderContent = () => (
    <View style={styles.content}>
      {iconPosition === 'left' && renderIcon()}
      {!loading && <Text style={getTextStyle()}>{title}</Text>}
      {iconPosition === 'right' && renderIcon()}
    </View>
  );

  return (
    <Animated.View
      style={[
        { transform: [{ scale: scaleAnim }], opacity: opacityAnim },
        style,
      ]}
    >
      <Pressable
        onPress={onPress}
        onPressIn={handlePressIn}
        onPressOut={handlePressOut}
        disabled={disabled || loading}
        style={getButtonStyle()}
        android_ripple={{
          color: variant === 'filled' ? 'rgba(255,255,255,0.3)' : 'rgba(103,80,164,0.1)',
          borderless: false,
        }}
      >
        {renderContent()}
      </Pressable>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  button: {
    borderRadius: 100,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },

  // Size variants
  small: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    minHeight: 32,
  },
  medium: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    minHeight: 40,
  },
  large: {
    paddingHorizontal: 32,
    paddingVertical: 16,
    minHeight: 48,
  },

  // Variant styles
  filled: {
    backgroundColor: '#6750A4',
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
  },
  outlined: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#79747E',
  },
  text: {
    backgroundColor: 'transparent',
  },

  // Text styles
  smallText: {
    fontSize: 12,
    fontWeight: '500',
  },
  mediumText: {
    fontSize: 14,
    fontWeight: '500',
  },
  largeText: {
    fontSize: 16,
    fontWeight: '600',
  },

  // Variant text colors
  filledText: {
    color: '#FFFFFF',
  },
  outlinedText: {
    color: '#6750A4',
  },
  textText: {
    color: '#6750A4',
  },

  // States
  disabled: {
    opacity: 0.38,
    elevation: 0,
    shadowOpacity: 0,
  },
  disabledText: {
    color: '#1C1B1F',
  },

  // Layout
  fullWidth: {
    width: '100%',
  },

  // Icons
  iconLeft: {
    marginRight: 8,
  },
  iconRight: {
    marginLeft: 8,
  },
  loadingIcon: {
    marginRight: 8,
  },
});

export default EnhancedButton;