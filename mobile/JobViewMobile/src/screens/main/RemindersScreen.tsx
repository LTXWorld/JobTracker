import React, { useState, useCallback, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Switch,
  Modal,
  TextInput,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import DatePicker from 'react-native-date-picker';
import { useAppSelector, useAppDispatch, loadApplications } from '@store';
import { NotificationService, NotificationSettings, ScheduledNotification } from '@services/notificationService';
import { ApplicationReminder, JobApplication } from '@types';

const RemindersScreen: React.FC = () => {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { applications } = useAppSelector(state => state.applications);

  const [reminders, setReminders] = useState<ApplicationReminder[]>([]);
  const [scheduledNotifications, setScheduledNotifications] = useState<ScheduledNotification[]>([]);
  const [notificationSettings, setNotificationSettings] = useState<NotificationSettings | null>(null);
  const [permissionStatus, setPermissionStatus] = useState<'granted' | 'denied' | 'unknown'>('unknown');
  const [isLoading, setIsLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [stats, setStats] = useState({
    totalScheduled: 0,
    upcomingToday: 0,
    upcomingWeek: 0,
    expired: 0,
  });

  // 新增提醒表单
  const [newReminder, setNewReminder] = useState({
    applicationId: '',
    title: '',
    description: '',
    reminderDate: new Date(Date.now() + 24 * 60 * 60 * 1000), // 默认明天
    type: 'interview' as 'interview' | 'follow_up' | 'custom',
  });
  const [showDatePicker, setShowDatePicker] = useState(false);

  // 加载数据
  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    try {
      // 加载通知设置
      const settings = await NotificationService.getSettings();
      setNotificationSettings(settings);

      // 检查通知权限
      const permission = await NotificationService.getPermissionStatus();
      setPermissionStatus(permission);

      // 加载已安排的通知
      const notifications = await NotificationService.getScheduledNotifications();
      setScheduledNotifications(notifications);

      // 加载提醒统计
      const notificationStats = await NotificationService.getNotificationStats();
      setStats(notificationStats);

      // 从应用记录中提取提醒
      const allReminders: ApplicationReminder[] = [];
      applications.forEach(app => {
        allReminders.push(...app.reminders);
      });
      setReminders(allReminders);

      // 清理过期通知
      await NotificationService.cleanupExpiredNotifications();
    } catch (error) {
      console.error('加载提醒数据失败:', error);
      Alert.alert('错误', '加载数据失败，请重试');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRequestPermission = async () => {
    const granted = await NotificationService.requestPermissions();
    if (granted) {
      setPermissionStatus('granted');
      Alert.alert('成功', '通知权限已开启');
    } else {
      Alert.alert('权限被拒绝', '无法开启通知功能，请在系统设置中手动开启');
    }
  };

  const handleToggleNotifications = async (enabled: boolean) => {
    if (!notificationSettings) return;

    const newSettings = { ...notificationSettings, enabled };
    try {
      await NotificationService.saveSettings(newSettings);
      setNotificationSettings(newSettings);

      if (!enabled) {
        await NotificationService.clearAllNotifications();
        Alert.alert('提示', '所有通知已关闭并清除');
      }
    } catch (error) {
      Alert.alert('错误', '设置保存失败');
    }
  };

  const handleUpdateSettings = async (settings: NotificationSettings) => {
    try {
      await NotificationService.saveSettings(settings);
      setNotificationSettings(settings);
      setShowSettingsModal(false);
      Alert.alert('成功', '通知设置已更新');
    } catch (error) {
      Alert.alert('错误', '设置保存失败');
    }
  };

  const handleAddReminder = async () => {
    if (!newReminder.title.trim()) {
      Alert.alert('错误', '请输入提醒标题');
      return;
    }

    if (!newReminder.applicationId) {
      Alert.alert('错误', '请选择相关的投递记录');
      return;
    }

    try {
      // 创建提醒对象
      const reminder: ApplicationReminder = {
        id: `reminder_${Date.now()}`,
        applicationId: newReminder.applicationId,
        title: newReminder.title,
        description: newReminder.description,
        reminderDate: newReminder.reminderDate,
        isCompleted: false,
        notificationType: 'local',
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      // 安排通知
      const scheduled = await NotificationService.scheduleReminder(reminder);
      if (scheduled) {
        Alert.alert('成功', '提醒已设置');
        setShowAddModal(false);
        // 重置表单
        setNewReminder({
          applicationId: '',
          title: '',
          description: '',
          reminderDate: new Date(Date.now() + 24 * 60 * 60 * 1000),
          type: 'interview',
        });
        // 重新加载数据
        await loadData();
      } else {
        Alert.alert('失败', '无法设置提醒，请检查通知权限');
      }
    } catch (error) {
      Alert.alert('错误', '设置提醒失败');
    }
  };

  const handleCancelReminder = async (reminderId: string) => {
    Alert.alert(
      '确认取消',
      '确定要取消这个提醒吗？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确定',
          onPress: async () => {
            try {
              await NotificationService.cancelReminder(reminderId);
              Alert.alert('成功', '提醒已取消');
              await loadData();
            } catch (error) {
              Alert.alert('错误', '取消提醒失败');
            }
          }
        }
      ]
    );
  };

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  const getRelativeTime = (date: Date) => {
    const now = new Date();
    const diff = date.getTime() - now.getTime();
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24));

    if (days < 0) return '已过期';
    if (days === 0) return '今天';
    if (days === 1) return '明天';
    if (days <= 7) return `${days}天后`;
    return `${Math.ceil(days / 7)}周后`;
  };

  const StatCard = ({ title, value, subtitle, color = '#6750A4' }: {
    title: string;
    value: number;
    subtitle?: string;
    color?: string;
  }) => (
    <View style={styles.statCard}>
      <Text style={[styles.statValue, { color }]}>{value}</Text>
      <Text style={styles.statTitle}>{title}</Text>
      {subtitle && <Text style={styles.statSubtitle}>{subtitle}</Text>}
    </View>
  );

  const ReminderCard = ({ notification }: { notification: ScheduledNotification }) => {
    const isExpired = notification.date < new Date();
    const application = applications.find(app => app.id === notification.data?.applicationId);

    return (
      <View style={[styles.reminderCard, isExpired && styles.expiredCard]}>
        <View style={styles.reminderHeader}>
          <Text style={[styles.reminderTitle, isExpired && styles.expiredText]}>
            {notification.title}
          </Text>
          <TouchableOpacity
            onPress={() => handleCancelReminder(notification.reminderId)}
            style={styles.cancelButton}
          >
            <Text style={styles.cancelButtonText}>×</Text>
          </TouchableOpacity>
        </View>

        {notification.message && (
          <Text style={[styles.reminderMessage, isExpired && styles.expiredText]}>
            {notification.message}
          </Text>
        )}

        {application && (
          <Text style={styles.reminderApplication}>
            {application.companyName} - {application.position}
          </Text>
        )}

        <View style={styles.reminderFooter}>
          <Text style={[styles.reminderDate, isExpired && styles.expiredText]}>
            {formatDate(notification.date)}
          </Text>
          <Text style={[
            styles.reminderRelative,
            isExpired ? styles.expiredText : styles.upcomingText
          ]}>
            {getRelativeTime(notification.date)}
          </Text>
        </View>
      </View>
    );
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#6750A4" />
          <Text style={styles.loadingText}>加载中...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>提醒中心</Text>
        <TouchableOpacity onPress={() => setShowSettingsModal(true)}>
          <Text style={styles.settingsButton}>设置</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* 权限状态 */}
        {permissionStatus !== 'granted' && (
          <View style={styles.permissionCard}>
            <Text style={styles.permissionTitle}>📢 需要通知权限</Text>
            <Text style={styles.permissionMessage}>
              为了及时提醒您重要的求职事项，请开启通知权限
            </Text>
            <TouchableOpacity
              style={styles.permissionButton}
              onPress={handleRequestPermission}
            >
              <Text style={styles.permissionButtonText}>开启通知</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* 统计信息 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>提醒统计</Text>
          <View style={styles.statsContainer}>
            <StatCard title="今日" value={stats.upcomingToday} subtitle="个提醒" color="#FF9500" />
            <StatCard title="本周" value={stats.upcomingWeek} subtitle="个提醒" color="#34C759" />
            <StatCard title="总计" value={stats.totalScheduled} subtitle="已安排" />
          </View>
        </View>

        {/* 快速设置 */}
        {notificationSettings && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>快速设置</Text>
            <View style={styles.quickSettings}>
              <View style={styles.settingRow}>
                <Text style={styles.settingLabel}>启用通知</Text>
                <Switch
                  value={notificationSettings.enabled}
                  onValueChange={handleToggleNotifications}
                  trackColor={{ false: '#F3EDF7', true: '#E8DEF8' }}
                  thumbColor={notificationSettings.enabled ? '#6750A4' : '#49454F'}
                />
              </View>
            </View>
          </View>
        )}

        {/* 操作按钮 */}
        <View style={styles.section}>
          <TouchableOpacity
            style={styles.addButton}
            onPress={() => setShowAddModal(true)}
            disabled={permissionStatus !== 'granted'}
          >
            <Text style={styles.addButtonText}>+ 添加提醒</Text>
          </TouchableOpacity>
        </View>

        {/* 提醒列表 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>我的提醒</Text>
          {scheduledNotifications.length === 0 ? (
            <View style={styles.emptyState}>
              <Text style={styles.emptyStateTitle}>暂无提醒</Text>
              <Text style={styles.emptyStateSubtitle}>
                点击上方按钮添加第一个提醒
              </Text>
            </View>
          ) : (
            scheduledNotifications
              .sort((a, b) => a.date.getTime() - b.date.getTime())
              .map((notification) => (
                <ReminderCard key={notification.id} notification={notification} />
              ))
          )}
        </View>

        <View style={styles.bottomPadding} />
      </ScrollView>

      {/* 添加提醒Modal */}
      <Modal
        visible={showAddModal}
        animationType="slide"
        presentationStyle="pageSheet"
      >
        <SafeAreaView style={styles.modalContainer}>
          <View style={styles.modalHeader}>
            <TouchableOpacity onPress={() => setShowAddModal(false)}>
              <Text style={styles.modalCancelButton}>取消</Text>
            </TouchableOpacity>
            <Text style={styles.modalTitle}>添加提醒</Text>
            <TouchableOpacity onPress={handleAddReminder}>
              <Text style={styles.modalSaveButton}>保存</Text>
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.modalContent}>
            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>相关投递 *</Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                {applications.map((app) => (
                  <TouchableOpacity
                    key={app.id}
                    style={[
                      styles.applicationOption,
                      newReminder.applicationId === app.id && styles.applicationOptionSelected
                    ]}
                    onPress={() => setNewReminder(prev => ({ ...prev, applicationId: app.id }))}
                  >
                    <Text style={[
                      styles.applicationOptionText,
                      newReminder.applicationId === app.id && styles.applicationOptionTextSelected
                    ]}>
                      {app.companyName}
                    </Text>
                    <Text style={styles.applicationOptionSubtext}>
                      {app.position}
                    </Text>
                  </TouchableOpacity>
                ))}
              </ScrollView>
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>提醒标题 *</Text>
              <TextInput
                style={styles.textInput}
                value={newReminder.title}
                onChangeText={(text) => setNewReminder(prev => ({ ...prev, title: text }))}
                placeholder="如：面试准备、简历跟进等"
                placeholderTextColor="#8E8E93"
              />
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>提醒描述</Text>
              <TextInput
                style={[styles.textInput, styles.textArea]}
                value={newReminder.description}
                onChangeText={(text) => setNewReminder(prev => ({ ...prev, description: text }))}
                placeholder="添加详细描述..."
                placeholderTextColor="#8E8E93"
                multiline={true}
                numberOfLines={3}
                textAlignVertical="top"
              />
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>提醒时间 *</Text>
              <TouchableOpacity
                style={styles.dateButton}
                onPress={() => setShowDatePicker(true)}
              >
                <Text style={styles.dateButtonText}>
                  {formatDate(newReminder.reminderDate)}
                </Text>
              </TouchableOpacity>
            </View>
          </ScrollView>

          <DatePicker
            modal
            open={showDatePicker}
            date={newReminder.reminderDate}
            mode="datetime"
            locale="zh"
            onConfirm={(date) => {
              setNewReminder(prev => ({ ...prev, reminderDate: date }));
              setShowDatePicker(false);
            }}
            onCancel={() => setShowDatePicker(false)}
          />
        </SafeAreaView>
      </Modal>

      {/* 设置Modal */}
      {notificationSettings && (
        <Modal
          visible={showSettingsModal}
          animationType="slide"
          presentationStyle="pageSheet"
        >
          <NotificationSettingsModal
            settings={notificationSettings}
            onSave={handleUpdateSettings}
            onCancel={() => setShowSettingsModal(false)}
          />
        </Modal>
      )}
    </SafeAreaView>
  );
};

