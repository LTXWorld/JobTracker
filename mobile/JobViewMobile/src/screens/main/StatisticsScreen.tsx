import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  RefreshControl,
  ActivityIndicator,
  Dimensions,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAppSelector, useAppDispatch, loadStatistics, loadApplications } from '@store';
import { ApplicationStatus, ApplicationPriority } from '@types';

const { width: screenWidth } = Dimensions.get('window');

const StatisticsScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { statistics, applications, isLoading } = useAppSelector(state => state.applications);

  const [refreshing, setRefreshing] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<'week' | 'month' | 'quarter'>('month');

  // Load statistics and applications on mount
  useEffect(() => {
    if (user?.id) {
      dispatch(loadStatistics(user.id));
      dispatch(loadApplications(user.id));
    }
  }, [dispatch, user?.id]);

  const handleRefresh = useCallback(async () => {
    if (user?.id) {
      setRefreshing(true);
      await Promise.all([
        dispatch(loadStatistics(user.id)),
        dispatch(loadApplications(user.id))
      ]);
      setRefreshing(false);
    }
  }, [dispatch, user?.id]);

  // Calculate additional statistics
  const getSuccessRate = (): number => {
    if (!statistics) return 0;
    const total = statistics.total;
    const successful = (statistics.byStatus['已收到offer'] || 0) +
                     (statistics.byStatus['已入职'] || 0);
    return total > 0 ? Math.round((successful / total) * 100) : 0;
  };

  const getInterviewRate = (): number => {
    if (!statistics) return 0;
    const total = statistics.total;
    const interviewed = (statistics.byStatus['一面中'] || 0) +
                       (statistics.byStatus['二面中'] || 0) +
                       (statistics.byStatus['三面中'] || 0) +
                       (statistics.byStatus['HR面中'] || 0) +
                       (statistics.byStatus['等待offer'] || 0) +
                       (statistics.byStatus['已收到offer'] || 0) +
                       (statistics.byStatus['已入职'] || 0);
    return total > 0 ? Math.round((interviewed / total) * 100) : 0;
  };

  const getTopCompanies = () => {
    const companyCounts: Record<string, number> = {};
    applications.forEach(app => {
      companyCounts[app.companyName] = (companyCounts[app.companyName] || 0) + 1;
    });

    return Object.entries(companyCounts)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 5)
      .map(([company, count]) => ({ company, count }));
  };

  const getRecentTrends = () => {
    const now = new Date();
    const periods = {
      week: 7,
      month: 30,
      quarter: 90,
    };

    const cutoffDays = periods[selectedPeriod];
    const cutoffDate = new Date(now.getTime() - cutoffDays * 24 * 60 * 60 * 1000);

    const recentApps = applications.filter(app =>
      new Date(app.applicationDate) >= cutoffDate
    );

    return {
      total: recentApps.length,
      thisWeek: recentApps.filter(app =>
        new Date(app.applicationDate) >= new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
      ).length,
    };
  };

  const getStatusColor = (status: ApplicationStatus) => {
    const colors = {
      '已保存': '#8E8E93',
      '已投递': '#6750A4',
      '简历筛选中': '#0066CC',
      '笔试中': '#0066CC',
      '一面中': '#FF9500',
      '二面中': '#FF9500',
      '三面中': '#FF9500',
      'HR面中': '#FF9500',
      '等待offer': '#34C759',
      '已收到offer': '#00C896',
      '已拒绝offer': '#FF3B30',
      '简历挂': '#FF3B30',
      '笔试挂': '#FF3B30',
      '一面挂': '#FF3B30',
      '二面挂': '#FF3B30',
      '三面挂': '#FF3B30',
      'HR面挂': '#FF3B30',
      '被拒绝': '#FF3B30',
      '已入职': '#00C896',
      '已放弃': '#8E8E93',
    } as const;
    return colors[status] || '#8E8E93';
  };

  const getPriorityColor = (priority: ApplicationPriority) => {
    const colors = {
      high: '#FF3B30',
      medium: '#FF9500',
      low: '#34C759',
    };
    return colors[priority];
  };

  if (isLoading && !statistics) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#6750A4" />
          <Text style={styles.loadingText}>加载统计数据...</Text>
        </View>
      </SafeAreaView>
    );
  }

  const trends = getRecentTrends();
  const topCompanies = getTopCompanies();

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>数据统计</Text>
        <Text style={styles.subtitle}>深度分析求职数据</Text>
      </View>

      <ScrollView
        style={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
        }
      >
        {/* Overview Cards */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>总览</Text>
          <View style={styles.overviewGrid}>
            <View style={[styles.overviewCard, styles.primaryCard]}>
              <Text style={styles.overviewNumber}>{statistics?.total || 0}</Text>
              <Text style={styles.overviewLabelPrimary}>总投递数</Text>
            </View>

            <View style={[styles.overviewCard, styles.successCard]}>
              <Text style={styles.overviewNumber}>{getSuccessRate()}%</Text>
              <Text style={styles.overviewLabelSuccess}>成功率</Text>
            </View>

            <View style={[styles.overviewCard, styles.interviewCard]}>
              <Text style={styles.overviewNumber}>{getInterviewRate()}%</Text>
              <Text style={styles.overviewLabelInterview}>面试率</Text>
            </View>

            <View style={[styles.overviewCard, styles.activityCard]}>
              <Text style={styles.overviewNumber}>{statistics?.recentActivity || 0}</Text>
              <Text style={styles.overviewLabelActivity}>本周活动</Text>
            </View>
          </View>
        </View>

        {/* Period Selector */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>时间趋势</Text>
          <View style={styles.periodSelector}>
            {(['week', 'month', 'quarter'] as const).map((period) => (
              <TouchableOpacity
                key={period}
                style={[
                  styles.periodButton,
                  selectedPeriod === period && styles.periodButtonActive
                ]}
                onPress={() => setSelectedPeriod(period)}
              >
                <Text style={[
                  styles.periodButtonText,
                  selectedPeriod === period && styles.periodButtonTextActive
                ]}>
                  {period === 'week' ? '本周' : period === 'month' ? '本月' : '本季度'}
                </Text>
              </TouchableOpacity>
            ))}
          </View>

          <View style={styles.trendsContainer}>
            <View style={styles.trendCard}>
              <Text style={styles.trendNumber}>{trends.total}</Text>
              <Text style={styles.trendLabel}>
                {selectedPeriod === 'week' ? '本周投递' :
                 selectedPeriod === 'month' ? '本月投递' : '本季度投递'}
              </Text>
            </View>
            <View style={styles.trendCard}>
              <Text style={styles.trendNumber}>{trends.thisWeek}</Text>
              <Text style={styles.trendLabel}>最近7天</Text>
            </View>
          </View>
        </View>

        {/* Status Distribution */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>状态分布</Text>
          <View style={styles.statusGrid}>
            {statistics && Object.entries(statistics.byStatus)
              .filter(([, count]) => count > 0)
              .sort(([, a], [, b]) => b - a)
              .slice(0, 6)
              .map(([status, count]) => (
                <View key={status} style={styles.statusCard}>
                  <View style={[
                    styles.statusIndicator,
                    { backgroundColor: getStatusColor(status as ApplicationStatus) }
                  ]} />
                  <Text style={styles.statusCount}>{count}</Text>
                  <Text style={styles.statusLabel} numberOfLines={2}>{status}</Text>
                </View>
              ))}
          </View>
        </View>

        {/* Priority Distribution */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>优先级分布</Text>
          <View style={styles.priorityContainer}>
            {statistics && Object.entries(statistics.byPriority)
              .filter(([, count]) => count > 0)
              .map(([priority, count]) => {
                const percentage = statistics.total > 0 ? Math.round((count / statistics.total) * 100) : 0;
                return (
                  <View key={priority} style={styles.priorityItem}>
                    <View style={styles.priorityHeader}>
                      <View style={styles.priorityInfo}>
                        <View style={[
                          styles.priorityDot,
                          { backgroundColor: getPriorityColor(priority as ApplicationPriority) }
                        ]} />
                        <Text style={styles.priorityLabel}>
                          {priority === 'high' ? '高优先级' :
                           priority === 'medium' ? '中优先级' : '低优先级'}
                        </Text>
                      </View>
                      <Text style={styles.priorityCount}>{count} ({percentage}%)</Text>
                    </View>
                    <View style={styles.priorityBar}>
                      <View style={[
                        styles.priorityBarFill,
                        {
                          width: `${percentage}%`,
                          backgroundColor: getPriorityColor(priority as ApplicationPriority)
                        }
                      ]} />
                    </View>
                  </View>
                );
              })}
          </View>
        </View>

        {/* Work Type Distribution */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>工作方式偏好</Text>
          <View style={styles.workTypeGrid}>
            {statistics && Object.entries(statistics.byWorkType)
              .filter(([, count]) => count > 0)
              .map(([type, count]) => (
                <View key={type} style={styles.workTypeCard}>
                  <Text style={styles.workTypeCount}>{count}</Text>
                  <Text style={styles.workTypeLabel}>
                    {type === 'remote' ? '远程工作' :
                     type === 'onsite' ? '现场工作' : '混合办公'}
                  </Text>
                </View>
              ))}
          </View>
        </View>

        {/* Top Companies */}
        {topCompanies.length > 0 && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>投递最多的公司</Text>
            <View style={styles.companiesContainer}>
              {topCompanies.map(({ company, count }, index) => (
                <View key={company} style={styles.companyItem}>
                  <View style={styles.companyRank}>
                    <Text style={styles.companyRankText}>{index + 1}</Text>
                  </View>
                  <Text style={styles.companyName}>{company}</Text>
                  <Text style={styles.companyCount}>{count} 次</Text>
                </View>
              ))}
            </View>
          </View>
        )}

        <View style={styles.bottomSpacing} />
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
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
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
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 16,
  },
  overviewGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  overviewCard: {
    flex: 1,
    minWidth: '45%',
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
  },
  primaryCard: {
    backgroundColor: '#EADDFF',
  },
  successCard: {
    backgroundColor: '#D1F2EB',
  },
  interviewCard: {
    backgroundColor: '#FFF3CD',
  },
  activityCard: {
    backgroundColor: '#E3F2FD',
  },
  overviewNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  overviewLabelPrimary: {
    fontSize: 12,
    color: '#21005D',
    textAlign: 'center',
  },
  overviewLabelSuccess: {
    fontSize: 12,
    color: '#00695C',
    textAlign: 'center',
  },
  overviewLabelInterview: {
    fontSize: 12,
    color: '#E65100',
    textAlign: 'center',
  },
  overviewLabelActivity: {
    fontSize: 12,
    color: '#1565C0',
    textAlign: 'center',
  },
  periodSelector: {
    flexDirection: 'row',
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    padding: 4,
    marginBottom: 16,
  },
  periodButton: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 12,
    borderRadius: 6,
    alignItems: 'center',
  },
  periodButtonActive: {
    backgroundColor: '#6750A4',
  },
  periodButtonText: {
    fontSize: 14,
    color: '#49454F',
    fontWeight: '500',
  },
  periodButtonTextActive: {
    color: '#FFFFFF',
  },
  trendsContainer: {
    flexDirection: 'row',
    gap: 12,
  },
  trendCard: {
    flex: 1,
    backgroundColor: '#F3EDF7',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  trendNumber: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  trendLabel: {
    fontSize: 12,
    color: '#49454F',
    textAlign: 'center',
  },
  statusGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  statusCard: {
    backgroundColor: '#FFFFFF',
    padding: 12,
    borderRadius: 8,
    alignItems: 'center',
    minWidth: '30%',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  statusIndicator: {
    width: 20,
    height: 4,
    borderRadius: 2,
    marginBottom: 8,
  },
  statusCount: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  statusLabel: {
    fontSize: 10,
    color: '#49454F',
    textAlign: 'center',
  },
  priorityContainer: {
    gap: 12,
  },
  priorityItem: {
    backgroundColor: '#FFFFFF',
    padding: 16,
    borderRadius: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  priorityHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  priorityInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  priorityDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: 8,
  },
  priorityLabel: {
    fontSize: 14,
    color: '#1C1B1F',
    fontWeight: '500',
  },
  priorityCount: {
    fontSize: 14,
    color: '#49454F',
  },
  priorityBar: {
    height: 4,
    backgroundColor: '#F3EDF7',
    borderRadius: 2,
    overflow: 'hidden',
  },
  priorityBarFill: {
    height: '100%',
    borderRadius: 2,
  },
  workTypeGrid: {
    flexDirection: 'row',
    gap: 12,
  },
  workTypeCard: {
    flex: 1,
    backgroundColor: '#FFFFFF',
    padding: 16,
    borderRadius: 8,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  workTypeCount: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  workTypeLabel: {
    fontSize: 12,
    color: '#49454F',
    textAlign: 'center',
  },
  companiesContainer: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  companyItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  companyRank: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#6750A4',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  companyRankText: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: 'bold',
  },
  companyName: {
    flex: 1,
    fontSize: 16,
    color: '#1C1B1F',
    fontWeight: '500',
  },
  companyCount: {
    fontSize: 14,
    color: '#49454F',
  },
  bottomSpacing: {
    height: 40,
  },
});

export default StatisticsScreen;
