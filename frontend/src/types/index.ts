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