// 投递状态枚举 - 与后端保持一致
export const ApplicationStatus = {
  // 基础状态
  APPLIED: '已投递',
  RESUME_SCREENING: '简历筛选中',
  RESUME_SCREENING_FAIL: '简历筛选未通过',
  
  // 笔试状态
  WRITTEN_TEST: '笔试中',
  WRITTEN_TEST_PASS: '笔试通过',
  WRITTEN_TEST_FAIL: '笔试未通过',
  
  // 一面状态
  FIRST_INTERVIEW: '一面中',
  FIRST_PASS: '一面通过',
  FIRST_FAIL: '一面未通过',
  
  // 二面状态
  SECOND_INTERVIEW: '二面中',
  SECOND_PASS: '二面通过',
  SECOND_FAIL: '二面未通过',
  
  // 三面状态
  THIRD_INTERVIEW: '三面中',
  THIRD_PASS: '三面通过',
  THIRD_FAIL: '三面未通过',
  
  // HR面状态
  HR_INTERVIEW: 'HR面中',
  HR_PASS: 'HR面通过',
  HR_FAIL: 'HR面未通过',
  
  // 最终状态
  OFFER_WAITING: '待发offer',
  REJECTED: '已拒绝',
  OFFER_RECEIVED: '已收到offer',
  OFFER_ACCEPTED: '已接受offer',
  PROCESS_FINISHED: '流程结束'
} as const

export type ApplicationStatus = typeof ApplicationStatus[keyof typeof ApplicationStatus]

// 状态分类辅助函数
export const StatusHelper = {
  // 失败状态
  isFailedStatus: (status: ApplicationStatus): boolean => {
    const failedStatuses: ApplicationStatus[] = [
      ApplicationStatus.RESUME_SCREENING_FAIL,
      ApplicationStatus.WRITTEN_TEST_FAIL,
      ApplicationStatus.FIRST_FAIL,
      ApplicationStatus.SECOND_FAIL,
      ApplicationStatus.THIRD_FAIL,
      ApplicationStatus.HR_FAIL,
      ApplicationStatus.REJECTED
    ]
    return failedStatuses.includes(status)
  },

  // 进行中状态
  isInProgressStatus: (status: ApplicationStatus): boolean => {
    const inProgressStatuses: ApplicationStatus[] = [
      ApplicationStatus.APPLIED,
      ApplicationStatus.RESUME_SCREENING,
      ApplicationStatus.WRITTEN_TEST,
      ApplicationStatus.FIRST_INTERVIEW,
      ApplicationStatus.SECOND_INTERVIEW,
      ApplicationStatus.THIRD_INTERVIEW,
      ApplicationStatus.HR_INTERVIEW
    ]
    return inProgressStatuses.includes(status)
  },

  // 通过状态
  isPassedStatus: (status: ApplicationStatus): boolean => {
    const passedStatuses: ApplicationStatus[] = [
      ApplicationStatus.WRITTEN_TEST_PASS,
      ApplicationStatus.FIRST_PASS,
      ApplicationStatus.SECOND_PASS,
      ApplicationStatus.THIRD_PASS,
      ApplicationStatus.HR_PASS,
      ApplicationStatus.OFFER_WAITING,
      ApplicationStatus.OFFER_RECEIVED,
      ApplicationStatus.OFFER_ACCEPTED,
      ApplicationStatus.PROCESS_FINISHED
    ]
    return passedStatuses.includes(status)
  },

  // 获取状态颜色
  getStatusColor: (status: ApplicationStatus): string => {
    const failedStatuses: ApplicationStatus[] = [
      ApplicationStatus.RESUME_SCREENING_FAIL,
      ApplicationStatus.WRITTEN_TEST_FAIL,
      ApplicationStatus.FIRST_FAIL,
      ApplicationStatus.SECOND_FAIL,
      ApplicationStatus.THIRD_FAIL,
      ApplicationStatus.HR_FAIL,
      ApplicationStatus.REJECTED
    ]
    
    const inProgressStatuses: ApplicationStatus[] = [
      ApplicationStatus.APPLIED,
      ApplicationStatus.RESUME_SCREENING,
      ApplicationStatus.WRITTEN_TEST,
      ApplicationStatus.FIRST_INTERVIEW,
      ApplicationStatus.SECOND_INTERVIEW,
      ApplicationStatus.THIRD_INTERVIEW,
      ApplicationStatus.HR_INTERVIEW
    ]
    
    const passedStatuses: ApplicationStatus[] = [
      ApplicationStatus.WRITTEN_TEST_PASS,
      ApplicationStatus.FIRST_PASS,
      ApplicationStatus.SECOND_PASS,
      ApplicationStatus.THIRD_PASS,
      ApplicationStatus.HR_PASS,
      ApplicationStatus.OFFER_WAITING,
      ApplicationStatus.OFFER_RECEIVED,
      ApplicationStatus.OFFER_ACCEPTED,
      ApplicationStatus.PROCESS_FINISHED
    ]
    
    if (failedStatuses.includes(status)) return 'red'
    if (inProgressStatuses.includes(status)) return 'blue'
    if (passedStatuses.includes(status)) return 'green'
    return 'default'
  },

  // 获取状态类别标签
  getStatusCategory: (status: ApplicationStatus): '进行中' | '已通过' | '已失败' => {
    if (StatusHelper.isFailedStatus(status)) return '已失败'
    if (StatusHelper.isInProgressStatus(status)) return '进行中'
    return '已通过'
  }
}

