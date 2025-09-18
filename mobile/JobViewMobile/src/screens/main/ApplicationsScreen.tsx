import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  TextInput,
  Alert,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAppSelector, useAppDispatch, loadApplications, updateApplication, deleteApplication } from '@store';
import { JobApplication, ApplicationStatus, ApplicationsStackParamList } from '@types';

type ApplicationsScreenNavigationProp = NativeStackNavigationProp<ApplicationsStackParamList, 'ApplicationsList'>;

const ApplicationsScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<ApplicationsScreenNavigationProp>();
  const { user } = useAppSelector(state => state.auth);
  const { applications, isLoading } = useAppSelector(state => state.applications);

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedStatus, setSelectedStatus] = useState<ApplicationStatus | 'all'>('all');
  const [refreshing, setRefreshing] = useState(false);

  // Load applications on mount
  useEffect(() => {
    if (user?.id) {
      dispatch(loadApplications(user.id));
    }
  }, [dispatch, user?.id]);

  // Filter applications based on search and status
  const filteredApplications = applications.filter(app => {
    const matchesSearch = searchQuery === '' ||
      app.companyName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      app.position.toLowerCase().includes(searchQuery.toLowerCase()) ||
      app.location.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesStatus = selectedStatus === 'all' || app.status === selectedStatus;

    return matchesSearch && matchesStatus;
  });

  const handleRefresh = useCallback(async () => {
    if (user?.id) {
      setRefreshing(true);
      await dispatch(loadApplications(user.id));
      setRefreshing(false);
    }
  }, [dispatch, user?.id]);

  const handleStatusUpdate = useCallback(async (application: JobApplication, newStatus: ApplicationStatus) => {
    try {
      await dispatch(updateApplication({
        id: application.id,
        updates: { status: newStatus }
      }));
    } catch (error) {
      Alert.alert('更新失败', '无法更新申请状态，请重试');
    }
  }, [dispatch]);

  const handleDelete = useCallback((application: JobApplication) => {
    Alert.alert(
      '确认删除',
      `确定要删除${application.companyName}的申请记录吗？`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '删除',
          style: 'destructive',
          onPress: async () => {
            try {
              await dispatch(deleteApplication(application.id));
            } catch (error) {
              Alert.alert('删除失败', '无法删除申请记录，请重试');
            }
          }
        }
      ]
    );
  }, [dispatch]);

  const handleEdit = useCallback((application: JobApplication) => {
    navigation.navigate('EditApplication', { applicationId: application.id });
  }, [navigation]);

  const handleViewDetails = useCallback((application: JobApplication) => {
    navigation.navigate('ApplicationDetails', { applicationId: application.id });
  }, [navigation]);

  const statusOptions: (ApplicationStatus | 'all')[] = [
    'all', '已投递', '简历筛选中', '一面中', '二面中', '等待offer', '已收到offer', '简历挂', '一面挂', '二面挂'
  ];

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

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
    }).format(new Date(date));
  };

  if (isLoading && applications.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#6750A4" />
          <Text style={styles.loadingText}>加载投递记录...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>投递记录</Text>
          <Text style={styles.subtitle}>共 {filteredApplications.length} 条记录</Text>
        </View>

        {/* Search and Filters */}
        <View style={styles.searchSection}>
          <TextInput
            style={styles.searchInput}
            placeholder="搜索公司名称或职位..."
            value={searchQuery}
            onChangeText={setSearchQuery}
            placeholderTextColor="#8E8E93"
          />

          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={styles.statusFilters}
            contentContainerStyle={styles.statusFiltersContent}
          >
            {statusOptions.map((status) => (
              <TouchableOpacity
                key={status}
                style={[
                  styles.statusFilter,
                  selectedStatus === status && styles.statusFilterActive
                ]}
                onPress={() => setSelectedStatus(status)}
              >
                <Text style={[
                  styles.statusFilterText,
                  selectedStatus === status && styles.statusFilterTextActive
                ]}>
                  {status === 'all' ? '全部' : status}
                </Text>
              </TouchableOpacity>
            ))}
          </ScrollView>
        </View>

        {/* Applications List */}
        <ScrollView
          style={styles.listContainer}
          showsVerticalScrollIndicator={false}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
          }
        >
          {filteredApplications.length === 0 ? (
            <View style={styles.emptyState}>
              <Text style={styles.emptyStateTitle}>
                {searchQuery || selectedStatus !== 'all' ? '未找到匹配的记录' : '还没有投递记录'}
              </Text>
              <Text style={styles.emptyStateSubtitle}>
                {searchQuery || selectedStatus !== 'all' ? '尝试调整搜索条件' : '点击右上角 + 按钮开始添加投递记录'}
              </Text>
            </View>
          ) : (
            filteredApplications.map((application) => (
              <ApplicationCard
                key={application.id}
                application={application}
                onStatusUpdate={handleStatusUpdate}
                onDelete={handleDelete}
                onEdit={handleEdit}
                onViewDetails={handleViewDetails}
                getStatusColor={getStatusColor}
                formatDate={formatDate}
              />
            ))
          )}
        </ScrollView>
      </View>
    </SafeAreaView>
  );
};

