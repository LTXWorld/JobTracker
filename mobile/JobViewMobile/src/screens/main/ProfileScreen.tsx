import React, { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  Switch,
  TextInput,
  Modal,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAppSelector, useAppDispatch, logoutUser } from '@store';
import { ProfileStackParamList } from '@types';

type ProfileScreenNavigationProp = NativeStackNavigationProp<ProfileStackParamList, 'ProfileMain'>;

const ProfileScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<ProfileScreenNavigationProp>();
  const { user } = useAppSelector(state => state.auth);
  const { applications } = useAppSelector(state => state.applications);

  const [settingsModal, setSettingsModal] = useState(false);
  const [editProfileModal, setEditProfileModal] = useState(false);
  const [notifications, setNotifications] = useState(true);
  const [darkMode, setDarkMode] = useState(false);
  const [autoSync, setAutoSync] = useState(true);
  const [isLoading, setIsLoading] = useState(false);

  // Profile edit form state
  const [profileForm, setProfileForm] = useState({
    username: user?.username || '',
    email: user?.email || '',
    phone: '',
    bio: '',
  });

  // Calculate user statistics
  const userStats = {
    totalApplications: applications.length,
    activeApplications: applications.filter(app =>
      !['简历挂', '一面挂', '二面挂', '三面挂', 'HR面挂', '被拒绝', '已入职', '已放弃'].includes(app.status)
    ).length,
    successfulApplications: applications.filter(app =>
      ['已收到offer', '已入职'].includes(app.status)
    ).length,
    joinDate: new Date().toLocaleDateString('zh-CN'),
  };

  const handleLogout = useCallback(() => {
    Alert.alert(
      '确认退出',
      '确定要退出登录吗？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '退出',
          style: 'destructive',
          onPress: () => {
            dispatch(logoutUser());
          }
        }
      ]
    );
  }, [dispatch]);

  const handleSaveProfile = useCallback(async () => {
    setIsLoading(true);
    try {
      // TODO: Integrate with real profile update API
      await new Promise<void>(resolve => setTimeout(resolve, 1000));
      Alert.alert('成功', '个人信息已更新');
      setEditProfileModal(false);
    } catch (error) {
      Alert.alert('错误', '更新失败，请重试');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleDataExport = useCallback(() => {
    Alert.alert(
      '导出数据',
      '将导出您的所有投递记录为 JSON 格式',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '导出',
          onPress: () => {
            // TODO: Implement data export functionality
            Alert.alert('提示', '数据导出功能即将推出');
          }
        }
      ]
    );
  }, []);

  const handleDataClear = useCallback(() => {
    Alert.alert(
      '清空数据',
      '⚠️ 这将删除所有投递记录，此操作不可恢复！',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确认清空',
          style: 'destructive',
          onPress: () => {
            // TODO: Implement data clear functionality
            Alert.alert('提示', '数据清空功能即将推出');
          }
        }
      ]
    );
  }, []);

  const MenuSection = ({ title, children }: { title: string; children: React.ReactNode }) => (
    <View style={styles.menuSection}>
      <Text style={styles.menuSectionTitle}>{title}</Text>
      {children}
    </View>
  );

  const MenuItem = ({
    title,
    subtitle,
    onPress,
    rightElement,
    danger = false
  }: {
    title: string;
    subtitle?: string;
    onPress: () => void;
    rightElement?: React.ReactNode;
    danger?: boolean;
  }) => (
    <TouchableOpacity
      style={styles.menuItem}
      onPress={onPress}
      activeOpacity={0.7}
    >
      <View style={styles.menuItemContent}>
        <Text style={[styles.menuItemTitle, danger && styles.dangerText]}>
          {title}
        </Text>
        {subtitle && (
          <Text style={styles.menuItemSubtitle}>{subtitle}</Text>
        )}
      </View>
      {rightElement && (
        <View style={styles.menuItemRight}>
          {rightElement}
        </View>
      )}
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.title}>个人中心</Text>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* User Profile Card */}
        <View style={styles.profileCard}>
          <View style={styles.avatarContainer}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>
                {user?.username?.charAt(0).toUpperCase() || 'U'}
              </Text>
            </View>
          </View>

          <View style={styles.userInfo}>
            <Text style={styles.username}>{user?.username || '未知用户'}</Text>
            <Text style={styles.userEmail}>{user?.email || '未设置邮箱'}</Text>
            <Text style={styles.joinDate}>加入时间：{userStats.joinDate}</Text>
          </View>

          <TouchableOpacity
            style={styles.editButton}
            onPress={() => setEditProfileModal(true)}
          >
            <Text style={styles.editButtonText}>编辑</Text>
          </TouchableOpacity>
        </View>

        {/* Statistics Cards */}
        <View style={styles.statsContainer}>
          <View style={styles.statCard}>
            <Text style={styles.statNumber}>{userStats.totalApplications}</Text>
            <Text style={styles.statLabel}>总投递</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statNumber}>{userStats.activeApplications}</Text>
            <Text style={styles.statLabel}>进行中</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statNumber}>{userStats.successfulApplications}</Text>
            <Text style={styles.statLabel}>已成功</Text>
          </View>
        </View>

        {/* Menu Sections */}
        <MenuSection title="应用设置">
          <MenuItem
            title="提醒中心"
            subtitle="管理面试和跟进提醒"
            onPress={() => navigation.navigate('Reminders')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
          <MenuItem
            title="通知推送"
            subtitle="接收面试提醒和状态更新"
            onPress={() => {}}
            rightElement={
              <Switch
                value={notifications}
                onValueChange={setNotifications}
                trackColor={{ false: '#F3EDF7', true: '#E8DEF8' }}
                thumbColor={notifications ? '#6750A4' : '#49454F'}
              />
            }
          />
          <MenuItem
            title="自动同步"
            subtitle="在网络可用时自动同步数据"
            onPress={() => {}}
            rightElement={
              <Switch
                value={autoSync}
                onValueChange={setAutoSync}
                trackColor={{ false: '#F3EDF7', true: '#E8DEF8' }}
                thumbColor={autoSync ? '#6750A4' : '#49454F'}
              />
            }
          />
          <MenuItem
            title="深色模式"
            subtitle="即将推出"
            onPress={() => {}}
            rightElement={
              <Switch
                value={darkMode}
                onValueChange={setDarkMode}
                disabled={true}
                trackColor={{ false: '#F3EDF7', true: '#E8DEF8' }}
                thumbColor={darkMode ? '#6750A4' : '#8E8E93'}
              />
            }
          />
        </MenuSection>

        <MenuSection title="数据管理">
          <MenuItem
            title="数据管理"
            subtitle="导入、导出和备份数据"
            onPress={() => navigation.navigate('DataManagement')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
          <MenuItem
            title="数据统计"
            subtitle="查看详细使用统计"
            onPress={() => Alert.alert('提示', '功能即将推出')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
          <MenuItem
            title="清空数据"
            subtitle="删除所有本地数据"
            onPress={handleDataClear}
            danger={true}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
        </MenuSection>

        <MenuSection title="帮助与支持">
          <MenuItem
            title="使用帮助"
            subtitle="查看使用指南"
            onPress={() => Alert.alert('提示', '帮助文档即将推出')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
          <MenuItem
            title="意见反馈"
            subtitle="向我们反馈问题或建议"
            onPress={() => Alert.alert('提示', '反馈功能即将推出')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
          <MenuItem
            title="关于应用"
            subtitle="版本信息和开源协议"
            onPress={() => Alert.alert('关于', 'JobView Mobile v1.0.0\n一个优雅的求职记录管理应用')}
            rightElement={<Text style={styles.menuArrow}>→</Text>}
          />
        </MenuSection>

        {/* Logout Button */}
        <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
          <Text style={styles.logoutButtonText}>退出登录</Text>
        </TouchableOpacity>

        <View style={styles.bottomPadding} />
      </ScrollView>

      {/* Edit Profile Modal */}
      <Modal
        visible={editProfileModal}
        animationType="slide"
        presentationStyle="pageSheet"
      >
        <SafeAreaView style={styles.modalContainer}>
          <View style={styles.modalHeader}>
            <TouchableOpacity onPress={() => setEditProfileModal(false)}>
              <Text style={styles.modalCancelButton}>取消</Text>
            </TouchableOpacity>
            <Text style={styles.modalTitle}>编辑资料</Text>
            <TouchableOpacity onPress={handleSaveProfile} disabled={isLoading}>
              <Text style={[styles.modalSaveButton, isLoading && styles.disabledButton]}>
                {isLoading ? '保存中...' : '保存'}
              </Text>
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.modalContent}>
            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>用户名</Text>
              <TextInput
                style={styles.formInput}
                value={profileForm.username}
                onChangeText={(text) => setProfileForm(prev => ({ ...prev, username: text }))}
                placeholder="请输入用户名"
                placeholderTextColor="#8E8E93"
              />
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>邮箱</Text>
              <TextInput
                style={styles.formInput}
                value={profileForm.email}
                onChangeText={(text) => setProfileForm(prev => ({ ...prev, email: text }))}
                placeholder="请输入邮箱地址"
                placeholderTextColor="#8E8E93"
                keyboardType="email-address"
                autoCapitalize="none"
              />
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>手机号</Text>
              <TextInput
                style={styles.formInput}
                value={profileForm.phone}
                onChangeText={(text) => setProfileForm(prev => ({ ...prev, phone: text }))}
                placeholder="请输入手机号"
                placeholderTextColor="#8E8E93"
                keyboardType="phone-pad"
              />
            </View>

            <View style={styles.formGroup}>
              <Text style={styles.formLabel}>个人简介</Text>
              <TextInput
                style={[styles.formInput, styles.textArea]}
                value={profileForm.bio}
                onChangeText={(text) => setProfileForm(prev => ({ ...prev, bio: text }))}
                placeholder="介绍一下自己..."
                placeholderTextColor="#8E8E93"
                multiline={true}
                numberOfLines={4}
                textAlignVertical="top"
              />
            </View>
          </ScrollView>

          {isLoading && (
            <View style={styles.loadingOverlay}>
              <ActivityIndicator size="large" color="#6750A4" />
            </View>
          )}
        </SafeAreaView>
      </Modal>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  header: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1C1B1F',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  profileCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 20,
    marginBottom: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
    flexDirection: 'row',
    alignItems: 'center',
  },
  avatarContainer: {
    marginRight: 16,
  },
  avatar: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: '#6750A4',
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatarText: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFFFFF',
  },
  userInfo: {
    flex: 1,
  },
  username: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  userEmail: {
    fontSize: 14,
    color: '#49454F',
    marginBottom: 4,
  },
  joinDate: {
    fontSize: 12,
    color: '#8E8E93',
  },
  editButton: {
    backgroundColor: '#E8DEF8',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 16,
  },
  editButtonText: {
    color: '#21005D',
    fontSize: 14,
    fontWeight: '600',
  },
  statsContainer: {
    flexDirection: 'row',
    marginBottom: 24,
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
  statNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#6750A4',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: '#49454F',
    fontWeight: '500',
  },
  menuSection: {
    marginBottom: 24,
  },
  menuSectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#49454F',
    marginBottom: 12,
    marginLeft: 4,
  },
  menuItem: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    marginBottom: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 1,
  },
  menuItemContent: {
    flex: 1,
    paddingVertical: 16,
    paddingHorizontal: 16,
  },
  menuItemTitle: {
    fontSize: 16,
    fontWeight: '500',
    color: '#1C1B1F',
    marginBottom: 2,
  },
  menuItemSubtitle: {
    fontSize: 14,
    color: '#49454F',
  },
  menuItemRight: {
    position: 'absolute',
    right: 16,
    top: 0,
    bottom: 0,
    justifyContent: 'center',
  },
  menuArrow: {
    fontSize: 16,
    color: '#8E8E93',
  },
  dangerText: {
    color: '#BA1A1A',
  },
  logoutButton: {
    backgroundColor: '#FFDAD6',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 12,
  },
  logoutButtonText: {
    color: '#BA1A1A',
    fontSize: 16,
    fontWeight: '600',
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
  disabledButton: {
    opacity: 0.5,
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
  formInput: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
  },
  textArea: {
    height: 100,
    paddingTop: 12,
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.3)',
    justifyContent: 'center',
    alignItems: 'center',
  },
});

export default ProfileScreen;
