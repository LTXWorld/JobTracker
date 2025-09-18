import React, { useState, useRef } from 'react';
import {
  View,
  TextInput,
  Text,
  StyleSheet,
  Animated,
  TouchableOpacity,
  TextInputProps,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';

interface EnhancedTextInputProps extends TextInputProps {
  label?: string;
  error?: string;
  helperText?: string;
  leftIcon?: string;
  rightIcon?: string;
  onRightIconPress?: () => void;
  variant?: 'outlined' | 'filled';
  required?: boolean;
  showCharCount?: boolean;
  maxLength?: number;
}

const EnhancedTextInput: React.FC<EnhancedTextInputProps> = ({
  label,
  error,
  helperText,
  leftIcon,
  rightIcon,
  onRightIconPress,
  variant = 'outlined',
  required = false,
  showCharCount = false,
  maxLength,
  value,
  onFocus,
  onBlur,
  style,
  ...textInputProps
}) => {
  const [isFocused, setIsFocused] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const animatedIsFocused = useRef(new Animated.Value(value ? 1 : 0)).current;
  const borderColorAnim = useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    Animated.timing(animatedIsFocused, {
      toValue: isFocused || value ? 1 : 0,
      duration: 200,
      useNativeDriver: false,
    }).start();
  }, [animatedIsFocused, isFocused, value]);

  React.useEffect(() => {
    Animated.timing(borderColorAnim, {
      toValue: error ? 2 : isFocused ? 1 : 0,
      duration: 200,
      useNativeDriver: false,
    }).start();
  }, [borderColorAnim, isFocused, error]);

  const handleFocus = (e: any) => {
    setIsFocused(true);
    onFocus?.(e);
  };

  const handleBlur = (e: any) => {
    setIsFocused(false);
    onBlur?.(e);
  };

  const labelStyle = {
    position: 'absolute' as const,
    left: leftIcon ? 44 : 16,
    top: animatedIsFocused.interpolate({
      inputRange: [0, 1],
      outputRange: [20, 8],
    }),
    fontSize: animatedIsFocused.interpolate({
      inputRange: [0, 1],
      outputRange: [16, 12],
    }),
    color: animatedIsFocused.interpolate({
      inputRange: [0, 1, 2],
      outputRange: ['#49454F', '#6750A4', '#B3261E'],
    }),
    backgroundColor: variant === 'outlined' ? '#FEF7FF' : 'transparent',
    paddingHorizontal: variant === 'outlined' ? 4 : 0,
    zIndex: 1,
  };

  const borderColor = borderColorAnim.interpolate({
    inputRange: [0, 1, 2],
    outputRange: ['#79747E', '#6750A4', '#B3261E'],
  });

  const getContainerStyle = () => {
    if (variant === 'outlined') {
      return [
        styles.outlinedContainer,
        { borderColor },
        isFocused && styles.focusedContainer,
        error && styles.errorContainer,
      ];
    } else {
      return [
        styles.filledContainer,
        isFocused && styles.focusedFilledContainer,
        error && styles.errorFilledContainer,
      ];
    }
  };

  const characterCount = value ? value.length : 0;
  const showError = !!error;
  const showHelper = !!helperText && !showError;

  return (
    <View style={[styles.wrapper, style]}>
      <Animated.View style={getContainerStyle()}>
        {leftIcon && (
          <Icon
            name={leftIcon}
            size={20}
            color={error ? '#B3261E' : isFocused ? '#6750A4' : '#49454F'}
            style={styles.leftIcon}
          />
        )}

        <View style={styles.inputWrapper}>
          {label && (
            <Animated.Text style={labelStyle}>
              {label}
              {required && <Text style={styles.required}> *</Text>}
            </Animated.Text>
          )}

          <TextInput
            {...textInputProps}
            value={value}
            onFocus={handleFocus}
            onBlur={handleBlur}
            maxLength={maxLength}
            style={[
              styles.input,
              leftIcon && styles.inputWithLeftIcon,
              rightIcon && styles.inputWithRightIcon,
              variant === 'filled' && styles.filledInput,
            ]}
            placeholderTextColor="#49454F"
          />
        </View>

        {rightIcon && (
          <TouchableOpacity
            onPress={onRightIconPress}
            style={styles.rightIconContainer}
            disabled={!onRightIconPress}
          >
            <Icon
              name={rightIcon}
              size={20}
              color={error ? '#B3261E' : isFocused ? '#6750A4' : '#49454F'}
            />
          </TouchableOpacity>
        )}
      </Animated.View>

      <View style={styles.bottomContainer}>
        <View style={styles.helperContainer}>
          {showError && (
            <View style={styles.errorContainer}>
              <Icon name="error" size={16} color="#B3261E" />
              <Text style={styles.errorText}>{error}</Text>
            </View>
          )}
          {showHelper && (
            <Text style={styles.helperText}>{helperText}</Text>
          )}
        </View>

        {showCharCount && maxLength && (
          <Text style={[
            styles.charCount,
            characterCount > maxLength * 0.9 && styles.charCountWarning,
            characterCount >= maxLength && styles.charCountError,
          ]}>
            {characterCount}/{maxLength}
          </Text>
        )}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  wrapper: {
    marginVertical: 8,
  },
  outlinedContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderRadius: 4,
    minHeight: 56,
    paddingHorizontal: 16,
    backgroundColor: '#FEF7FF',
  },
  filledContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#F7F2FA',
    borderTopLeftRadius: 4,
    borderTopRightRadius: 4,
    borderBottomWidth: 1,
    borderBottomColor: '#49454F',
    minHeight: 56,
    paddingHorizontal: 16,
  },
  focusedContainer: {
    borderWidth: 2,
  },
  errorContainer: {
    borderColor: '#B3261E',
  },
  focusedFilledContainer: {
    borderBottomWidth: 2,
    borderBottomColor: '#6750A4',
  },
  errorFilledContainer: {
    borderBottomColor: '#B3261E',
  },
  leftIcon: {
    marginRight: 12,
  },
  inputWrapper: {
    flex: 1,
    position: 'relative',
  },
  input: {
    fontSize: 16,
    color: '#1C1B1F',
    paddingTop: 20,
    paddingBottom: 8,
    margin: 0,
    padding: 0,
  },
  filledInput: {
    paddingTop: 24,
    paddingBottom: 8,
  },
  inputWithLeftIcon: {
    marginLeft: 0,
  },
  inputWithRightIcon: {
    marginRight: 12,
  },
  rightIconContainer: {
    padding: 4,
  },
  bottomContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginTop: 4,
    minHeight: 16,
  },
  helperContainer: {
    flex: 1,
  },
  errorContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  errorText: {
    fontSize: 12,
    color: '#B3261E',
    marginLeft: 4,
    flexShrink: 1,
  },
  helperText: {
    fontSize: 12,
    color: '#49454F',
  },
  charCount: {
    fontSize: 12,
    color: '#49454F',
    marginLeft: 8,
  },
  charCountWarning: {
    color: '#F9AB00',
  },
  charCountError: {
    color: '#B3261E',
  },
  required: {
    color: '#B3261E',
  },
});

export default EnhancedTextInput;