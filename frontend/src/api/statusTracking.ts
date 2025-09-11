import request from './request'
import type { 
  StatusHistory,
  StatusAnalytics,
  StatusFlowTemplate,
  UserStatusPreferences,
  UpdateStatusRequest,
  BatchStatusUpdateRequest,
  StatusHistoryParams,
  StatusDurationStats,
  StatusTransitionRule
} from '../types'
import { StatusHelper } from '../types'

/**
 * 状态跟踪API服务
 * 基于后端API实施报告的接口规范实现
 */
export class StatusTrackingAPI {
  
  // ========== 核心状态跟踪API ==========
  
  /**
   * 获取岗位状态历史记录
   * @param applicationId 岗位ID
   * @param params 查询参数
   */
  static async getStatusHistory(applicationId: number, params?: {
    page?: number;
    page_size?: number;
  }): Promise<StatusHistory> {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.append('page', params.page.toString())
    if (params?.page_size) queryParams.append('page_size', params.page_size.toString())
    
    const url = `/api/v1/job-applications/${applicationId}/status-history` + 
                (queryParams.toString() ? `?${queryParams.toString()}` : '')
    
    const response = await request.get(url)
    if (!response.data?.data) {
      throw new Error('获取状态历史失败')
    }

    // 后端返回: { history: [{ new_status, old_status, status_changed_at, duration_minutes, metadata }], total, ... }
    const raw = response.data.data as any
    const rows: Array<any> = Array.isArray(raw.history) ? raw.history : []

    // 转换为前端通用结构，按时间升序
    const sorted = [...rows]
      .filter(Boolean)
      .sort((a, b) => new Date(a.status_changed_at).getTime() - new Date(b.status_changed_at).getTime())

    // 压缩相邻重复状态（防止重复记录导致的双写）
    const compressed: any[] = []
    let prevStatus: string | undefined
    for (const r of sorted) {
      const newStatus = String(r.new_status || r.status || '')
      if (newStatus && newStatus === prevStatus) continue
      compressed.push(r)
      prevStatus = newStatus
    }

    const history = compressed.map((r: any) => ({
      status: r.new_status as any,
      timestamp: r.status_changed_at,
      duration: r.duration_minutes ?? undefined,
      note: r.note ?? r.metadata?.note,
      trigger: r.trigger,
      user_id: r.user_id,
      interview_scheduled: r.metadata?.interview_time || r.metadata?.interview_scheduled || undefined,
      metadata: r.metadata || undefined,
    }))

    const totalDuration = history.reduce((sum: number, h: any) => sum + (Number(h.duration || 0)), 0)
    const lastUpdated = history.length ? history[history.length - 1].timestamp : ''
    const currentStage = history.length ? history[history.length - 1].status : '未知阶段'
    let initialStatus = compressed.length && compressed[0].old_status ? compressed[0].old_status : undefined
    if (!initialStatus) {
      // 默认展示“已投递”起点，便于可视化
      initialStatus = '已投递'
    }

    return {
      history,
      metadata: {
        total_duration: totalDuration,
        status_count: history.length,
        last_updated: lastUpdated,
        current_stage: currentStage,
        initial_status: initialStatus,
      }
    }
  }

  /**
   * 更新岗位状态
   * @param applicationId 岗位ID
   * @param data 状态更新数据
   */
  static async updateStatus(applicationId: number, data: UpdateStatusRequest): Promise<void> {
    const response = await request.post(`/api/v1/job-applications/${applicationId}/status`, data)
    if (!response.data?.success) {
      throw new Error(response.data?.message || '状态更新失败')
    }
  }

  /**
   * 获取状态时间轴数据
   * @param applicationId 岗位ID
   */
  static async getStatusTimeline(applicationId: number): Promise<{
    timeline: Array<{
      status: string;
      timestamp: string;
      duration?: number;
      note?: string;
      is_current: boolean;
    }>;
    total_duration: number;
    current_stage: string;
  }> {
    const response = await request.get(`/api/v1/job-applications/${applicationId}/status-timeline`)
    if (!response.data.data) {
      throw new Error('获取时间轴数据失败')
    }
    return response.data.data
  }

  /**
   * 批量状态更新
   * @param data 批量更新数据
   */
  static async batchUpdateStatus(data: BatchStatusUpdateRequest): Promise<void> {
    // 限制最多100条记录
    if (data.updates.length > 100) {
      throw new Error('批量更新最多支持100条记录')
    }
    
    const response = await request.put('/api/v1/job-applications/status/batch', data)
    if (!response.data?.success) {
      throw new Error(response.data?.message || '批量更新失败')
    }
  }

