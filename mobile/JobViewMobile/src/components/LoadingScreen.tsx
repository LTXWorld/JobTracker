import React from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';

interface LoadingScreenProps {
  message?: string;
}

const LoadingScreen: React.FC<LoadingScreenProps> = ({ message = '加载中...' }) => {
  return (
    <View style={styles.container}>
      <ActivityIndicator size="large" color="#6750A4" />
      <Text style={styles.message}>{message}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#FEF7FF',
    padding: 20,
  },
  message: {
    marginTop: 16,
    fontSize: 16,
    color: '#49454F',
    textAlign: 'center',
  },
});

export default LoadingScreen;
