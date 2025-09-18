import React from 'react';
import {
  SafeAreaView,
  StatusBar,
  Text,
  View,
  StyleSheet,
} from 'react-native';

function App(): React.JSX.Element {
  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="dark-content" backgroundColor="#ffffff" />
      <View style={styles.content}>
        <Text style={styles.title}>🎉 NEW VERSION LOADED! 🎉</Text>
        <Text style={styles.subtitle}>JobView Mobile v2.0</Text>
        <Text style={styles.description}>
          应用已成功从简单欢迎页面升级到功能版本！
        </Text>
        <View style={styles.statusBox}>
          <Text style={styles.statusTitle}>✅ 技术状态</Text>
          <Text style={styles.statusText}>• React Native 0.76.5 正常运行</Text>
          <Text style={styles.statusText}>• Metro 开发服务器已连接</Text>
          <Text style={styles.statusText}>• 热重载功能已激活</Text>
          <Text style={styles.statusText}>• JavaScript bundle 已更新</Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#e3f2fd',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1976d2',
    marginBottom: 10,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 20,
    color: '#2e7d32',
    marginBottom: 20,
    fontWeight: '600',
  },
  description: {
    fontSize: 16,
    color: '#424242',
    textAlign: 'center',
    marginBottom: 30,
    lineHeight: 24,
  },
  statusBox: {
    backgroundColor: '#ffffff',
    padding: 20,
    borderRadius: 12,
    borderWidth: 2,
    borderColor: '#4caf50',
    minWidth: '90%',
  },
  statusTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#2e7d32',
    marginBottom: 10,
    textAlign: 'center',
  },
  statusText: {
    fontSize: 14,
    color: '#424242',
    marginBottom: 5,
  },
});

export default App;