  // ========== 状态配置管理API ==========

  /**
   * 获取状态流转模板列表
   */
  static async getStatusFlowTemplates(): Promise<StatusFlowTemplate[]> {
    const response = await request.get('/api/v1/status-flow-templates')
    return response.data.data || []
  }

  /**
   * 创建自定义状态流转模板
   * @param template 模板数据
   */
  static async createStatusFlowTemplate(template: Omit<StatusFlowTemplate, 'id' | 'created_at' | 'updated_at'>): Promise<StatusFlowTemplate> {
    const response = await request.post('/api/v1/status-flow-templates', template)
    if (!response.data.data) {
      throw new Error('创建模板失败')
    }
    return response.data.data
  }

  /**
   * 更新状态流转模板
   * @param templateId 模板ID
   * @param template 模板数据
   */
  static async updateStatusFlowTemplate(
    templateId: number, 
    template: Partial<StatusFlowTemplate>
  ): Promise<StatusFlowTemplate> {
    const response = await request.put(`/api/v1/status-flow-templates/${templateId}`, template)
    if (!response.data.data) {
      throw new Error('更新模板失败')
    }
    return response.data.data
  }

  /**
   * 删除状态流转模板
   * @param templateId 模板ID
   */
  static async deleteStatusFlowTemplate(templateId: number): Promise<void> {
    await request.delete(`/api/v1/status-flow-templates/${templateId}`)
  }

  /**
   * 获取用户状态偏好设置
   */
  static async getUserStatusPreferences(): Promise<UserStatusPreferences> {
    const response = await request.get('/api/v1/user-status-preferences')
    if (!response.data.data) {
      throw new Error('获取用户偏好失败')
    }
    return response.data.data
  }

  /**
   * 更新用户状态偏好设置
   * @param preferences 偏好设置数据
   */
  static async updateUserStatusPreferences(
    preferences: Partial<UserStatusPreferences['preference_config']>
  ): Promise<UserStatusPreferences> {
    const response = await request.put('/api/v1/user-status-preferences', { preference_config: preferences })
    if (!response.data.data) {
      throw new Error('更新用户偏好失败')
    }
    return response.data.data
  }

  /**
   * 获取可用状态转换选项
   * @param currentStatus 当前状态
   */
  static async getStatusTransitions(currentStatus: string): Promise<StatusTransitionRule[]> {
    const response = await request.get(`/api/v1/status-transitions/${encodeURIComponent(currentStatus)}`)
    return response.data.data || []
  }

  /**
   * 获取所有状态定义和分类
   */
  static async getStatusDefinitions(): Promise<{
    statuses: Array<{
      status: string;
      category: string;
      description?: string;
      color: string;
      icon: string;
    }>;
    categories: Array<{
      name: string;
      statuses: string[];
      color: string;
    }>;
  }> {
    const response = await request.get('/api/v1/status-definitions')
    if (!response.data.data) {
      throw new Error('获取状态定义失败')
    }
    return response.data.data
  }

  // ========== 数据分析和统计API ==========

