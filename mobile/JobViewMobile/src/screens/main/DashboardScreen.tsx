import React, { useEffect } from 'react';
import { View, Text, ScrollView, StyleSheet, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAppSelector, useAppDispatch, logoutUser, loadApplications, loadStatistics } from '@store';
import { ApplicationStatus, ApplicationPriority } from '@types';

const DashboardScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { applications, statistics, isLoading } = useAppSelector(state => state.applications);

  useEffect(() => {
    if (user?.id) {
      dispatch(loadApplications(user.id));
      dispatch(loadStatistics(user.id));
    }
  }, [dispatch, user?.id]);

  const handleLogout = () => {
    dispatch(logoutUser());
  };

  // Calculate statistics from loaded data
  const stats = statistics || {
    total: 0,
    byStatus: {} as Record<ApplicationStatus, number>,
    byPriority: {} as Record<ApplicationPriority, number>,
    byWorkType: {} as Record<string, number>,
    recentActivity: 0
  };

  const inProgressCount = (stats.byStatus['一面中'] || 0) +
                          (stats.byStatus['二面中'] || 0) +
                          (stats.byStatus['三面中'] || 0) +
                          (stats.byStatus['HR面中'] || 0);

  const interviewCount = (stats.byStatus['一面中'] || 0) +
                        (stats.byStatus['二面中'] || 0) +
                        (stats.byStatus['三面中'] || 0);

  const offerCount = (stats.byStatus['已收到offer'] || 0) +
                    (stats.byStatus['等待offer'] || 0);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        {/* Header */}
        <View style={styles.header}>
          <View>
            <Text style={styles.title}>工作台</Text>
            <Text style={styles.subtitle}>
              欢迎回来, {user?.firstName || user?.username || '用户'}！
            </Text>
          </View>

          {/* Logout button for demo */}
          <TouchableOpacity
            style={styles.logoutButton}
            onPress={handleLogout}
          >
            <Text style={styles.logoutButtonText}>退出</Text>
          </TouchableOpacity>
        </View>

        <ScrollView showsVerticalScrollIndicator={false}>
          {/* Statistics Cards */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>今日概览</Text>

            <View style={styles.statsContainer}>
              <View style={[styles.statCard, styles.primaryCard]}>
                <Text style={styles.statNumber}>{stats.total}</Text>
                <Text style={styles.statLabelPrimary}>总投递数</Text>
              </View>

              <View style={[styles.statCard, styles.secondaryCard]}>
                <Text style={styles.statNumber}>{inProgressCount}</Text>
                <Text style={styles.statLabelSecondary}>进行中</Text>
              </View>

              <View style={[styles.statCard, styles.tertiaryCard]}>
                <Text style={styles.statNumber}>{interviewCount}</Text>
                <Text style={styles.statLabelTertiary}>面试中</Text>
              </View>

              <View style={[styles.statCard, styles.neutralCard]}>
                <Text style={styles.statNumber}>{offerCount}</Text>
                <Text style={styles.statLabelNeutral}>Offer相关</Text>
              </View>
            </View>
          </View>

          {/* Recent Activity */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>最近活动</Text>

            <View style={styles.activityCard}>
              {stats.recentActivity > 0 ? (
                <View>
                  <Text style={styles.activityText}>
                    本周有 {stats.recentActivity} 条记录更新
                  </Text>
                  {applications.slice(0, 3).map((app) => (
                    <View key={app.id} style={styles.activityItem}>
                      <Text style={styles.activityCompany}>{app.companyName}</Text>
                      <Text style={styles.activityStatus}>{app.status}</Text>
                    </View>
                  ))}
                </View>
              ) : (
                <View style={styles.emptyState}>
                  <Text style={styles.emptyStateTitle}>暂无最近活动</Text>
                  <Text style={styles.emptyStateSubtitle}>
                    开始创建您的第一条投递记录吧！
                  </Text>
                </View>
              )}
            </View>
          </View>

          {/* Quick Actions */}
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>快速操作</Text>

            <View style={styles.quickActions}>
              <TouchableOpacity style={styles.actionButton}>
                <Text style={styles.actionButtonText}>新建投递</Text>
              </TouchableOpacity>

              <TouchableOpacity style={styles.actionButtonSecondary}>
                <Text style={styles.actionButtonSecondaryText}>查看统计</Text>
              </TouchableOpacity>
            </View>
          </View>

          {/* Loading Indicator */}
          {isLoading && (
            <View style={styles.loadingContainer}>
              <Text style={styles.loadingText}>加载中...</Text>
            </View>
          )}
        </ScrollView>
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 16,
    color: '#49454F',
  },
  logoutButton: {
    backgroundColor: '#E8DEF8',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  logoutButtonText: {
    color: '#21005D',
    fontSize: 14,
    fontWeight: '500',
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 12,
  },
  statsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  statCard: {
    flex: 1,
    minWidth: '45%',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  primaryCard: {
    backgroundColor: '#EADDFF',
  },
  secondaryCard: {
    backgroundColor: '#E8DEF8',
  },
  tertiaryCard: {
    backgroundColor: '#FFD8E4',
  },
  neutralCard: {
    backgroundColor: '#F3EDF7',
  },
  statNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  statLabelPrimary: {
    fontSize: 12,
    color: '#21005D',
    opacity: 0.8,
  },
  statLabelSecondary: {
    fontSize: 12,
    color: '#1D192B',
    opacity: 0.8,
  },
  statLabelTertiary: {
    fontSize: 12,
    color: '#31111D',
    opacity: 0.8,
  },
  statLabelNeutral: {
    fontSize: 12,
    color: '#1C1B1F',
    opacity: 0.7,
  },
  activityCard: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    padding: 24,
  },
  emptyState: {
    alignItems: 'center',
  },
  emptyStateTitle: {
    fontSize: 16,
    color: '#1C1B1F',
    opacity: 0.7,
    textAlign: 'center',
    marginBottom: 8,
  },
  emptyStateSubtitle: {
    fontSize: 14,
    color: '#1C1B1F',
    opacity: 0.5,
    textAlign: 'center',
  },
  quickActions: {
    flexDirection: 'row',
    gap: 12,
  },
  actionButton: {
    flex: 1,
    backgroundColor: '#6750A4',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  actionButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  actionButtonSecondary: {
    flex: 1,
    backgroundColor: 'transparent',
    borderColor: '#6750A4',
    borderWidth: 1,
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  actionButtonSecondaryText: {
    color: '#6750A4',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingContainer: {
    padding: 20,
    alignItems: 'center',
  },
  loadingText: {
    fontSize: 16,
    color: '#49454F',
  },
  activityText: {
    fontSize: 14,
    color: '#1C1B1F',
    marginBottom: 12,
  },
  activityItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#E3E3E3',
  },
  activityCompany: {
    fontSize: 14,
    fontWeight: '500',
    color: '#1C1B1F',
  },
  activityStatus: {
    fontSize: 12,
    color: '#6750A4',
    backgroundColor: '#EADDFF',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
});

export default DashboardScreen;