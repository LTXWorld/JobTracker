import React, { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Modal,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { useAppSelector, useAppDispatch, loadApplications } from '@store';
import { DataImportExportService, ImportResult } from '@services/dataImportExportService';
import { mockDataService } from '@services/mockDataService';

const DataManagementScreen: React.FC = () => {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { applications } = useAppSelector(state => state.applications);

  const [isExporting, setIsExporting] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);

  const exportStats = DataImportExportService.getExportStats(applications);

  const handleExportJSON = useCallback(async () => {
    setIsExporting(true);
    setShowExportModal(false);

    try {
      const success = await DataImportExportService.exportToJSON(applications, user?.username);
      if (success) {
        Alert.alert('导出成功', `已成功导出 ${applications.length} 条投递记录`);
      }
    } catch (error) {
      Alert.alert('导出失败', '无法导出数据，请重试');
    } finally {
      setIsExporting(false);
    }
  }, [applications, user?.username]);

  const handleExportCSV = useCallback(async () => {
    setIsExporting(true);
    setShowExportModal(false);

    try {
      const success = await DataImportExportService.exportToCSV(applications, user?.username);
      if (success) {
        Alert.alert('导出成功', `已成功导出 ${applications.length} 条投递记录到CSV文件`);
      }
    } catch (error) {
      Alert.alert('导出失败', '无法导出CSV文件，请重试');
    } finally {
      setIsExporting(false);
    }
  }, [applications, user?.username]);

  const handleImportJSON = useCallback(async () => {
    setIsImporting(true);
    setShowImportModal(false);

    try {
      const result = await DataImportExportService.importFromJSON();
      setImportResult(result);

      if (result.success && result.data) {
        Alert.alert(
          '导入预览',
          `将导入 ${result.data.length} 条记录${result.warnings ? `\n\n警告:\n${result.warnings.join('\n')}` : ''}`,
          [
            { text: '取消', style: 'cancel' },
            {
              text: '确认导入',
              onPress: () => confirmImport(result.data!)
            }
          ]
        );
      } else {
        Alert.alert('导入失败', result.error || '未知错误');
      }
    } catch (error) {
      Alert.alert('导入失败', '文件读取失败，请重试');
    } finally {
      setIsImporting(false);
    }
  }, []);

  const handleImportCSV = useCallback(async () => {
    setIsImporting(true);
    setShowImportModal(false);

    try {
      const result = await DataImportExportService.importFromCSV();
      setImportResult(result);

      if (result.success && result.data) {
        Alert.alert(
          '导入预览',
          `将导入 ${result.data.length} 条记录${result.warnings ? `\n\n警告:\n${result.warnings.slice(0, 5).join('\n')}${result.warnings.length > 5 ? '\n...' : ''}` : ''}`,
          [
            { text: '取消', style: 'cancel' },
            {
              text: '确认导入',
              onPress: () => confirmImport(result.data!)
            }
          ]
        );
      } else {
        Alert.alert('导入失败', result.error || '未知错误');
      }
    } catch (error) {
      Alert.alert('导入失败', 'CSV文件读取失败，请重试');
    } finally {
      setIsImporting(false);
    }
  }, []);

  const confirmImport = useCallback(async (importedApplications: any[]) => {
    setIsImporting(true);

    try {
      // 批量创建应用记录
      for (const app of importedApplications) {
        await mockDataService.createApplication({
          ...app,
          userId: user?.id || 'imported_user',
        });
      }

      // 重新加载数据
      if (user?.id) {
        await dispatch(loadApplications(user.id));
      }

      Alert.alert(
        '导入成功',
        `成功导入 ${importedApplications.length} 条投递记录`,
        [{ text: '确定', onPress: () => navigation.goBack() }]
      );
    } catch (error) {
      Alert.alert('导入失败', '数据保存失败，请重试');
    } finally {
      setIsImporting(false);
    }
  }, [user?.id, dispatch, navigation]);

  const handleClearAllData = useCallback(() => {
    Alert.alert(
      '⚠️ 危险操作',
      '这将删除所有投递记录，此操作不可恢复！\n\n建议在清空数据前先导出备份。',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确认清空',
          style: 'destructive',
          onPress: () => {
            Alert.alert(
              '最终确认',
              '您确定要删除所有数据吗？',
              [
                { text: '取消', style: 'cancel' },
                {
                  text: '删除所有数据',
                  style: 'destructive',
                  onPress: async () => {
                    // TODO: 实现清空所有数据的功能
                    Alert.alert('提示', '数据清空功能即将推出');
                  }
                }
              ]
            );
          }
        }
      ]
    );
  }, []);

  const StatCard = ({ title, value, subtitle }: { title: string; value: string; subtitle?: string }) => (
    <View style={styles.statCard}>
      <Text style={styles.statValue}>{value}</Text>
      <Text style={styles.statTitle}>{title}</Text>
      {subtitle && <Text style={styles.statSubtitle}>{subtitle}</Text>}
    </View>
  );

  const ActionButton = ({
    title,
    subtitle,
    onPress,
    icon,
    loading = false,
    danger = false
  }: {
    title: string;
    subtitle: string;
    onPress: () => void;
    icon: string;
    loading?: boolean;
    danger?: boolean;
  }) => (
    <TouchableOpacity
      style={[styles.actionButton, danger && styles.dangerButton]}
      onPress={onPress}
      disabled={loading}
      activeOpacity={0.7}
    >
      <View style={styles.actionButtonContent}>
        <View style={styles.actionButtonLeft}>
          <View style={[styles.actionIcon, danger && styles.dangerIcon]}>
            <Text style={[styles.iconText, danger && styles.dangerIconText]}>{icon}</Text>
          </View>
          <View style={styles.actionTextContainer}>
            <Text style={[styles.actionTitle, danger && styles.dangerText]}>{title}</Text>
            <Text style={styles.actionSubtitle}>{subtitle}</Text>
          </View>
        </View>
        {loading ? (
          <ActivityIndicator size="small" color={danger ? '#BA1A1A' : '#6750A4'} />
        ) : (
          <Text style={[styles.actionArrow, danger && styles.dangerText]}>→</Text>
        )}
      </View>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>数据管理</Text>
        <View style={{ width: 50 }} />
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* 数据统计 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>数据概览</Text>
          <View style={styles.statsContainer}>
            <StatCard
              title="总记录数"
              value={exportStats.total.toString()}
              subtitle="条投递记录"
            />
            <StatCard
              title="活跃记录"
              value={(exportStats.total - (exportStats.byStatus['简历挂'] || 0) - (exportStats.byStatus['一面挂'] || 0) - (exportStats.byStatus['二面挂'] || 0) - (exportStats.byStatus['被拒绝'] || 0)).toString()}
              subtitle="仍在进行中"
            />
          </View>

          {exportStats.dateRange && (
            <View style={styles.dateRangeContainer}>
              <Text style={styles.dateRangeLabel}>记录时间范围：</Text>
              <Text style={styles.dateRangeText}>
                {exportStats.dateRange.earliest.toLocaleDateString('zh-CN')} - {exportStats.dateRange.latest.toLocaleDateString('zh-CN')}
              </Text>
            </View>
          )}
        </View>

        {/* 导出功能 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>数据导出</Text>
          <Text style={styles.sectionDescription}>
            将您的投递记录导出为文件，用于备份或分享
          </Text>

          <ActionButton
            title="导出数据"
            subtitle={`导出 ${applications.length} 条记录`}
            icon="📤"
            onPress={() => setShowExportModal(true)}
            loading={isExporting}
          />
        </View>

        {/* 导入功能 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>数据导入</Text>
          <Text style={styles.sectionDescription}>
            从备份文件恢复数据或合并其他来源的投递记录
          </Text>

          <ActionButton
            title="导入数据"
            subtitle="从JSON或CSV文件导入"
            icon="📥"
            onPress={() => setShowImportModal(true)}
            loading={isImporting}
          />
        </View>

        {/* 数据清理 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>数据清理</Text>
          <Text style={styles.sectionDescription}>
            ⚠️ 危险操作：删除所有本地数据
          </Text>

          <ActionButton
            title="清空所有数据"
            subtitle="不可恢复，请谨慎操作"
            icon="🗑️"
            onPress={handleClearAllData}
            danger={true}
          />
        </View>

        <View style={styles.bottomPadding} />
      </ScrollView>

      {/* 导出选择Modal */}
      <Modal
        visible={showExportModal}
        animationType="slide"
        presentationStyle="pageSheet"
      >
        <SafeAreaView style={styles.modalContainer}>
          <View style={styles.modalHeader}>
            <TouchableOpacity onPress={() => setShowExportModal(false)}>
              <Text style={styles.modalCancelButton}>取消</Text>
            </TouchableOpacity>
            <Text style={styles.modalTitle}>选择导出格式</Text>
            <View style={{ width: 50 }} />
          </View>

          <View style={styles.modalContent}>
            <TouchableOpacity style={styles.formatOption} onPress={handleExportJSON}>
              <View style={styles.formatIcon}>
                <Text style={styles.formatIconText}>📄</Text>
              </View>
              <View style={styles.formatInfo}>
                <Text style={styles.formatTitle}>JSON 格式</Text>
                <Text style={styles.formatDescription}>
                  完整数据格式，包含所有字段信息，适合程序处理和完整备份
                </Text>
              </View>
            </TouchableOpacity>

            <TouchableOpacity style={styles.formatOption} onPress={handleExportCSV}>
              <View style={styles.formatIcon}>
                <Text style={styles.formatIconText}>📊</Text>
              </View>
              <View style={styles.formatInfo}>
                <Text style={styles.formatTitle}>CSV 格式</Text>
                <Text style={styles.formatDescription}>
                  表格格式，可用Excel等软件打开，适合数据分析和打印
                </Text>
              </View>
            </TouchableOpacity>
          </View>
        </SafeAreaView>
      </Modal>

      {/* 导入选择Modal */}
      <Modal
        visible={showImportModal}
        animationType="slide"
        presentationStyle="pageSheet"
      >
        <SafeAreaView style={styles.modalContainer}>
          <View style={styles.modalHeader}>
            <TouchableOpacity onPress={() => setShowImportModal(false)}>
              <Text style={styles.modalCancelButton}>取消</Text>
            </TouchableOpacity>
            <Text style={styles.modalTitle}>选择导入格式</Text>
            <View style={{ width: 50 }} />
          </View>

          <View style={styles.modalContent}>
            <TouchableOpacity style={styles.formatOption} onPress={handleImportJSON}>
              <View style={styles.formatIcon}>
                <Text style={styles.formatIconText}>📄</Text>
              </View>
              <View style={styles.formatInfo}>
                <Text style={styles.formatTitle}>JSON 文件</Text>
                <Text style={styles.formatDescription}>
                  导入JobView备份文件，保持所有数据完整性
                </Text>
              </View>
            </TouchableOpacity>

            <TouchableOpacity style={styles.formatOption} onPress={handleImportCSV}>
              <View style={styles.formatIcon}>
                <Text style={styles.formatIconText}>📊</Text>
              </View>
              <View style={styles.formatInfo}>
                <Text style={styles.formatTitle}>CSV 文件</Text>
                <Text style={styles.formatDescription}>
                  导入表格文件，支持从Excel等软件导出的数据
                </Text>
              </View>
            </TouchableOpacity>
          </View>
        </SafeAreaView>
      </Modal>

      {/* Loading Overlay */}
      {(isExporting || isImporting) && (
        <View style={styles.loadingOverlay}>
          <View style={styles.loadingContainer}>
            <ActivityIndicator size="large" color="#6750A4" />
            <Text style={styles.loadingText}>
              {isExporting ? '正在导出数据...' : '正在处理文件...'}
            </Text>
          </View>
        </View>
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  backButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '500',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 8,
  },
  sectionDescription: {
    fontSize: 14,
    color: '#49454F',
    marginBottom: 16,
    lineHeight: 20,
  },
  statsContainer: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 16,
  },
  statCard: {
    flex: 1,
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  statValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#6750A4',
    marginBottom: 4,
  },
  statTitle: {
    fontSize: 12,
    color: '#49454F',
    fontWeight: '500',
    textAlign: 'center',
  },
  statSubtitle: {
    fontSize: 10,
    color: '#8E8E93',
    marginTop: 2,
    textAlign: 'center',
  },
  dateRangeContainer: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    padding: 12,
    marginTop: 8,
  },
  dateRangeLabel: {
    fontSize: 12,
    color: '#49454F',
    fontWeight: '500',
    marginBottom: 4,
  },
  dateRangeText: {
    fontSize: 14,
    color: '#1C1B1F',
  },
  actionButton: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  dangerButton: {
    backgroundColor: '#FFDAD6',
  },
  actionButtonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: 16,
  },
  actionButtonLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  actionIcon: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: '#E8DEF8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  dangerIcon: {
    backgroundColor: '#FFDAD6',
  },
  iconText: {
    fontSize: 20,
  },
  dangerIconText: {
    fontSize: 20,
  },
  actionTextContainer: {
    flex: 1,
  },
  actionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 2,
  },
  actionSubtitle: {
    fontSize: 14,
    color: '#49454F',
  },
  actionArrow: {
    fontSize: 16,
    color: '#8E8E93',
    marginLeft: 8,
  },
  dangerText: {
    color: '#BA1A1A',
  },
  bottomPadding: {
    height: 20,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  modalCancelButton: {
    fontSize: 16,
    color: '#49454F',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  modalContent: {
    flex: 1,
    padding: 16,
  },
  formatOption: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    flexDirection: 'row',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  formatIcon: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: '#E8DEF8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  formatIconText: {
    fontSize: 24,
  },
  formatInfo: {
    flex: 1,
  },
  formatTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  formatDescription: {
    fontSize: 14,
    color: '#49454F',
    lineHeight: 20,
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingContainer: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 24,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 8,
    elevation: 8,
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#1C1B1F',
    fontWeight: '500',
  },
});

export default DataManagementScreen;