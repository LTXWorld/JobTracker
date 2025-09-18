import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Animated,
  Alert,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';

interface ErrorDisplayProps {
  error: string | Error;
  onRetry?: () => void;
  onDismiss?: () => void;
  showDetails?: boolean;
  type?: 'inline' | 'modal' | 'toast';
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({
  error,
  onRetry,
  onDismiss,
  showDetails = false,
  type = 'inline'
}) => {
  const slideAnim = React.useRef(new Animated.Value(0)).current;
  const [showFullError, setShowFullError] = React.useState(false);

  React.useEffect(() => {
    Animated.spring(slideAnim, {
      toValue: 1,
      useNativeDriver: true,
      tension: 100,
      friction: 8,
    }).start();
  }, [slideAnim]);

  const errorMessage = typeof error === 'string' ? error : error.message;
  const errorDetails = typeof error === 'object' ? error.stack : undefined;

  const handleShowDetails = () => {
    if (showDetails && errorDetails) {
      Alert.alert(
        '错误详情',
        errorDetails,
        [{ text: '确定', style: 'default' }]
      );
    } else {
      setShowFullError(!showFullError);
    }
  };

  const renderContent = () => (
    <Animated.View style={[
      styles.container,
      styles[type],
      { transform: [{ scale: slideAnim }] }
    ]}>
      <View style={styles.header}>
        <Icon name="error-outline" size={24} color="#B3261E" />
        <Text style={styles.title}>出错了</Text>
        {onDismiss && (
          <TouchableOpacity onPress={onDismiss} style={styles.closeButton}>
            <Icon name="close" size={20} color="#49454F" />
          </TouchableOpacity>
        )}
      </View>

      <Text style={styles.message}>
        {showFullError ? errorMessage : errorMessage.slice(0, 100)}
        {!showFullError && errorMessage.length > 100 && '...'}
      </Text>

      <View style={styles.actions}>
        {errorMessage.length > 100 && (
          <TouchableOpacity
            onPress={handleShowDetails}
            style={[styles.button, styles.secondaryButton]}
          >
            <Text style={styles.secondaryButtonText}>
              {showFullError ? '收起' : '详情'}
            </Text>
          </TouchableOpacity>
        )}

        {onRetry && (
          <TouchableOpacity
            onPress={onRetry}
            style={[styles.button, styles.primaryButton]}
          >
            <Icon name="refresh" size={16} color="#FFFFFF" />
            <Text style={styles.primaryButtonText}>重试</Text>
          </TouchableOpacity>
        )}
      </View>
    </Animated.View>
  );

  return renderContent();
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    margin: 16,
    elevation: 3,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 3.84,
  },
  inline: {
    borderLeftWidth: 4,
    borderLeftColor: '#B3261E',
  },
  modal: {
    marginHorizontal: 32,
    borderRadius: 16,
  },
  toast: {
    position: 'absolute',
    top: 50,
    left: 16,
    right: 16,
    zIndex: 1000,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1B1F',
    marginLeft: 8,
    flex: 1,
  },
  closeButton: {
    padding: 4,
  },
  message: {
    fontSize: 14,
    color: '#49454F',
    lineHeight: 20,
    marginBottom: 16,
  },
  actions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
  button: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    gap: 4,
  },
  primaryButton: {
    backgroundColor: '#6750A4',
  },
  secondaryButton: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#79747E',
  },
  primaryButtonText: {
    color: '#FFFFFF',
    fontWeight: '500',
    fontSize: 14,
  },
  secondaryButtonText: {
    color: '#6750A4',
    fontWeight: '500',
    fontSize: 14,
  },
});

export default ErrorDisplay;