// 统计数据接口
export interface JobApplicationStatistics {
  total_applications: number
  in_progress: number
  passed: number
  failed: number
  pass_rate: string
  status_breakdown: Record<string, number>
}

// 投递记录接口
export interface JobApplication {
  id: number;
  company_name: string;
  position_title: string;
  application_date: string;
  status: ApplicationStatus;
  job_description?: string | null;
  salary_range?: string | null;
  work_location?: string | null;
  contact_info?: string | null;
  notes?: string | null;
  interview_time?: string | null; // 面试时间
  reminder_time?: string | null; // 提醒时间
  reminder_enabled?: boolean; // 是否启用提醒
  follow_up_date?: string | null; // 跟进日期
  hr_name?: string | null; // HR姓名
  hr_phone?: string | null; // HR电话
  hr_email?: string | null; // HR邮箱
  interview_location?: string | null; // 面试地点
  interview_type?: string | null; // 面试类型
  created_at: string;
  updated_at: string;
}

// 创建投递记录请求
export interface CreateJobApplicationRequest {
  company_name: string;
  position_title: string;
  application_date?: string;
  status?: ApplicationStatus;
  job_description?: string;
  salary_range?: string;
  work_location?: string;
  contact_info?: string;
  notes?: string;
  interview_time?: string;
  reminder_time?: string;
  reminder_enabled?: boolean;
  follow_up_date?: string;
  hr_name?: string;
  hr_phone?: string;
  hr_email?: string;
  interview_location?: string;
  interview_type?: string;
}

// 更新投递记录请求
export interface UpdateJobApplicationRequest {
  company_name?: string;
  position_title?: string;
  application_date?: string;
  status?: ApplicationStatus;
  job_description?: string;
  salary_range?: string;
  work_location?: string;
  contact_info?: string;
  notes?: string;
  interview_time?: string;
  reminder_time?: string;
  reminder_enabled?: boolean;
  follow_up_date?: string;
  hr_name?: string;
  hr_phone?: string;
  hr_email?: string;
  interview_location?: string;
  interview_type?: string;
}

// API响应格式
export interface APIResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

// 筛选条件
export interface FilterOptions {
  status?: ApplicationStatus;
  company?: string;
  dateRange?: [string, string];
}

// 提醒类型
export interface Reminder {
  id: number;
  application_id: number;
  type: 'interview' | 'follow_up';
  reminder_time: string;
  is_sent: boolean;
  company_name: string;
  position_title: string;
  message?: string;
}

// ======== 状态跟踪功能类型定义 ========

// 状态历史记录条目
export interface StatusHistoryEntry {
  status: ApplicationStatus;
  timestamp: string;
  duration?: number | null; // 停留时长（分钟）
  note?: string | null;
  trigger?: 'manual' | 'auto' | 'system';
  user_id?: number;
  interview_scheduled?: string | null; // 面试安排时间
  metadata?: Record<string, any>; // 额外元数据
}

// 状态历史记录
export interface StatusHistory {
  history: StatusHistoryEntry[];
  metadata: {
    total_duration: number; // 总时长（分钟）
    status_count: number; // 状态变更次数
    last_updated: string; // 最后更新时间
    current_stage: string; // 当前阶段
    initial_status?: ApplicationStatus; // 初始状态（可选）
  };
}

