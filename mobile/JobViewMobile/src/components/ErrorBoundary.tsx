import React, { Component, ErrorInfo, ReactNode } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by ErrorBoundary:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (this.state.hasError) {
      return (
        <SafeAreaView style={styles.container}>
          <View style={styles.content}>
            <Text style={styles.title}>应用出现错误</Text>
            <Text style={styles.subtitle}>
              抱歉，应用遇到了一个意外错误
            </Text>
            {__DEV__ && this.state.error && (
              <Text style={styles.errorMessage}>
                {this.state.error.message}
              </Text>
            )}
            <TouchableOpacity
              style={styles.retryButton}
              onPress={this.handleReset}
            >
              <Text style={styles.retryButtonText}>重试</Text>
            </TouchableOpacity>
          </View>
        </SafeAreaView>
      );
    }

    return this.props.children;
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 16,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    color: '#49454F',
    marginBottom: 32,
    textAlign: 'center',
    lineHeight: 24,
  },
  errorMessage: {
    fontSize: 14,
    color: '#BA1A1A',
    marginBottom: 32,
    textAlign: 'center',
    backgroundColor: '#FFDAD6',
    padding: 16,
    borderRadius: 8,
  },
  retryButton: {
    backgroundColor: '#6750A4',
    paddingHorizontal: 32,
    paddingVertical: 16,
    borderRadius: 8,
  },
  retryButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
});

export default ErrorBoundary;
