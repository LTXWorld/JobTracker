import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { StatusTrackingAPI } from '../api/statusTracking'
import { StatusHelper } from '../types'
import type { 
  StatusHistory,
  StatusAnalytics,
  StatusFlowTemplate,
  UserStatusPreferences,
  UpdateStatusRequest,
  BatchStatusUpdateRequest,
  StatusTimelineItem,
  StatusStatsCard,
  ApplicationStatus
} from '../types'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

/**
 * 状态跟踪功能的Pinia Store
 * 管理状态历史、分析数据、用户偏好等状态
 */
export const useStatusTrackingStore = defineStore('statusTracking', () => {
  
  // ========== 状态定义 ==========
  
  const loading = ref(false)
  const statusHistories = ref<Map<number, StatusHistory>>(new Map()) // 按application_id缓存历史记录
  const analytics = ref<StatusAnalytics | null>(null)
  const flowTemplates = ref<StatusFlowTemplate[]>([])
  const userPreferences = ref<UserStatusPreferences | null>(null)
  const dashboardData = ref<any>(null)
  const statusDefinitions = ref<any>(null)

  // 加载状态
  const analyticsLoading = ref(false)
  const historyLoading = ref(false)
  const templatesLoading = ref(false)
  const preferencesLoading = ref(false)
  const dashboardLoading = ref(false)

  // ========== 计算属性 ==========
  
  /**
   * 获取状态统计卡片数据
   */
  const statusStatsCards = computed((): StatusStatsCard[] => {
    if (!analytics.value) return []

    // 后端返回格式: { total_applications, status_distribution, success_rate, average_durations }
    const total = (analytics.value as any).total_applications || 0
    const distribution = (analytics.value as any).status_distribution || {}
    const successRate = (analytics.value as any).success_rate || 0

    // 计算活跃申请：进行中状态合计
    let active = 0
    Object.entries(distribution).forEach(([status, count]) => {
      if (StatusHelper.isInProgressStatus(status as any)) active += Number(count as number)
    })

    // 简单估算平均周期（天）：基于平均阶段时长的总和
    const avgDurations = (analytics.value as any).average_durations || {}
    const avgMinutes = Object.values(avgDurations).reduce((sum: number, v: any) => sum + Number(v || 0), 0)
    const avgDays = avgMinutes > 0 ? avgMinutes / 60 / 24 : 0

    return [
      {
        title: '总申请数',
        value: total,
        icon: 'FileTextOutlined',
        color: '#1890ff',
        trend: { direction: 'up', value: '+12%', period: '本月' }
      },
      {
        title: '活跃申请',
        value: active,
        icon: 'ClockCircleOutlined',
        color: '#52c41a',
        trend: { direction: 'stable', value: '0%', period: '本周' }
      },
      {
        title: '成功率',
        value: `${(successRate * 100).toFixed(1)}%`,
        icon: 'TrophyOutlined',
        color: '#faad14',
        trend: { direction: 'up', value: '+2.3%', period: '较上月' }
      },
      {
        title: '平均周期',
        value: `${Math.ceil(avgDays || 0)}天`,
        icon: 'FieldTimeOutlined',
        color: '#722ed1',
        trend: { direction: 'down', value: '-1.2天', period: '较平均' }
      }
    ]
  })

  /**
   * 获取状态分布数据（用于饼图）
   */
  const statusDistributionData = computed(() => {
    if (!analytics.value || !(analytics.value as any).status_distribution) return []

    const dist = (analytics.value as any).status_distribution as Record<string, number>
    const total = Object.values(dist).reduce((sum, v) => sum + Number(v || 0), 0) || 1

    return Object.entries(dist).map(([status, count]) => ({
      name: status,
      value: Number(count || 0),
      percentage: Number(((Number(count || 0) / total) * 100).toFixed(2)),
      color: StatusHelper.getStatusColor(status as ApplicationStatus)
    }))
  })

  /**
   * 获取流程洞察列表
   */
  const processInsights = computed(() => {
    if (!analytics.value) return []
    return analytics.value.insights || []
  })

  // ========== 状态历史相关方法 ==========

  /**
   * 获取特定岗位的状态历史
   * @param applicationId 岗位ID
   * @param forceRefresh 是否强制刷新
   */
  const fetchStatusHistory = async (applicationId: number, forceRefresh = false) => {
    // 如果已有缓存且不强制刷新，直接返回
    if (!forceRefresh && statusHistories.value.has(applicationId)) {
      return statusHistories.value.get(applicationId)!
    }

    historyLoading.value = true
    try {
      const history = await StatusTrackingAPI.getStatusHistory(applicationId)
      statusHistories.value.set(applicationId, history)
      return history
    } catch (error) {
      message.error('获取状态历史失败: ' + (error as Error).message)
      throw error
    } finally {
      historyLoading.value = false
    }
  }

  /**
   * 更新岗位状态
   * @param applicationId 岗位ID
   * @param data 状态更新数据
   */
  const updateApplicationStatus = async (applicationId: number, data: UpdateStatusRequest) => {
    loading.value = true
    try {
      await StatusTrackingAPI.updateStatus(applicationId, data)
      
      // 清除缓存，强制刷新
      statusHistories.value.delete(applicationId)
      
      // 重新获取历史记录
      await fetchStatusHistory(applicationId, true)
      
      message.success('状态更新成功')
      
      // 刷新分析数据
      if (analytics.value) {
        await fetchAnalytics(true)
      }
    } catch (error) {
      message.error('状态更新失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量状态更新
   * @param updates 批量更新数据
   */
  const batchUpdateStatuses = async (updates: BatchStatusUpdateRequest) => {
    loading.value = true
    try {
      await StatusTrackingAPI.batchUpdateStatus(updates)
      
      // 清除相关缓存
      updates.updates.forEach(update => {
        statusHistories.value.delete(update.application_id)
      })
      
      message.success(`成功更新 ${updates.updates.length} 条记录`)
      
      // 刷新分析数据
      if (analytics.value) {
        await fetchAnalytics(true)
      }
    } catch (error) {
      message.error('批量更新失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 将状态历史转换为时间轴数据
   * @param history 状态历史记录
   */
  const convertToTimelineData = (history: StatusHistory): StatusTimelineItem[] => {
    return history.history.map((entry, index) => {
      const isCurrentStatus = index === history.history.length - 1
      const status = entry.status as ApplicationStatus
      
      return {
        id: `${entry.timestamp}_${entry.status}`,
        status,
        timestamp: entry.timestamp,
        duration: entry.duration,
        note: entry.note || undefined,
        is_current: isCurrentStatus,
        is_failed: StatusHelper.isFailedStatus(status),
        is_passed: StatusHelper.isPassedStatus(status),
        icon: getStatusIcon(status),
        color: StatusHelper.getStatusColor(status),
        interview_scheduled: entry.interview_scheduled || undefined
      }
    })
  }

  /**
   * 获取状态图标
   * @param status 状态
   */
  const getStatusIcon = (status: ApplicationStatus): string => {
    const iconMap: Record<string, string> = {
      '已投递': 'SendOutlined',
      '简历筛选中': 'EyeOutlined',
      '简历筛选未通过': 'CloseCircleOutlined',
      '笔试中': 'EditOutlined',
      '笔试通过': 'CheckCircleOutlined',
      '笔试未通过': 'CloseCircleOutlined',
      '一面中': 'UserOutlined',
      '一面通过': 'CheckCircleOutlined',
      '一面未通过': 'CloseCircleOutlined',
      '二面中': 'TeamOutlined',
      '二面通过': 'CheckCircleOutlined',
      '二面未通过': 'CloseCircleOutlined',
      '三面中': 'CrownOutlined',
      '三面通过': 'CheckCircleOutlined',
      '三面未通过': 'CloseCircleOutlined',
      'HR面中': 'ContactsOutlined',
      'HR面通过': 'CheckCircleOutlined',
      'HR面未通过': 'CloseCircleOutlined',
      '待发offer': 'GiftOutlined',
      '已拒绝': 'StopOutlined',
      '已收到offer': 'TrophyOutlined',
      '已接受offer': 'CrownOutlined',
      '流程结束': 'FlagOutlined'
    }
    return iconMap[status] || 'QuestionCircleOutlined'
  }

  // ========== 数据分析相关方法 ==========

  /**
   * 获取状态分析数据
   * @param forceRefresh 是否强制刷新
   * @param dateRange 时间范围
   */
  const fetchAnalytics = async (forceRefresh = false, dateRange?: { start_date?: string; end_date?: string }) => {
    if (!forceRefresh && analytics.value) return analytics.value

    analyticsLoading.value = true
    try {
      const data = await StatusTrackingAPI.getStatusAnalytics(dateRange)
      analytics.value = data
      return data
    } catch (error) {
      message.error('获取分析数据失败: ' + (error as Error).message)
      throw error
    } finally {
      analyticsLoading.value = false
    }
  }

  /**
   * 获取状态趋势数据
   * @param params 查询参数
   */
  const fetchStatusTrends = async (params?: {
    period?: 'week' | 'month' | 'quarter';
    start_date?: string;
    end_date?: string;
  }) => {
    loading.value = true
    try {
      const trends = await StatusTrackingAPI.getStatusTrends(params)
      return trends
    } catch (error) {
      message.error('获取趋势数据失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取流程洞察
   */
  const fetchProcessInsights = async () => {
    try {
      const insights = await StatusTrackingAPI.getProcessInsights()
      return insights
    } catch (error) {
      message.error('获取流程洞察失败: ' + (error as Error).message)
      throw error
    }
  }

  // ========== 配置管理相关方法 ==========

  /**
   * 获取状态流转模板
   * @param forceRefresh 是否强制刷新
   */
  const fetchFlowTemplates = async (forceRefresh = false) => {
    if (!forceRefresh && flowTemplates.value.length > 0) return flowTemplates.value

    templatesLoading.value = true
    try {
      const templates = await StatusTrackingAPI.getStatusFlowTemplates()
      flowTemplates.value = templates
      return templates
    } catch (error) {
      message.error('获取流转模板失败: ' + (error as Error).message)
      throw error
    } finally {
      templatesLoading.value = false
    }
  }

  /**
   * 获取用户偏好设置
   * @param forceRefresh 是否强制刷新
   */
  const fetchUserPreferences = async (forceRefresh = false) => {
    if (!forceRefresh && userPreferences.value) return userPreferences.value

    preferencesLoading.value = true
    try {
      const preferences = await StatusTrackingAPI.getUserStatusPreferences()
      userPreferences.value = preferences
      return preferences
    } catch (error) {
      message.error('获取用户偏好失败: ' + (error as Error).message)
      throw error
    } finally {
      preferencesLoading.value = false
    }
  }

  /**
   * 更新用户偏好设置
   * @param preferences 偏好设置
   */
  const updateUserPreferences = async (preferences: Partial<UserStatusPreferences['preference_config']>) => {
    loading.value = true
    try {
      const updated = await StatusTrackingAPI.updateUserStatusPreferences(preferences)
      userPreferences.value = updated
      message.success('偏好设置更新成功')
      return updated
    } catch (error) {
      message.error('更新偏好设置失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取可用状态转换选项
   * @param currentStatus 当前状态
   */
  const getAvailableTransitions = async (currentStatus: ApplicationStatus) => {
    try {
      const transitions = await StatusTrackingAPI.getStatusTransitions(currentStatus)
      return transitions
    } catch (error) {
      message.error('获取状态转换选项失败: ' + (error as Error).message)
      throw error
    }
  }

  // ========== 仪表板数据方法 ==========

  /**
   * 获取仪表板数据
   * @param forceRefresh 是否强制刷新
   */
  const fetchDashboardData = async (forceRefresh = false) => {
    if (!forceRefresh && dashboardData.value) return dashboardData.value

    dashboardLoading.value = true
    try {
      const data = await StatusTrackingAPI.getDashboardData()
      dashboardData.value = data
      return data
    } catch (error) {
      message.error('获取仪表板数据失败: ' + (error as Error).message)
      throw error
    } finally {
      dashboardLoading.value = false
    }
  }

  /**
   * 获取状态定义
   */
  const fetchStatusDefinitions = async () => {
    if (statusDefinitions.value) return statusDefinitions.value

    try {
      const definitions = await StatusTrackingAPI.getStatusDefinitions()
      statusDefinitions.value = definitions
      return definitions
    } catch (error) {
      message.error('获取状态定义失败: ' + (error as Error).message)
      throw error
    }
  }

  // ========== 工具方法 ==========

  /**
   * 格式化持续时间
   * @param minutes 分钟数
   */
  const formatDuration = (minutes?: number): string => {
    if (!minutes) return '0分钟'
    
    const days = Math.floor(minutes / (24 * 60))
    const hours = Math.floor((minutes % (24 * 60)) / 60)
    const mins = minutes % 60
    
    if (days > 0) {
      return `${days}天${hours > 0 ? hours + '小时' : ''}`
    } else if (hours > 0) {
      return `${hours}小时${mins > 0 ? mins + '分钟' : ''}`
    } else {
      return `${mins}分钟`
    }
  }

  /**
   * 格式化时间戳
   * @param timestamp 时间戳
   * @param format 格式化模式
   */
  const formatTimestamp = (timestamp: string, format = 'YYYY-MM-DD HH:mm'): string => {
    return dayjs(timestamp).format(format)
  }

  /**
   * 计算时间差
   * @param startTime 开始时间
   * @param endTime 结束时间
   */
  const calculateDuration = (startTime: string, endTime?: string): number => {
    const start = dayjs(startTime)
    const end = endTime ? dayjs(endTime) : dayjs()
    return end.diff(start, 'minute')
  }

  /**
   * 清除所有缓存
   */
  const clearCache = () => {
    statusHistories.value.clear()
    analytics.value = null
    dashboardData.value = null
    flowTemplates.value = []
    userPreferences.value = null
    statusDefinitions.value = null
  }

  return {
    // 状态
    loading,
    statusHistories,
    analytics,
    flowTemplates,
    userPreferences,
    dashboardData,
    statusDefinitions,
    
    // 加载状态
    analyticsLoading,
    historyLoading,
    templatesLoading,
    preferencesLoading,
    dashboardLoading,
    
    // 计算属性
    statusStatsCards,
    statusDistributionData,
    processInsights,
    
    // 状态历史方法
    fetchStatusHistory,
    updateApplicationStatus,
    batchUpdateStatuses,
    convertToTimelineData,
    
    // 分析数据方法
    fetchAnalytics,
    fetchStatusTrends,
    fetchProcessInsights,
    
    // 配置管理方法
    fetchFlowTemplates,
    fetchUserPreferences,
    updateUserPreferences,
    getAvailableTransitions,
    
    // 仪表板方法
    fetchDashboardData,
    fetchStatusDefinitions,
    
    // 工具方法
    formatDuration,
    formatTimestamp,
    calculateDuration,
    clearCache
  }
})