// 状态持续时间统计
export interface StatusDurationStats {
  duration_stats: Record<ApplicationStatus, {
    total_minutes: number;
    percentage: number;
  }>;
  milestones: {
    first_response?: string; // 首次回复时间
    first_interview?: string; // 首次面试时间
    offer_time?: string; // 收到offer时间
  };
  analytics: {
    average_response_time: number; // 平均响应时间（分钟）
    total_process_time: number; // 总流程时间（分钟）
    success_probability: number; // 成功概率
  };
}

// 状态流转模板
export interface StatusFlowTemplate {
  id: number;
  name: string;
  description?: string;
  flow_config: {
    stages: Array<{
      name: string;
      statuses: ApplicationStatus[];
      transitions: ApplicationStatus[];
      estimated_duration?: number; // 预计时长（天）
    }>;
    rules: {
      auto_transitions?: Array<{
        from: ApplicationStatus;
        to: ApplicationStatus;
        conditions: Record<string, any>;
      }>;
      reminders?: Array<{
        status: ApplicationStatus;
        delay_days: number;
        message: string;
      }>;
    };
  };
  is_default: boolean;
  created_by?: number;
  created_at: string;
  updated_at: string;
}

// 用户状态偏好设置
export interface UserStatusPreferences {
  id: number;
  user_id: number;
  preference_config: {
    default_template_id?: number;
    notification_settings: {
      email_enabled: boolean;
      push_enabled: boolean;
      reminder_frequency: 'daily' | 'weekly' | 'custom';
    };
    display_preferences: {
      show_durations: boolean;
      show_probabilities: boolean;
      timeline_compact: boolean;
      kanban_show_counts: boolean;
    };
    auto_reminder_rules: Array<{
      status: ApplicationStatus;
      delay_days: number;
      enabled: boolean;
    }>;
  };
  created_at: string;
  updated_at: string;
}

// 状态分析数据
export interface StatusAnalytics {
  summary: {
    total_applications: number;
    active_applications: number;
    success_rate: number;
    average_process_time: number; // 平均流程时长（天）
  };
  status_distribution: Record<ApplicationStatus, {
    count: number;
    percentage: number;
    average_duration: number;
  }>;
  trends: {
    period: 'week' | 'month' | 'quarter';
    data: Array<{
      date: string;
      applications: number;
      success_rate: number;
    }>;
  };
  insights: Array<{
    type: 'warning' | 'info' | 'success';
    title: string;
    description: string;
    action_suggestion?: string;
  }>;
}

// 状态转换规则
export interface StatusTransitionRule {
  from: ApplicationStatus;
  to: ApplicationStatus[];
  conditions?: Record<string, any>;
  auto_transition?: boolean;
  estimated_duration?: number; // 预计时长（天）
}

// 状态跟踪API请求/响应类型

// 更新状态请求
export interface UpdateStatusRequest {
  status: ApplicationStatus;
  note?: string;
  metadata?: Record<string, any>;
  interview_scheduled?: string;
}

// 批量状态更新请求
export interface BatchStatusUpdateRequest {
  updates: Array<{
    application_id: number;
    status: ApplicationStatus;
    note?: string;
  }>;
}

// 状态历史查询参数
export interface StatusHistoryParams {
  application_id?: number;
  status?: ApplicationStatus;
  start_date?: string;
  end_date?: string;
  page?: number;
  page_size?: number;
}

// 状态流转时间轴项目
export interface StatusTimelineItem {
  id: string;
  status: ApplicationStatus;
  timestamp: string;
  duration?: number;
  note?: string;
  is_current: boolean;
  is_failed: boolean;
  is_passed: boolean;
  icon: string;
  color: string;
  interview_scheduled?: string;
}

// 看板拖拽事件数据
export interface DragChangeEvent {
  added?: { element: JobApplication; newIndex: number };
  removed?: { element: JobApplication; oldIndex: number };
  moved?: { element: JobApplication; newIndex: number; oldIndex: number };
}

// 状态统计卡片数据
export interface StatusStatsCard {
  title: string;
  value: number | string;
  icon: string;
  color: string;
  trend?: {
    direction: 'up' | 'down' | 'stable';
    value: string;
    period: string;
  };
}

// 状态筛选选项
export interface StatusFilterOptions {
  statuses?: ApplicationStatus[];
  dateRange?: [string, string];
  companies?: string[];
  only_active?: boolean;
  only_failed?: boolean;
  sort_by?: 'date' | 'company' | 'status' | 'duration';
  sort_order?: 'asc' | 'desc';
}
