import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Switch,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useAppDispatch, useAppSelector } from '../store';
import {
  setUseRealAPI,
  setBaseUrl,
  testApiConnection,
  switchToRealAPI,
  clearConnectionError
} from '../store/slices/apiConfigSlice';
import { useNetworkMonitor } from '../hooks/useNetworkMonitor';
import {
  EnhancedCard,
  EnhancedButton,
  EnhancedTextInput,
  LoadingSpinner,
  ErrorDisplay,
  useToastHelpers
} from '../components/Enhanced';
import { useColors, useTypography } from '../theme/ThemeProvider';

const ApiSettingsScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const colors = useColors();
  const typography = useTypography();
  const { showSuccess, showError, showInfo } = useToastHelpers();

  const {
    useRealAPI,
    baseUrl,
    isConnected,
    lastConnectionCheck,
    connectionError
  } = useAppSelector(state => state.apiConfig);

  const {
    isFullyConnected,
    connectionQuality,
    getNetworkStatusMessage,
    getNetworkIcon,
    forceApiCheck
  } = useNetworkMonitor();

  const [localBaseUrl, setLocalBaseUrl] = useState(baseUrl);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setLocalBaseUrl(baseUrl);
  }, [baseUrl]);

  const handleToggleAPI = async (value: boolean) => {
    if (value) {
      // Switching to real API
      setIsLoading(true);
      try {
        const result = await dispatch(switchToRealAPI(localBaseUrl));
        if (switchToRealAPI.fulfilled.match(result)) {
          showSuccess('已切换到真实API模式');
        } else {
          showError(`切换失败: ${result.payload}`);
        }
      } catch (error: any) {
        showError(`切换失败: ${error.message}`);
      } finally {
        setIsLoading(false);
      }
    } else {
      // Switching to mock API
      dispatch(setUseRealAPI(false));
      showInfo('已切换到模拟API模式');
    }
  };

  const handleTestConnection = async () => {
    if (!localBaseUrl.trim()) {
      showError('请输入有效的API地址');
      return;
    }

    setIsLoading(true);
    dispatch(clearConnectionError());

    try {
      const result = await dispatch(testApiConnection(localBaseUrl));
      if (testApiConnection.fulfilled.match(result)) {
        showSuccess('API连接测试成功');
        dispatch(setBaseUrl(localBaseUrl));
      } else {
        showError(`连接测试失败: ${result.payload}`);
      }
    } catch (error: any) {
      showError(`连接测试失败: ${error.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveUrl = () => {
    if (localBaseUrl !== baseUrl) {
      dispatch(setBaseUrl(localBaseUrl));
      showSuccess('API地址已保存');
    }
  };

  const resetToDefault = () => {
    Alert.alert(
      '重置确认',
      '确定要重置为默认API设置吗？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确定',
          onPress: () => {
            const defaultUrl = __DEV__ ? 'http://10.0.2.2:8080' : 'https://api.jobview.app';
            setLocalBaseUrl(defaultUrl);
            dispatch(setBaseUrl(defaultUrl));
            dispatch(setUseRealAPI(__DEV__ ? false : true));
            showInfo('已重置为默认设置');
          },
        },
      ]
    );
  };

  const getConnectionStatusColor = () => {
    if (!useRealAPI) return colors.info;
    if (isConnected && isFullyConnected) return colors.success;
    if (connectionError) return colors.error;
    return colors.warning;
  };

  const getConnectionStatusText = () => {
    if (!useRealAPI) return '模拟模式';
    if (isConnected && isFullyConnected) return '已连接';
    if (connectionError) return '连接失败';
    return '未知状态';
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: colors.background }]}>
      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Connection Status */}
        <EnhancedCard title="连接状态" style={styles.card}>
          <View style={styles.statusContainer}>
            <View style={styles.statusItem}>
              <Icon
                name={getNetworkIcon()}
                size={24}
                color={getConnectionStatusColor()}
              />
              <View style={styles.statusText}>
                <Text style={[typography.titleMedium, { color: colors.onSurface }]}>
                  网络状态
                </Text>
                <Text style={[typography.bodyMedium, { color: colors.onSurfaceVariant }]}>
                  {getNetworkStatusMessage()}
                </Text>
              </View>
            </View>

            <View style={styles.statusItem}>
              <Icon
                name={useRealAPI ? 'cloud' : 'cloud-off'}
                size={24}
                color={getConnectionStatusColor()}
              />
              <View style={styles.statusText}>
                <Text style={[typography.titleMedium, { color: colors.onSurface }]}>
                  API状态
                </Text>
                <Text style={[typography.bodyMedium, { color: colors.onSurfaceVariant }]}>
                  {getConnectionStatusText()}
                </Text>
              </View>
              <TouchableOpacity
                onPress={forceApiCheck}
                style={styles.refreshButton}
              >
                <Icon name="refresh" size={20} color={colors.primary} />
              </TouchableOpacity>
            </View>

            {lastConnectionCheck && (
              <Text style={[typography.bodySmall, { color: colors.onSurfaceVariant, marginTop: 8 }]}>
                最后检查: {lastConnectionCheck.toLocaleString('zh-CN')}
              </Text>
            )}
          </View>
        </EnhancedCard>

        {/* API Mode Toggle */}
        <EnhancedCard title="API模式" style={styles.card}>
          <View style={styles.toggleContainer}>
            <View style={styles.toggleInfo}>
              <Text style={[typography.titleMedium, { color: colors.onSurface }]}>
                使用真实API
              </Text>
              <Text style={[typography.bodyMedium, { color: colors.onSurfaceVariant }]}>
                {useRealAPI ? '连接到服务器API' : '使用本地模拟数据'}
              </Text>
            </View>
            <Switch
              value={useRealAPI}
              onValueChange={handleToggleAPI}
              disabled={isLoading}
              trackColor={{ false: colors.outline, true: colors.primary }}
              thumbColor={useRealAPI ? colors.onPrimary : colors.onSurfaceVariant}
            />
          </View>
        </EnhancedCard>

        {/* API Configuration */}
        <EnhancedCard title="API配置" style={styles.card}>
          <View style={styles.configContainer}>
            <EnhancedTextInput
              label="API基础地址"
              value={localBaseUrl}
              onChangeText={setLocalBaseUrl}
              placeholder="输入API服务器地址"
              style={styles.urlInput}
              helperText="例如: https://api.jobview.app 或 http://192.168.1.100:8080"
            />

            <View style={styles.buttonRow}>
              <EnhancedButton
                title="测试连接"
                onPress={handleTestConnection}
                disabled={isLoading || !localBaseUrl.trim()}
                loading={isLoading}
                icon="wifi-tethering"
                variant="outlined"
                style={styles.testButton}
              />

              <EnhancedButton
                title="保存地址"
                onPress={handleSaveUrl}
                disabled={localBaseUrl === baseUrl}
                icon="save"
                style={styles.saveButton}
              />
            </View>
          </View>
        </EnhancedCard>

        {/* Error Display */}
        {connectionError && (
          <ErrorDisplay
            error={connectionError}
            onRetry={handleTestConnection}
            type="inline"
          />
        )}

        {/* Debug Information */}
        {__DEV__ && (
          <EnhancedCard title="调试信息" style={styles.card}>
            <View style={styles.debugContainer}>
              <Text style={[typography.bodySmall, { color: colors.onSurfaceVariant }]}>
                环境: {__DEV__ ? '开发' : '生产'}
              </Text>
              <Text style={[typography.bodySmall, { color: colors.onSurfaceVariant }]}>
                网络类型: {connectionQuality}
              </Text>
              <Text style={[typography.bodySmall, { color: colors.onSurfaceVariant }]}>
                完全连接: {isFullyConnected ? '是' : '否'}
              </Text>
            </View>
          </EnhancedCard>
        )}

        {/* Actions */}
        <View style={styles.actionsContainer}>
          <EnhancedButton
            title="重置为默认"
            onPress={resetToDefault}
            variant="outlined"
            icon="restore"
            fullWidth
          />
        </View>
      </ScrollView>

      {/* Loading Overlay */}
      {isLoading && (
        <LoadingSpinner
          message="正在连接..."
          overlay={true}
        />
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    flex: 1,
    padding: 16,
  },
  card: {
    marginBottom: 16,
  },
  statusContainer: {
    gap: 16,
  },
  statusItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  statusText: {
    flex: 1,
  },
  refreshButton: {
    padding: 8,
  },
  toggleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  toggleInfo: {
    flex: 1,
  },
  configContainer: {
    gap: 16,
  },
  urlInput: {
    marginBottom: 8,
  },
  buttonRow: {
    flexDirection: 'row',
    gap: 12,
  },
  testButton: {
    flex: 1,
  },
  saveButton: {
    flex: 1,
  },
  debugContainer: {
    gap: 4,
  },
  actionsContainer: {
    marginTop: 16,
    marginBottom: 32,
  },
});

export default ApiSettingsScreen;