interface ApplicationCardProps {
  application: JobApplication;
  onStatusUpdate: (app: JobApplication, status: ApplicationStatus) => void;
  onDelete: (app: JobApplication) => void;
  onEdit: (app: JobApplication) => void;
  onViewDetails: (app: JobApplication) => void;
  getStatusColor: (status: ApplicationStatus) => string;
  formatDate: (date: Date) => string;
}

const ApplicationCard: React.FC<ApplicationCardProps> = ({
  application,
  onStatusUpdate,
  onDelete,
  onEdit,
  onViewDetails,
  getStatusColor,
  formatDate,
}) => {
  const [showActions, setShowActions] = useState(false);

  const nextStatusOptions: ApplicationStatus[] = (() => {
    switch (application.status) {
      case '已投递':
        return ['简历筛选中', '简历挂'];
      case '简历筛选中':
        return ['一面中', '简历挂'];
      case '一面中':
        return ['二面中', '一面挂'];
      case '二面中':
        return ['三面中', '等待offer', '二面挂'];
      case '三面中':
        return ['HR面中', '等待offer', '三面挂'];
      case 'HR面中':
        return ['等待offer', 'HR面挂'];
      case '等待offer':
        return ['已收到offer', '被拒绝'];
      default:
        return [];
    }
  })();

  return (
    <View style={styles.applicationCard}>
      <TouchableOpacity
        style={styles.cardHeader}
        onPress={() => onViewDetails(application)}
      >
        <View style={styles.cardMainInfo}>
          <Text style={styles.companyName}>{application.companyName}</Text>
          <Text style={styles.position}>{application.position}</Text>
          <View style={styles.cardMeta}>
            <Text style={styles.location}>{application.location}</Text>
            {application.salaryRange && (
              <Text style={styles.salary}>{application.salaryRange}</Text>
            )}
          </View>
        </View>

        <View style={styles.cardRight}>
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
          <Text style={styles.dateText}>
            {formatDate(application.applicationDate)}
          </Text>
          <TouchableOpacity
            style={styles.moreButton}
            onPress={(e) => {
              e.stopPropagation();
              setShowActions(!showActions);
            }}
          >
            <Text style={styles.moreButtonText}>⋯</Text>
          </TouchableOpacity>
        </View>
      </TouchableOpacity>

      {showActions && (
        <View style={styles.cardActions}>
          {nextStatusOptions.length > 0 && (
            <View style={styles.statusActions}>
              <Text style={styles.actionsLabel}>更新状态：</Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                {nextStatusOptions.map((status) => (
                  <TouchableOpacity
                    key={status}
                    style={[
                      styles.statusActionButton,
                      { borderColor: getStatusColor(status) }
                    ]}
                    onPress={() => {
                      onStatusUpdate(application, status);
                      setShowActions(false);
                    }}
                  >
                    <Text style={[
                      styles.statusActionText,
                      { color: getStatusColor(status) }
                    ]}>
                      {status}
                    </Text>
                  </TouchableOpacity>
                ))}
              </ScrollView>
            </View>
          )}

          <View style={styles.cardActionsRow}>
            <TouchableOpacity
              style={styles.editButton}
              onPress={() => {
                onEdit(application);
                setShowActions(false);
              }}
            >
              <Text style={styles.editButtonText}>编辑</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.deleteButton}
              onPress={() => {
                onDelete(application);
                setShowActions(false);
              }}
            >
              <Text style={styles.deleteButtonText}>删除</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
    </View>
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
    marginBottom: 20,
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
  searchSection: {
    marginBottom: 20,
  },
  searchInput: {
    backgroundColor: '#F3EDF7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1C1B1F',
    marginBottom: 12,
  },
  statusFilters: {
    marginBottom: 8,
  },
  statusFiltersContent: {
    paddingRight: 16,
  },
  statusFilter: {
    backgroundColor: '#F3EDF7',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    marginRight: 8,
  },
  statusFilterActive: {
    backgroundColor: '#6750A4',
  },
  statusFilterText: {
    fontSize: 14,
    color: '#49454F',
    fontWeight: '500',
  },
  statusFilterTextActive: {
    color: '#FFFFFF',
  },
  listContainer: {
    flex: 1,
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
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyStateTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 8,
    textAlign: 'center',
  },
  emptyStateSubtitle: {
    fontSize: 14,
    color: '#49454F',
    textAlign: 'center',
    lineHeight: 20,
  },
  applicationCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  cardHeader: {
    flexDirection: 'row',
    padding: 16,
  },
  cardMainInfo: {
    flex: 1,
    marginRight: 12,
  },
  companyName: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  position: {
    fontSize: 16,
    color: '#49454F',
    marginBottom: 8,
  },
  cardMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  location: {
    fontSize: 14,
    color: '#8E8E93',
    marginRight: 12,
  },
  salary: {
    fontSize: 14,
    color: '#6750A4',
    fontWeight: '500',
  },
  cardRight: {
    alignItems: 'flex-end',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    marginBottom: 8,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  dateText: {
    fontSize: 12,
    color: '#8E8E93',
  },
  cardActions: {
    borderTopWidth: 1,
    borderTopColor: '#F3EDF7',
    padding: 16,
  },
  statusActions: {
    marginBottom: 16,
  },
  actionsLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#49454F',
    marginBottom: 8,
  },
  statusActionButton: {
    borderWidth: 1,
    borderRadius: 16,
    paddingHorizontal: 12,
    paddingVertical: 6,
    marginRight: 8,
  },
  statusActionText: {
    fontSize: 12,
    fontWeight: '500',
  },
  cardActionsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  editButton: {
    flex: 1,
    backgroundColor: '#E8DEF8',
    paddingVertical: 12,
    borderRadius: 8,
    marginRight: 8,
    alignItems: 'center',
  },
  editButtonText: {
    color: '#21005D',
    fontWeight: '600',
  },
  deleteButton: {
    flex: 1,
    backgroundColor: '#FFDAD6',
    paddingVertical: 12,
    borderRadius: 8,
    marginLeft: 8,
    alignItems: 'center',
  },
  deleteButtonText: {
    color: '#BA1A1A',
    fontWeight: '600',
  },
  moreButton: {
    marginTop: 8,
    paddingHorizontal: 8,
    paddingVertical: 4,
  },
  moreButtonText: {
    fontSize: 16,
    color: '#8E8E93',
    fontWeight: 'bold',
  },
});

export default ApplicationsScreen;