import { defineStore } from 'pinia'
import { ref, computed, h } from 'vue'
import { StatusTrackingAPI } from '../api/statusTracking'
import { StatusHelper, ApplicationStatus as ApplicationStatusEnum } from '../types'
import type { 
  StatusHistory,
  StatusAnalytics,
  StatusFlowTemplate,
  UserStatusPreferences,
  UpdateStatusRequest,
  BatchStatusUpdateRequest,
  StatusTimelineItem,
  StatusStatsCard,
  ApplicationStatus,
  StatusUpdateResult,
  StatusUndoRequest,
  InterviewExperienceRating,
  InterviewExperienceSubmission,
  InterviewExperienceMetadata,
  InterviewExperienceCaptureContext,
  StatusHistoryEntryMetadata
} from '../types'
import { message, notification, Button } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { useJobApplicationStore } from './jobApplication'
import { triggerOfferCelebration, triggerEncouragement } from '../utils/offerCelebration'

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

  type UndoContext = {
    historyId: number
    expiresAt: Dayjs
    notificationKey: string
    timer?: ReturnType<typeof setInterval>
    version?: number
    onUndo?: () => Promise<void> | void
  }

  const pendingUndo = ref<Map<number, UndoContext>>(new Map())
  const jobApplicationStore = useJobApplicationStore()
  const lastInterviewExperience = ref<InterviewExperienceSubmission | null>(null)

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
    const avgDays = Number(avgMinutes) > 0 ? Number(avgMinutes) / 60 / 24 : 0

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

  const clearUndoContext = (applicationId: number, options?: { silent?: boolean }) => {
    const ctx = pendingUndo.value.get(applicationId)
    if (!ctx) return

    if (ctx.timer) {
      clearInterval(ctx.timer)
    }

    if (!options?.silent) {
      notification.close(ctx.notificationKey)
    }

    pendingUndo.value.delete(applicationId)
  }

  const undoStatusUpdate = async (applicationId: number) => {
    const ctx = pendingUndo.value.get(applicationId)
    if (!ctx) {
      message.warning('当前没有可撤销的状态更新')
      return
    }

    const payload: StatusUndoRequest = {
      history_id: ctx.historyId
    }
    if (ctx.version !== undefined) {
      payload.version = ctx.version
    }

    if (ctx.timer) {
      clearInterval(ctx.timer)
      ctx.timer = undefined
    }

    loading.value = true
    try {
      const result = await StatusTrackingAPI.undoStatus(applicationId, payload)
      clearUndoContext(applicationId)

      statusHistories.value.delete(applicationId)
      await fetchStatusHistory(applicationId, true)
      await jobApplicationStore.fetchApplications()
      if (analytics.value) {
        await fetchAnalytics(true)
      }

      if (ctx.onUndo) {
        await ctx.onUndo()
      }

      const reverted = result?.job?.status
      message.success(reverted ? `已撤销至「${reverted}」` : '已撤销最近一次状态更新')
    } catch (error) {
      clearUndoContext(applicationId)
      const msg = (error as Error)?.message || '撤销状态失败'
      if (msg.includes('undo window expired')) {
        message.error('撤销失败：倒计时已结束，请刷新后重试')
      } else if (msg.includes('version conflict')) {
        message.error('撤销失败：数据版本冲突，请刷新后重试')
      } else if (msg.includes('status mismatch')) {
        message.error('撤销失败：当前状态已变化，请刷新后重试')
      } else if (msg.includes('no undoable history') || msg.includes('UNDO_HISTORY_NOT_FOUND')) {
        message.error('撤销失败：找不到可撤销的记录')
      } else {
        message.error('撤销失败: ' + msg)
      }
    } finally {
      loading.value = false
    }
  }

  const scheduleUndoNotification = (applicationId: number, result: StatusUpdateResult, options?: { onUndo?: () => Promise<void> | void }) => {
    if (!result || !result.history_id || !result.undo_available_until) {
      message.success('状态更新成功')
      return
    }

    const deadline = dayjs(result.undo_available_until)
    if (!deadline.isValid() || deadline.isBefore(dayjs())) {
      message.success('状态更新成功')
      return
    }

    clearUndoContext(applicationId)

    const key = `status-undo-${applicationId}-${result.history_id}`
    const context: UndoContext = {
      historyId: result.history_id,
      expiresAt: deadline,
      notificationKey: key,
      version: result.status_version ?? result.job?.status_version ?? undefined,
      onUndo: options?.onUndo
    }

    const openNotification = () => {
      const secondsLeft = Math.max(0, Math.ceil(context.expiresAt.diff(dayjs(), 'second')))
      notification.open({
        key,
        message: '状态更新成功',
        description: `状态已更新为「${result.job?.status ?? '新状态'}」，可在倒计时结束前撤销。`,
        duration: 0,
        btn: h(Button, {
          type: 'link',
          size: 'small',
          disabled: secondsLeft <= 0,
          onClick: () => undoStatusUpdate(applicationId)
        }, {
          default: () => `撤销（${secondsLeft}s）`
        }),
        onClose: () => {
          clearUndoContext(applicationId, { silent: true })
        }
      })
    }

    openNotification()

    context.timer = setInterval(() => {
      const secondsLeft = Math.ceil(context.expiresAt.diff(dayjs(), 'second'))
      if (secondsLeft <= 0) {
        clearUndoContext(applicationId)
        message.info('撤销机会已过期')
        return
      }
      openNotification()
    }, 1000)

    pendingUndo.value.set(applicationId, context)
  }

  /**
   * 更新岗位状态
   * @param applicationId 岗位ID
   * @param data 状态更新数据
   */
  const updateApplicationStatus = async (
    applicationId: number,
    data: UpdateStatusRequest,
    options?: { onUndo?: () => Promise<void> | void }
  ): Promise<StatusUpdateResult | undefined> => {
    loading.value = true
    try {
      const result = await StatusTrackingAPI.updateStatus(applicationId, data)
      const nextStatus = (result.job?.status ?? data.status) as ApplicationStatus | undefined
      if (nextStatus === ApplicationStatusEnum.OFFER_ACCEPTED) {
        message.success('恭喜拿下offer，这段时间辛苦啦！')
        triggerOfferCelebration()
      } else if (nextStatus && StatusHelper.isFailedStatus(nextStatus)) {
        triggerEncouragement()
      }
      
      // 清除缓存，强制刷新
      statusHistories.value.delete(applicationId)
      
      // 重新获取历史记录
      await fetchStatusHistory(applicationId, true)

      scheduleUndoNotification(applicationId, result, options)
      
      // 刷新分析数据
      if (analytics.value) {
        await fetchAnalytics(true)
      }

      return result
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

  const interviewExperienceRatings = new Set<InterviewExperienceRating>(['good', 'average', 'bad'])

  const normalizeExperienceText = (value: unknown): string | undefined => {
    if (typeof value !== 'string') return undefined
    const trimmed = value.trim()
    return trimmed.length > 0 ? trimmed : undefined
  }

  const cacheInterviewExperience = (submission: InterviewExperienceSubmission | null) => {
    lastInterviewExperience.value = submission ? { ...submission } : null
  }

  const getCachedInterviewExperience = (): InterviewExperienceSubmission | null => {
    const cached = lastInterviewExperience.value
    return cached ? { ...cached } : null
  }

  const buildInterviewExperienceMetadata = (
    context: InterviewExperienceCaptureContext,
    submission: InterviewExperienceSubmission
  ): InterviewExperienceMetadata => {
    const note = normalizeExperienceText(submission.note) ?? null
    const metadata: InterviewExperienceMetadata = {
      skip: submission.skip,
      recorded_at: dayjs().toISOString(),
      recorded_by: null,
      from_status: context.fromStatus,
      to_status: context.toStatus,
      note
    }

    if (submission.skip) {
      metadata.skip_reason = normalizeExperienceText(submission.skip_reason) ?? null
    } else {
      metadata.rating = submission.rating
      metadata.skip_reason = null
    }

    return metadata
  }

  const assembleUpdateRequestWithInterviewExperience = (
    base: UpdateStatusRequest,
    context: InterviewExperienceCaptureContext,
    submission?: InterviewExperienceSubmission | null
  ): UpdateStatusRequest => {
    if (!submission) {
      return { ...base }
    }

    const metadata = {
      ...(base.metadata ?? {}),
      interview_experience: buildInterviewExperienceMetadata(context, submission)
    }

    cacheInterviewExperience(submission)

    return {
      ...base,
      metadata
    }
  }

  const normalizeInterviewExperience = (
    raw: StatusHistoryEntryMetadata['interview_experience']
  ): StatusTimelineItem['interviewExperience'] | undefined => {
    if (!raw || typeof raw !== 'object') return undefined
    const data = raw as Record<string, unknown>

    const ratingCandidate = data['rating']
    const rating = typeof ratingCandidate === 'string' && interviewExperienceRatings.has(ratingCandidate as InterviewExperienceRating)
      ? ratingCandidate as InterviewExperienceRating
      : undefined

    const skipValue = data['skip']
    const skip = skipValue !== undefined ? Boolean(skipValue) : false

    const recordedByValue = data['recorded_by']
    const fromStatusValue = data['from_status']
    const toStatusValue = data['to_status']

    return {
      skip,
      rating,
      note: normalizeExperienceText(data['note']),
      skipReason: normalizeExperienceText(data['skip_reason']),
      recordedAt: normalizeExperienceText(data['recorded_at']),
      recordedBy: typeof recordedByValue === 'number' ? recordedByValue : undefined,
      fromStatus: typeof fromStatusValue === 'string' ? fromStatusValue : undefined,
      toStatus: typeof toStatusValue === 'string' ? toStatusValue : undefined,
      raw: { ...data }
    }
  }

  /**
   * 将状态历史转换为时间轴数据
   * @param history 状态历史记录
   */
  const convertToTimelineData = (history: StatusHistory): StatusTimelineItem[] => {
    const mappedEntries = history.history.map((entry, index) => {
      const isCurrentStatus = index === history.history.length - 1
      const status = entry.status as ApplicationStatus
      const metadata = (entry.metadata ?? {}) as StatusHistoryEntryMetadata
      const item: StatusTimelineItem = {
        id: `${entry.timestamp}_${entry.status}`,
        status,
        timestamp: entry.timestamp,
        duration: entry.duration ?? undefined,
        note: entry.note || undefined,
        is_current: isCurrentStatus,
        is_failed: StatusHelper.isFailedStatus(status),
        is_passed: StatusHelper.isPassedStatus(status),
        icon: getStatusIcon(status),
        color: StatusHelper.getStatusColor(status),
        interview_scheduled: entry.interview_scheduled || undefined,
        interviewExperience: normalizeInterviewExperience(metadata['interview_experience'] ?? null)
      }
      return { item, metadata }
    })

    mappedEntries.forEach(({ metadata }, index) => {
      const undoApplied = typeof metadata['undo_applied'] === 'boolean' ? (metadata['undo_applied'] as boolean) : false
      const isUndoRecord = typeof metadata['undo'] === 'boolean' ? (metadata['undo'] as boolean) : false

      if (undoApplied) {
        mappedEntries[index].item.interviewExperience = undefined
      }

      if (isUndoRecord) {
        mappedEntries[index].item.interviewExperience = undefined
        if (index > 0) {
          mappedEntries[index - 1].item.interviewExperience = undefined
        }
      }
    })

    return mappedEntries.map(entry => entry.item)
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
      '已接受offer': 'CrownOutlined',
      '已拒绝offer': 'StopOutlined'
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
    undoStatusUpdate,
    
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
    cacheInterviewExperience,
    getCachedInterviewExperience,
    assembleUpdateRequestWithInterviewExperience,
    clearCache
  }
})
