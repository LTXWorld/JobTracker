import React from 'react';
import {
  SafeAreaView,
  StatusBar,
  Text,
  View,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
} from 'react-native';

function App(): React.JSX.Element {
  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="dark-content" backgroundColor="#ffffff" />
      <View style={styles.header}>
        <Text style={styles.title}>JobView Mobile</Text>
        <Text style={styles.subtitle}>完整版本正在加载中...</Text>
      </View>

      <ScrollView style={styles.content}>
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>核心功能</Text>

          <TouchableOpacity style={styles.menuItem}>
            <Text style={styles.menuText}>📊 求职看板</Text>
            <Text style={styles.menuDesc}>管理求职进度</Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.menuItem}>
            <Text style={styles.menuText}>📝 求职记录</Text>
            <Text style={styles.menuDesc}>查看投递历史</Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.menuItem}>
            <Text style={styles.menuText}>⏰ 提醒中心</Text>
            <Text style={styles.menuDesc}>重要事项提醒</Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.menuItem}>
            <Text style={styles.menuText}>📈 数据统计</Text>
            <Text style={styles.menuDesc}>投递成功率分析</Text>
          </TouchableOpacity>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>当前状态</Text>
          <View style={styles.statusItem}>
            <Text style={styles.statusText}>✅ 应用基础框架已启动</Text>
          </View>
          <View style={styles.statusItem}>
            <Text style={styles.statusText}>🔄 正在加载完整功能模块</Text>
          </View>
          <View style={styles.statusItem}>
            <Text style={styles.statusText}>🔧 修复依赖和配置问题</Text>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    padding: 20,
    backgroundColor: '#ffffff',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#2196F3',
    marginBottom: 5,
  },
  subtitle: {
    fontSize: 16,
    color: '#666666',
  },
  content: {
    flex: 1,
    padding: 20,
  },
  section: {
    backgroundColor: '#ffffff',
    borderRadius: 12,
    padding: 20,
    marginBottom: 20,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333333',
    marginBottom: 15,
  },
  menuItem: {
    padding: 15,
    borderRadius: 8,
    backgroundColor: '#f8f9fa',
    marginBottom: 10,
    borderLeftWidth: 4,
    borderLeftColor: '#2196F3',
  },
  menuText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
    marginBottom: 5,
  },
  menuDesc: {
    fontSize: 14,
    color: '#666666',
  },
  statusItem: {
    paddingVertical: 8,
  },
  statusText: {
    fontSize: 14,
    color: '#333333',
  },
});

export default App;