// 通知设置Modal组件
const NotificationSettingsModal: React.FC<{
  settings: NotificationSettings;
  onSave: (settings: NotificationSettings) => void;
  onCancel: () => void;
}> = ({ settings, onSave, onCancel }) => {
  const [localSettings, setLocalSettings] = useState(settings);

  const updateSetting = (key: keyof NotificationSettings, value: any) => {
    setLocalSettings(prev => ({ ...prev, [key]: value }));
  };

  const updateQuietHours = (key: keyof NotificationSettings['quietHours'], value: any) => {
    setLocalSettings(prev => ({
      ...prev,
      quietHours: { ...prev.quietHours, [key]: value }
    }));
  };

  return (
    <SafeAreaView style={styles.modalContainer}>
      <View style={styles.modalHeader}>
        <TouchableOpacity onPress={onCancel}>
          <Text style={styles.modalCancelButton}>取消</Text>
        </TouchableOpacity>
        <Text style={styles.modalTitle}>通知设置</Text>
        <TouchableOpacity onPress={() => onSave(localSettings)}>
          <Text style={styles.modalSaveButton}>保存</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.modalContent}>
        <View style={styles.settingsSection}>
          <Text style={styles.settingsSectionTitle}>通知类型</Text>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>面试提醒</Text>
            <Switch
              value={localSettings.interviewReminders}
              onValueChange={(value) => updateSetting('interviewReminders', value)}
            />
          </View>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>跟进提醒</Text>
            <Switch
              value={localSettings.followUpReminders}
              onValueChange={(value) => updateSetting('followUpReminders', value)}
            />
          </View>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>状态更新</Text>
            <Switch
              value={localSettings.statusUpdates}
              onValueChange={(value) => updateSetting('statusUpdates', value)}
            />
          </View>
        </View>

        <View style={styles.settingsSection}>
          <Text style={styles.settingsSectionTitle}>通知方式</Text>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>声音</Text>
            <Switch
              value={localSettings.sound}
              onValueChange={(value) => updateSetting('sound', value)}
            />
          </View>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>震动</Text>
            <Switch
              value={localSettings.vibration}
              onValueChange={(value) => updateSetting('vibration', value)}
            />
          </View>
        </View>

        <View style={styles.settingsSection}>
          <Text style={styles.settingsSectionTitle}>静音时间</Text>

          <View style={styles.settingRow}>
            <Text style={styles.settingLabel}>启用静音时间</Text>
            <Switch
              value={localSettings.quietHours.enabled}
              onValueChange={(value) => updateQuietHours('enabled', value)}
            />
          </View>

          {localSettings.quietHours.enabled && (
            <>
              <View style={styles.settingRow}>
                <Text style={styles.settingLabel}>开始时间</Text>
                <Text style={styles.settingValue}>
                  {localSettings.quietHours.startHour.toString().padStart(2, '0')}:00
                </Text>
              </View>

              <View style={styles.settingRow}>
                <Text style={styles.settingLabel}>结束时间</Text>
                <Text style={styles.settingValue}>
                  {localSettings.quietHours.endHour.toString().padStart(2, '0')}:00
                </Text>
              </View>
            </>
          )}
        </View>
      </ScrollView>
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
  settingsButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '500',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#49454F',
  },
  permissionCard: {
    backgroundColor: '#FFF4E6',
    borderRadius: 12,
    padding: 16,
    marginBottom: 20,
    borderLeftWidth: 4,
    borderLeftColor: '#FF9500',
  },
  permissionTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 8,
  },
  permissionMessage: {
    fontSize: 14,
    color: '#49454F',
    lineHeight: 20,
    marginBottom: 12,
  },
  permissionButton: {
    backgroundColor: '#FF9500',
    borderRadius: 8,
    paddingVertical: 10,
    paddingHorizontal: 16,
    alignSelf: 'flex-start',
  },
  permissionButtonText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 12,
  },
  statsContainer: {
    flexDirection: 'row',
    gap: 12,
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
  quickSettings: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  settingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
  settingLabel: {
    fontSize: 16,
    color: '#1C1B1F',
    fontWeight: '500',
  },
  settingValue: {
    fontSize: 16,
    color: '#49454F',
  },
  addButton: {
    backgroundColor: '#6750A4',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  addButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  emptyState: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 32,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  emptyStateTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 8,
  },
  emptyStateSubtitle: {
    fontSize: 14,
    color: '#49454F',
    textAlign: 'center',
    lineHeight: 20,
  },
  reminderCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  expiredCard: {
    backgroundColor: '#F8F8F8',
    opacity: 0.7,
  },
  reminderHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  reminderTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1B1F',
    flex: 1,
  },
  expiredText: {
    color: '#8E8E93',
  },
  cancelButton: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#F3EDF7',
    justifyContent: 'center',
    alignItems: 'center',
  },
  cancelButtonText: {
    fontSize: 16,
    color: '#49454F',
    fontWeight: 'bold',
  },
  reminderMessage: {
    fontSize: 14,
    color: '#49454F',
    marginBottom: 8,
    lineHeight: 20,
  },
  reminderApplication: {
    fontSize: 12,
    color: '#6750A4',
    fontWeight: '500',
    marginBottom: 8,
  },
  reminderFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  reminderDate: {
    fontSize: 12,
    color: '#49454F',
  },
  reminderRelative: {
    fontSize: 12,
    fontWeight: '500',
  },
  upcomingText: {
    color: '#34C759',
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
  modalSaveButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '600',
  },
  modalContent: {
    flex: 1,
    padding: 16,
  },
  formGroup: {
    marginBottom: 20,
  },
  formLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#49454F',
    marginBottom: 8,
  },
  textInput: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
  },
  textArea: {
    height: 80,
    paddingTop: 12,
  },
  applicationOption: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    padding: 12,
    marginRight: 8,
    minWidth: 120,
  },
  applicationOptionSelected: {
    backgroundColor: '#E8DEF8',
    borderWidth: 2,
    borderColor: '#6750A4',
  },
  applicationOptionText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#1C1B1F',
    marginBottom: 2,
  },
  applicationOptionTextSelected: {
    color: '#21005D',
  },
  applicationOptionSubtext: {
    fontSize: 12,
    color: '#49454F',
  },
  dateButton: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  dateButtonText: {
    fontSize: 16,
    color: '#1C1B1F',
  },
  settingsSection: {
    marginBottom: 24,
  },
  settingsSectionTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 12,
  },
});

export default RemindersScreen;