  /**
   * 获取用户状态分析数据
   * @param params 查询参数
   */
  static async getStatusAnalytics(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<StatusAnalytics> {
    const queryParams = new URLSearchParams()
    if (params?.start_date) queryParams.append('start_date', params.start_date)
    if (params?.end_date) queryParams.append('end_date', params.end_date)
    
    const url = '/api/v1/job-applications/status-analytics' + 
                (queryParams.toString() ? `?${queryParams.toString()}` : '')
    
    const response = await request.get(url)
    if (!response.data.data) {
      throw new Error('获取分析数据失败')
    }
    return response.data.data
  }

  /**
   * 获取状态趋势分析
   * @param params 查询参数
   */
  static async getStatusTrends(params?: {
    period?: 'week' | 'month' | 'quarter';
    start_date?: string; // 目前后端未使用
    end_date?: string;   // 目前后端未使用
  }): Promise<{
    trends: Array<{
      date: string;
      total_applications: number;
      success_rate: number;
      status_distribution: Record<string, number>;
    }>;
  }> {
    // 兼容后端实现：使用 days 参数，而不是 period
    const period = params?.period
    const days = period === 'week' ? 7 : period === 'quarter' ? 90 : 30

    const url = `/api/v1/job-applications/status-trends?days=${days}`
    const response = await request.get(url)
    if (!response.data?.data) {
      throw new Error('获取趋势数据失败')
    }

    // 后端返回形如 { days: number, trends: Array<{date,status,count}> }，有可能 trends 为 null
    const raw = response.data.data
    const list: Array<{ date: string; status: string; count: number }> = Array.isArray(raw)
      ? raw as any
      : Array.isArray(raw.trends) ? raw.trends : []

    // 按日期聚合：总申请量 + 成功率
    const byDate = new Map<string, { total: number; success: number; dist: Record<string, number> }>()
    for (const item of list) {
      const d = item.date
      if (!byDate.has(d)) byDate.set(d, { total: 0, success: 0, dist: {} })
      const entry = byDate.get(d)!
      entry.total += item.count
      entry.dist[item.status] = (entry.dist[item.status] || 0) + item.count
      // 粗略计算成功率：将“通过/offer/结束”等视为成功
      if (StatusHelper.isPassedStatus(item.status as any)) {
        entry.success += item.count
      }
    }

    const trends = Array.from(byDate.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([date, agg]) => ({
        date,
        total_applications: agg.total,
        success_rate: agg.total > 0 ? agg.success / agg.total : 0,
        status_distribution: agg.dist,
      }))

    return { trends }
  }

  /**
   * 获取流程洞察和建议
   */
  static async getProcessInsights(): Promise<{
    insights: Array<{
      type: 'warning' | 'info' | 'success';
      title: string;
      description: string;
      action_suggestion?: string;
      priority: number;
    }>;
    recommendations: Array<{
      category: string;
      title: string;
      description: string;
      impact_level: 'high' | 'medium' | 'low';
    }>;
  }> {
    const response = await request.get('/api/v1/job-applications/process-insights')
    if (!response.data.data) {
      throw new Error('获取洞察数据失败')
    }
    return response.data.data
  }

  // ========== 增强查询和筛选API ==========

  /**
   * 高级状态和阶段筛选
   * @param params 筛选参数
   */
  static async getApplicationsWithStatusFilter(params: {
    status?: string;
    stage?: string;
    page?: number;
    page_size?: number;
  }): Promise<{
    applications: any[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  }> {
    const queryParams = new URLSearchParams()
    if (params.status) queryParams.append('status', params.status)
    if (params.stage) queryParams.append('stage', params.stage)
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.page_size) queryParams.append('page_size', params.page_size.toString())
    
    const response = await request.get(`/api/v1/applications?${queryParams.toString()}`)
    if (!response.data.data) {
      throw new Error('获取筛选数据失败')
    }
    return response.data
  }

  /**
   * 全文搜索功能
   * @param params 搜索参数
   */
  static async searchApplications(params: {
    q: string;
    filters?: Record<string, any>;
    page?: number;
    page_size?: number;
  }): Promise<{
    applications: any[];
    total: number;
    page: number;
    page_size: number;
    highlighted_fields: string[];
  }> {
    const queryParams = new URLSearchParams()
    queryParams.append('q', params.q)
    if (params.filters) queryParams.append('filters', JSON.stringify(params.filters))
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.page_size) queryParams.append('page_size', params.page_size.toString())
    
    const response = await request.get(`/api/v1/applications/search?${queryParams.toString()}`)
    if (!response.data.data) {
      throw new Error('搜索失败')
    }
    return response.data
  }

  /**
   * 获取仪表板数据聚合
   */
  static async getDashboardData(): Promise<{
    summary: {
      total_applications: number;
      active_applications: number;
      success_rate: number;
      recent_updates: number;
    };
    recent_activity: Array<{
      application_id: number;
      company_name: string;
      position_title: string;
      old_status: string;
      new_status: string;
      timestamp: string;
    }>;
    status_overview: Record<string, {
      count: number;
      percentage: number;
      trend: 'up' | 'down' | 'stable';
    }>;
    upcoming_actions: Array<{
      type: 'interview' | 'follow_up' | 'reminder';
      application_id: number;
      company_name: string;
      scheduled_time: string;
      priority: 'high' | 'medium' | 'low';
    }>;
  }> {
    const response = await request.get('/api/v1/applications/dashboard')
    if (!response.data.data) {
      throw new Error('获取仪表板数据失败')
    }
    return response.data.data
  }
}
