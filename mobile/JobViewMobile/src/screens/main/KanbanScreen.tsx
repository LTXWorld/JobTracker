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
import { useAppSelector, useAppDispatch, loadKanbanData, updateApplication } from '@store';
import { JobApplication, ApplicationStatus } from '@types';

const { width: screenWidth } = Dimensions.get('window');
const COLUMN_WIDTH = screenWidth * 0.8;

const KanbanScreen: React.FC = () => {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { kanbanData, isLoading } = useAppSelector(state => state.applications);

  const [refreshing, setRefreshing] = useState(false);

  // Define column configuration
  const columns: { status: ApplicationStatus; title: string; color: string }[] = [
    { status: '已投递', title: '已投递', color: '#6750A4' },
    { status: '简历筛选中', title: '简历筛选', color: '#0066CC' },
    { status: '一面中', title: '一面', color: '#FF9500' },
    { status: '二面中', title: '二面', color: '#FF9500' },
    { status: '等待offer', title: '等待Offer', color: '#34C759' },
    { status: '已收到offer', title: '已收到Offer', color: '#00C896' },
  ];

  // Load kanban data on mount
  useEffect(() => {
    if (user?.id) {
      dispatch(loadKanbanData(user.id));
    }
  }, [dispatch, user?.id]);

  const handleRefresh = useCallback(async () => {
    if (user?.id) {
      setRefreshing(true);
      await dispatch(loadKanbanData(user.id));
      setRefreshing(false);
    }
  }, [dispatch, user?.id]);

  const handleStatusUpdate = useCallback(async (application: JobApplication, newStatus: ApplicationStatus) => {
    try {
      await dispatch(updateApplication({
        id: application.id,
        updates: { status: newStatus }
      }));
      // Reload kanban data to reflect changes
      if (user?.id) {
        dispatch(loadKanbanData(user.id));
      }
    } catch (error) {
      Alert.alert('更新失败', '无法更新申请状态，请重试');
    }
  }, [dispatch, user?.id]);

  const getNextStatuses = (currentStatus: ApplicationStatus): ApplicationStatus[] => {
    switch (currentStatus) {
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
  };

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
    }).format(new Date(date));
  };

  if (isLoading && Object.keys(kanbanData).every(key => kanbanData[key as ApplicationStatus].length === 0)) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#6750A4" />
          <Text style={styles.loadingText}>加载看板数据...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>看板视图</Text>
        <Text style={styles.subtitle}>拖拽式状态管理</Text>
      </View>

      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        style={styles.kanbanContainer}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
        }
      >
        {columns.map((column) => (
          <KanbanColumn
            key={column.status}
            column={column}
            applications={kanbanData[column.status] || []}
            onStatusUpdate={handleStatusUpdate}
            getNextStatuses={getNextStatuses}
            formatDate={formatDate}
          />
        ))}
      </ScrollView>
    </SafeAreaView>
  );
};

interface KanbanColumnProps {
  column: { status: ApplicationStatus; title: string; color: string };
  applications: JobApplication[];
  onStatusUpdate: (app: JobApplication, status: ApplicationStatus) => void;
  getNextStatuses: (status: ApplicationStatus) => ApplicationStatus[];
  formatDate: (date: Date) => string;
}

const KanbanColumn: React.FC<KanbanColumnProps> = ({
  column,
  applications,
  onStatusUpdate,
  getNextStatuses,
  formatDate,
}) => {
  return (
    <View style={styles.column}>
      <View style={[styles.columnHeader, { backgroundColor: column.color + '20' }]}>
        <Text style={[styles.columnTitle, { color: column.color }]}>
          {column.title}
        </Text>
        <View style={[styles.countBadge, { backgroundColor: column.color }]}>
          <Text style={styles.countText}>{applications.length}</Text>
        </View>
      </View>

      <ScrollView
        style={styles.columnContent}
        showsVerticalScrollIndicator={false}
        nestedScrollEnabled={true}
      >
        {applications.length === 0 ? (
          <View style={styles.emptyColumn}>
            <Text style={styles.emptyColumnText}>暂无记录</Text>
          </View>
        ) : (
          applications.map((application) => (
            <KanbanCard
              key={application.id}
              application={application}
              onStatusUpdate={onStatusUpdate}
              getNextStatuses={getNextStatuses}
              formatDate={formatDate}
              columnColor={column.color}
            />
          ))
        )}
      </ScrollView>
    </View>
  );
};

interface KanbanCardProps {
  application: JobApplication;
  onStatusUpdate: (app: JobApplication, status: ApplicationStatus) => void;
  getNextStatuses: (status: ApplicationStatus) => ApplicationStatus[];
  formatDate: (date: Date) => string;
  columnColor: string;
}

