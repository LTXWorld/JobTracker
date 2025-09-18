import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  RefreshControl,
  FlatList,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useAppSelector, useAppDispatch, loadApplications, updateApplication, deleteApplication } from '@store';
import { JobApplication, ApplicationStatus, ApplicationsStackParamList } from '@types';
import {
  LoadingSpinner,
  ErrorDisplay,
  EnhancedCard,
  EnhancedButton,
  EnhancedTextInput,
  StatusBadge,
  SwipeableRow,
} from '../components/Enhanced';

type ApplicationsScreenNavigationProp = NativeStackNavigationProp<ApplicationsStackParamList, 'ApplicationsList'>;

const ApplicationsScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<ApplicationsScreenNavigationProp>();
  const { user } = useAppSelector(state => state.auth);
  const { applications, isLoading, error } = useAppSelector(state => state.applications);

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

  const getSwipeActions = (application: JobApplication) => {
    const rightActions = [
      {
        id: 'edit',
        title: '编辑',
        icon: 'edit',
        color: '#FFFFFF',
        backgroundColor: '#6750A4',
        onPress: () => handleEdit(application),
      },
      {
        id: 'delete',
        title: '删除',
        icon: 'delete',
        color: '#FFFFFF',
        backgroundColor: '#B3261E',
        onPress: () => handleDelete(application),
      },
    ];

    const leftActions = [
      {
        id: 'view',
        title: '详情',
        icon: 'visibility',
        color: '#FFFFFF',
        backgroundColor: '#0B57D0',
        onPress: () => handleViewDetails(application),
      },
    ];

    return { leftActions, rightActions };
  };

  const renderApplicationCard = ({ item: application }: { item: JobApplication }) => {
    const { leftActions, rightActions } = getSwipeActions(application);

    return (
      <SwipeableRow
        leftActions={leftActions}
        rightActions={rightActions}
        style={styles.swipeableContainer}
      >
        <EnhancedCard
          onPress={() => handleViewDetails(application)}
          style={styles.applicationCard}
          elevation={2}
          animated={true}
        >
          <View style={styles.cardHeader}>
            <View style={styles.cardTitleContainer}>
              <Text style={styles.companyName}>{application.companyName}</Text>
              <Text style={styles.position}>{application.position}</Text>
            </View>
            <StatusBadge status={application.status} size="medium" />
          </View>

          <View style={styles.cardContent}>
            <View style={styles.infoRow}>
              <Icon name="location-on" size={16} color="#49454F" />
              <Text style={styles.infoText}>{application.location}</Text>
            </View>

            {application.salaryRange && (
              <View style={styles.infoRow}>
                <Icon name="attach-money" size={16} color="#49454F" />
                <Text style={styles.infoText}>{application.salaryRange}</Text>
              </View>
            )}

            <View style={styles.infoRow}>
              <Icon name="work" size={16} color="#49454F" />
              <Text style={styles.infoText}>{application.workType}</Text>
            </View>

            <View style={styles.infoRow}>
              <Icon name="event" size={16} color="#49454F" />
              <Text style={styles.infoText}>
                {new Date(application.applicationDate).toLocaleDateString('zh-CN')}
              </Text>
            </View>
          </View>

          {application.tags && application.tags.length > 0 && (
            <View style={styles.tagsContainer}>
              {application.tags.slice(0, 3).map((tag, index) => (
                <View key={index} style={styles.tag}>
                  <Text style={styles.tagText}>{tag}</Text>
                </View>
              ))}
              {application.tags.length > 3 && (
                <View style={styles.tag}>
                  <Text style={styles.tagText}>+{application.tags.length - 3}</Text>
                </View>
              )}
            </View>
          )}
        </EnhancedCard>
      </SwipeableRow>
    );
  };

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Icon name="work-off" size={64} color="#79747E" />
      <Text style={styles.emptyTitle}>暂无申请记录</Text>
      <Text style={styles.emptySubtitle}>
        {searchQuery || selectedStatus !== 'all'
          ? '没有找到符合条件的申请记录'
          : '点击下方按钮添加您的第一个申请记录'}
      </Text>
      <EnhancedButton
        title="添加申请"
        onPress={() => navigation.navigate('AddApplication')}
        icon="add"
        style={styles.emptyStateButton}
      />
    </View>
  );

  if (isLoading && applications.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <LoadingSpinner message="加载申请记录中..." fullScreen={true} />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>申请记录</Text>
        <TouchableOpacity
          onPress={() => navigation.navigate('AddApplication')}
          style={styles.addButton}
        >
          <Icon name="add" size={24} color="#6750A4" />
        </TouchableOpacity>
      </View>

      {/* Search and Filter */}
      <View style={styles.searchSection}>
        <EnhancedTextInput
          placeholder="搜索公司、职位或地点..."
          value={searchQuery}
          onChangeText={setSearchQuery}
          leftIcon="search"
          style={styles.searchInput}
          variant="filled"
        />

        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          style={styles.filterContainer}
          contentContainerStyle={styles.filterContent}
        >
          {statusOptions.map((status) => (
            <TouchableOpacity
              key={status}
              onPress={() => setSelectedStatus(status)}
              style={[
                styles.filterChip,
                selectedStatus === status && styles.filterChipActive
              ]}
            >
              <Text style={[
                styles.filterChipText,
                selectedStatus === status && styles.filterChipTextActive
              ]}>
                {status === 'all' ? '全部' : status}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* Error Display */}
      {error && (
        <ErrorDisplay
          error={error}
          onRetry={() => handleRefresh()}
          type="inline"
        />
      )}

      {/* Applications List */}
      <FlatList
        data={filteredApplications}
        renderItem={renderApplicationCard}
        keyExtractor={(item) => item.id}
        style={styles.list}
        contentContainerStyle={[
          styles.listContent,
          filteredApplications.length === 0 && styles.emptyListContent
        ]}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={handleRefresh}
            colors={['#6750A4']}
            tintColor="#6750A4"
          />
        }
        showsVerticalScrollIndicator={false}
        ListEmptyComponent={renderEmptyState}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
      />

      {/* Loading Overlay */}
      {isLoading && applications.length > 0 && (
        <LoadingSpinner
          message="更新中..."
          overlay={true}
          size="small"
        />
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
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E7E0EC',
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1C1B1F',
  },
  addButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#F3E5F5',
    alignItems: 'center',
    justifyContent: 'center',
  },
  searchSection: {
    backgroundColor: '#FFFFFF',
    paddingHorizontal: 16,
    paddingBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#E7E0EC',
  },
  searchInput: {
    marginBottom: 16,
  },
  filterContainer: {
    flexGrow: 0,
  },
  filterContent: {
    paddingRight: 16,
  },
  filterChip: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    marginRight: 8,
    borderRadius: 16,
    backgroundColor: '#F7F2FA',
    borderWidth: 1,
    borderColor: '#E7E0EC',
  },
  filterChipActive: {
    backgroundColor: '#6750A4',
    borderColor: '#6750A4',
  },
  filterChipText: {
    fontSize: 14,
    color: '#49454F',
    fontWeight: '500',
  },
  filterChipTextActive: {
    color: '#FFFFFF',
  },
  list: {
    flex: 1,
  },
  listContent: {
    padding: 16,
  },
  emptyListContent: {
    flexGrow: 1,
  },
  swipeableContainer: {
    marginBottom: 12,
  },
  applicationCard: {
    marginHorizontal: 0,
    marginVertical: 0,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  cardTitleContainer: {
    flex: 1,
    marginRight: 12,
  },
  companyName: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  position: {
    fontSize: 16,
    color: '#49454F',
    fontWeight: '500',
  },
  cardContent: {
    marginBottom: 12,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 6,
  },
  infoText: {
    fontSize: 14,
    color: '#49454F',
    marginLeft: 8,
    flex: 1,
  },
  tagsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 8,
  },
  tag: {
    backgroundColor: '#F3E5F5',
    borderRadius: 8,
    paddingHorizontal: 8,
    paddingVertical: 4,
    marginRight: 6,
    marginBottom: 4,
  },
  tagText: {
    fontSize: 12,
    color: '#6750A4',
    fontWeight: '500',
  },
  separator: {
    height: 8,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1C1B1F',
    marginTop: 16,
    marginBottom: 8,
  },
  emptySubtitle: {
    fontSize: 16,
    color: '#49454F',
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 24,
  },
  emptyStateButton: {
    minWidth: 120,
  },
});

export default ApplicationsScreen;