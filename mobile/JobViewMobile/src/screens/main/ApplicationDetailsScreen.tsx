import React, { useState, useCallback, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Linking,
  Dimensions,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAppSelector, useAppDispatch, updateApplication, deleteApplication } from '@store';
import { ApplicationsStackParamList, JobApplication, ApplicationStatus } from '@types';

type ApplicationDetailsScreenRouteProp = RouteProp<ApplicationsStackParamList, 'ApplicationDetails'>;
type ApplicationDetailsScreenNavigationProp = NativeStackNavigationProp<ApplicationsStackParamList, 'ApplicationDetails'>;

const { width: screenWidth } = Dimensions.get('window');

const ApplicationDetailsScreen: React.FC = () => {
  const navigation = useNavigation<ApplicationDetailsScreenNavigationProp>();
  const route = useRoute<ApplicationDetailsScreenRouteProp>();
  const dispatch = useAppDispatch();

  const { applications, isLoading } = useAppSelector(state => state.applications);
  const { applicationId } = route.params;

  const [updating, setUpdating] = useState(false);
  const [application, setApplication] = useState<JobApplication | null>(null);

  // Find the application
  useEffect(() => {
    const app = applications.find(app => app.id === applicationId);
    setApplication(app || null);
  }, [applications, applicationId]);

  const handleStatusUpdate = useCallback(async (newStatus: ApplicationStatus) => {
    if (!application) return;

    Alert.alert(
      '确认状态更新',
      `确定要将状态更新为"${newStatus}"吗？`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确定',
          onPress: async () => {
            setUpdating(true);
            try {
              await dispatch(updateApplication({
                id: application.id,
                updates: { status: newStatus }
              }));
              Alert.alert('成功', '状态已更新');
            } catch (error) {
              Alert.alert('错误', '状态更新失败，请重试');
            } finally {
              setUpdating(false);
            }
          }
        }
      ]
    );
  }, [application, dispatch]);

  const handleEdit = useCallback(() => {
    if (!application) return;
    navigation.navigate('EditApplication', { applicationId: application.id });
  }, [application, navigation]);

  const handleDelete = useCallback(() => {
    if (!application) return;

    Alert.alert(
      '确认删除',
      `确定要删除"${application.companyName}"的申请记录吗？此操作不可恢复。`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '删除',
          style: 'destructive',
          onPress: async () => {
            try {
              await dispatch(deleteApplication(application.id));
              Alert.alert('成功', '记录已删除', [
                { text: '确定', onPress: () => navigation.goBack() }
              ]);
            } catch (error) {
              Alert.alert('错误', '删除失败，请重试');
            }
          }
        }
      ]
    );
  }, [application, dispatch, navigation]);

  const handleOpenUrl = useCallback((url: string) => {
    if (!url) return;

    Linking.canOpenURL(url).then(supported => {
      if (supported) {
        Linking.openURL(url);
      } else {
        Alert.alert('错误', '无法打开此链接');
      }
    });
  }, []);

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

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return '#FF3B30';
      case 'medium': return '#FF9500';
      case 'low': return '#34C759';
      default: return '#8E8E93';
    }
  };

  const getNextStatuses = (currentStatus: ApplicationStatus): ApplicationStatus[] => {
    switch (currentStatus) {
      case '已投递':
        return ['简历筛选中', '简历挂'];
      case '简历筛选中':
        return ['笔试中', '一面中', '简历挂'];
      case '笔试中':
        return ['一面中', '笔试挂'];
      case '一面中':
        return ['二面中', '一面挂'];
      case '二面中':
        return ['三面中', 'HR面中', '等待offer', '二面挂'];
      case '三面中':
        return ['HR面中', '等待offer', '三面挂'];
      case 'HR面中':
        return ['等待offer', 'HR面挂'];
      case '等待offer':
        return ['已收到offer', '被拒绝'];
      case '已收到offer':
        return ['已入职', '已拒绝offer'];
      default:
        return [];
    }
  };

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      weekday: 'short',
    }).format(new Date(date));
  };

  const formatDateTime = (date: Date) => {
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(date));
  };

  const InfoCard = ({ title, children }: { title: string; children: React.ReactNode }) => (
    <View style={styles.infoCard}>
      <Text style={styles.cardTitle}>{title}</Text>
      {children}
    </View>
  );

  const InfoRow = ({ label, value, onPress }: { label: string; value: string; onPress?: () => void }) => (
    <TouchableOpacity
      style={styles.infoRow}
      onPress={onPress}
      disabled={!onPress}
      activeOpacity={onPress ? 0.7 : 1}
    >
      <Text style={styles.infoLabel}>{label}</Text>
      <Text style={[styles.infoValue, onPress && styles.linkValue]} numberOfLines={2}>
        {value}
      </Text>
    </TouchableOpacity>
  );

  if (!application) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#6750A4" />
          <Text style={styles.loadingText}>加载中...</Text>
        </View>
      </SafeAreaView>
    );
  }

  const nextStatuses = getNextStatuses(application.status);

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>投递详情</Text>
        <TouchableOpacity onPress={handleEdit}>
          <Text style={styles.editButton}>编辑</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Main Info Card */}
        <View style={styles.mainCard}>
          <View style={styles.mainCardHeader}>
            <View style={styles.companyInfo}>
              <Text style={styles.companyName}>{application.companyName}</Text>
              <Text style={styles.position}>{application.position}</Text>
            </View>
            <View style={styles.statusContainer}>
              <View style={[
                styles.statusBadge,
                { backgroundColor: getStatusColor(application.status) + '20' }
              ]}>
                <Text style={[
                  styles.statusText,
                  { color: getStatusColor(application.status) }
                ]}>
                  {application.status}
                </Text>
              </View>
              <View style={[
                styles.priorityDot,
                { backgroundColor: getPriorityColor(application.priority) }
              ]} />
            </View>
          </View>

          <View style={styles.metaContainer}>
            <Text style={styles.location}>{application.location}</Text>
            {application.salaryRange && (
              <Text style={styles.salary}>{application.salaryRange}</Text>
            )}
          </View>

          <View style={styles.dateContainer}>
            <Text style={styles.dateLabel}>投递时间：</Text>
            <Text style={styles.dateValue}>{formatDate(application.applicationDate)}</Text>
          </View>

          {application.tags.length > 0 && (
            <View style={styles.tagsContainer}>
              {application.tags.map((tag, index) => (
                <View key={index} style={styles.tag}>
                  <Text style={styles.tagText}>{tag}</Text>
                </View>
              ))}
            </View>
          )}
        </View>

        {/* Job Details */}
        <InfoCard title="职位信息">
          <InfoRow label="工作类型" value={
            application.workType === 'remote' ? '远程工作' :
            application.workType === 'onsite' ? '现场工作' : '混合工作'
          } />
          <InfoRow label="工作地点" value={application.location} />
          {application.salaryRange && (
            <InfoRow label="薪资范围" value={application.salaryRange} />
          )}
          {application.department && (
            <InfoRow label="所属部门" value={application.department} />
          )}
          {application.experienceLevel && (
            <InfoRow label="经验要求" value={application.experienceLevel} />
          )}
        </InfoCard>

        {/* Contact Info */}
        {(application.contactPerson || application.contactEmail || application.contactPhone) && (
          <InfoCard title="联系信息">
            {application.contactPerson && (
              <InfoRow label="联系人" value={application.contactPerson} />
            )}
            {application.contactEmail && (
              <InfoRow
                label="邮箱"
                value={application.contactEmail}
                onPress={() => handleOpenUrl(`mailto:${application.contactEmail}`)}
              />
            )}
            {application.contactPhone && (
              <InfoRow
                label="电话"
                value={application.contactPhone}
                onPress={() => handleOpenUrl(`tel:${application.contactPhone}`)}
              />
            )}
          </InfoCard>
        )}

        {/* Links */}
        {(application.jobUrl || application.companyUrl) && (
          <InfoCard title="相关链接">
            {application.jobUrl && (
              <InfoRow
                label="职位链接"
                value={application.jobUrl}
                onPress={() => handleOpenUrl(application.jobUrl!)}
              />
            )}
            {application.companyUrl && (
              <InfoRow
                label="公司官网"
                value={application.companyUrl}
                onPress={() => handleOpenUrl(application.companyUrl!)}
              />
            )}
          </InfoCard>
        )}

        {/* Notes */}
        {application.notes && (
          <InfoCard title="备注">
            <Text style={styles.notesText}>{application.notes}</Text>
          </InfoCard>
        )}

        {/* Status Actions */}
        {nextStatuses.length > 0 && (
          <InfoCard title="状态更新">
            <Text style={styles.actionsLabel}>可更新至：</Text>
            <View style={styles.statusActions}>
              {nextStatuses.map((status) => (
                <TouchableOpacity
                  key={status}
                  style={[
                    styles.statusActionButton,
                    { borderColor: getStatusColor(status) }
                  ]}
                  onPress={() => handleStatusUpdate(status)}
                  disabled={updating}
                >
                  <Text style={[
                    styles.statusActionText,
                    { color: getStatusColor(status) }
                  ]}>
                    {status}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </InfoCard>
        )}

        {/* Timestamps */}
        <InfoCard title="记录信息">
          <InfoRow label="创建时间" value={formatDateTime(application.createdAt)} />
          <InfoRow label="更新时间" value={formatDateTime(application.updatedAt)} />
        </InfoCard>

        {/* Delete Button */}
        <TouchableOpacity style={styles.deleteButton} onPress={handleDelete}>
          <Text style={styles.deleteButtonText}>删除记录</Text>
        </TouchableOpacity>

        <View style={styles.bottomPadding} />
      </ScrollView>

      {/* Loading Overlay */}
      {(updating || isLoading) && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#6750A4" />
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
  editButton: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '500',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  mainCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 20,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  mainCardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  companyInfo: {
    flex: 1,
    marginRight: 12,
  },
  companyName: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  position: {
    fontSize: 18,
    color: '#49454F',
    lineHeight: 24,
  },
  statusContainer: {
    alignItems: 'flex-end',
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 8,
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
  },
  priorityDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
  },
  metaContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  location: {
    fontSize: 16,
    color: '#8E8E93',
    marginRight: 16,
  },
  salary: {
    fontSize: 16,
    color: '#6750A4',
    fontWeight: '500',
  },
  dateContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  dateLabel: {
    fontSize: 14,
    color: '#49454F',
    marginRight: 8,
  },
  dateValue: {
    fontSize: 14,
    color: '#1C1B1F',
    fontWeight: '500',
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  tag: {
    backgroundColor: '#EADDFF',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  tagText: {
    fontSize: 12,
    color: '#21005D',
    fontWeight: '500',
  },
  infoCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  cardTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 12,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#F3EDF7',
  },
  infoLabel: {
    fontSize: 14,
    color: '#49454F',
    fontWeight: '500',
    flex: 1,
    marginRight: 12,
  },
  infoValue: {
    fontSize: 14,
    color: '#1C1B1F',
    flex: 2,
    textAlign: 'right',
  },
  linkValue: {
    color: '#6750A4',
    textDecorationLine: 'underline',
  },
  notesText: {
    fontSize: 14,
    color: '#1C1B1F',
    lineHeight: 20,
  },
  actionsLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#49454F',
    marginBottom: 12,
  },
  statusActions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  statusActionButton: {
    borderWidth: 1,
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  statusActionText: {
    fontSize: 12,
    fontWeight: '500',
  },
  deleteButton: {
    backgroundColor: '#FFDAD6',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  deleteButtonText: {
    color: '#BA1A1A',
    fontSize: 16,
    fontWeight: '600',
  },
  bottomPadding: {
    height: 20,
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

export default ApplicationDetailsScreen;