const KanbanCard: React.FC<KanbanCardProps> = ({
  application,
  onStatusUpdate,
  getNextStatuses,
  formatDate,
  columnColor,
}) => {
  const [showActions, setShowActions] = useState(false);

  const nextStatuses = getNextStatuses(application.status);

  const handleStatusChange = (newStatus: ApplicationStatus) => {
    Alert.alert(
      '确认状态更新',
      `确定要将${application.companyName}的状态更新为"${newStatus}"吗？`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '确定',
          onPress: () => {
            onStatusUpdate(application, newStatus);
            setShowActions(false);
          }
        }
      ]
    );
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return '#FF3B30';
      case 'medium': return '#FF9500';
      case 'low': return '#34C759';
      default: return '#8E8E93';
    }
  };

  return (
    <TouchableOpacity
      style={styles.card}
      onPress={() => setShowActions(!showActions)}
      activeOpacity={0.7}
    >
      <View style={styles.cardHeader}>
        <Text style={styles.companyName} numberOfLines={1}>
          {application.companyName}
        </Text>
        <View style={[
          styles.priorityDot,
          { backgroundColor: getPriorityColor(application.priority) }
        ]} />
      </View>

      <Text style={styles.position} numberOfLines={2}>
        {application.position}
      </Text>

      <View style={styles.cardMeta}>
        <Text style={styles.location} numberOfLines={1}>
          {application.location}
        </Text>
        <Text style={styles.date}>
          {formatDate(application.applicationDate)}
        </Text>
      </View>

      {application.salaryRange && (
        <Text style={styles.salary}>{application.salaryRange}</Text>
      )}

      {application.tags.length > 0 && (
        <View style={styles.tagsContainer}>
          {application.tags.slice(0, 2).map((tag, index) => (
            <View key={index} style={styles.tag}>
              <Text style={styles.tagText}>{tag}</Text>
            </View>
          ))}
          {application.tags.length > 2 && (
            <Text style={styles.moreTags}>+{application.tags.length - 2}</Text>
          )}
        </View>
      )}

      {showActions && nextStatuses.length > 0 && (
        <View style={styles.cardActions}>
          <Text style={styles.actionsLabel}>移动到:</Text>
          <ScrollView horizontal showsHorizontalScrollIndicator={false}>
            {nextStatuses.map((status) => (
              <TouchableOpacity
                key={status}
                style={[styles.actionButton, { borderColor: columnColor }]}
                onPress={() => handleStatusChange(status)}
              >
                <Text style={[styles.actionButtonText, { color: columnColor }]}>
                  {status}
                </Text>
              </TouchableOpacity>
            ))}
          </ScrollView>
        </View>
      )}
    </TouchableOpacity>
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
  kanbanContainer: {
    flex: 1,
    paddingVertical: 16,
  },
  column: {
    width: COLUMN_WIDTH,
    marginHorizontal: 8,
    backgroundColor: '#F3EDF7',
    borderRadius: 12,
    overflow: 'hidden',
  },
  columnHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
  },
  columnTitle: {
    fontSize: 16,
    fontWeight: 'bold',
  },
  countBadge: {
    minWidth: 24,
    height: 24,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 8,
  },
  countText: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: 'bold',
  },
  columnContent: {
    flex: 1,
    padding: 8,
  },
  emptyColumn: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  emptyColumnText: {
    fontSize: 14,
    color: '#8E8E93',
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  companyName: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#1C1B1F',
    flex: 1,
  },
  priorityDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginLeft: 8,
  },
  position: {
    fontSize: 14,
    color: '#49454F',
    marginBottom: 8,
    lineHeight: 18,
  },
  cardMeta: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  location: {
    fontSize: 12,
    color: '#8E8E93',
    flex: 1,
  },
  date: {
    fontSize: 12,
    color: '#8E8E93',
  },
  salary: {
    fontSize: 12,
    color: '#6750A4',
    fontWeight: '500',
    marginBottom: 8,
  },
  tagsContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  tag: {
    backgroundColor: '#EADDFF',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 8,
    marginRight: 4,
  },
  tagText: {
    fontSize: 10,
    color: '#21005D',
  },
  moreTags: {
    fontSize: 10,
    color: '#8E8E93',
  },
  cardActions: {
    borderTopWidth: 1,
    borderTopColor: '#F3EDF7',
    paddingTop: 8,
    marginTop: 8,
  },
  actionsLabel: {
    fontSize: 12,
    color: '#49454F',
    marginBottom: 6,
  },
  actionButton: {
    borderWidth: 1,
    borderRadius: 12,
    paddingHorizontal: 8,
    paddingVertical: 4,
    marginRight: 6,
  },
  actionButtonText: {
    fontSize: 10,
    fontWeight: '500',
  },
});

export default KanbanScreen;
