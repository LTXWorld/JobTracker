import request from './request'
import type { 
  ExportRequest,
  ExportTask,
  ExportHistoryResponse,
  ExportFormatInfo,
  ExportFieldGroup,
  ExportableField
} from '../types'
import { 
  ExportableFields,
  FieldDisplayNames
} from '../types'

export class ExportAPI {
  // 启动导出任务
  static async startExport(exportRequest: ExportRequest): Promise<ExportTask> {
    const response = await request.post('/api/v1/export/applications', exportRequest)
    if (!response.data.success) {
      throw new Error(response.data.message || '启动导出失败')
    }
    return response.data.data
  }

  // 查询任务状态
  static async getTaskStatus(taskId: string): Promise<ExportTask> {
    const response = await request.get(`/api/v1/export/status/${taskId}`)
    if (!response.data.success) {
      throw new Error(response.data.message || '获取任务状态失败')
    }
    return response.data.data
  }

  // 下载导出文件
  static async downloadFile(taskId: string): Promise<void> {
    try {
      const response = await request.get(`/api/v1/export/download/${taskId}`, {
        responseType: 'blob',
        timeout: 60000 // 增加下载超时时间到60秒
      })
      
      // 从响应头获取文件名
      const contentDisposition = response.headers['content-disposition']
      let filename = '求职投递记录.xlsx'
      
      if (contentDisposition) {
        const filenameMatch = contentDisposition.match(/filename[*]?="([^"]+)"/)
        if (filenameMatch) {
          filename = decodeURIComponent(filenameMatch[1])
        }
      }
      
      // 创建下载链接
      const blob = new Blob([response.data], {
        type: response.headers['content-type'] || 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
      })
      
      const downloadUrl = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = downloadUrl
      link.download = filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      
      // 清理对象URL
      window.URL.revokeObjectURL(downloadUrl)
    } catch (error: any) {
      console.error('文件下载失败:', error)
      throw new Error('文件下载失败，请重试')
    }
  }

  // 获取导出历史
  static async getExportHistory(page: number = 1, pageSize: number = 10): Promise<ExportHistoryResponse['data']> {
    const response = await request.get(`/api/v1/export/history?page=${page}&limit=${pageSize}`)
    if (!response.data.success) {
      throw new Error(response.data.message || '获取导出历史失败')
    }
    return response.data.data
  }

  // 取消导出任务
  static async cancelTask(taskId: string): Promise<void> {
    await request.delete(`/api/v1/export/cancel/${taskId}`)
  }

  // 获取支持的导出格式
  static async getSupportedFormats(): Promise<ExportFormatInfo[]> {
    const response = await request.get('/api/v1/export/formats')
    if (!response.data.success) {
      throw new Error(response.data.message || '获取支持格式失败')
    }
    return response.data.data
  }

  // 获取可导出的字段
  static async getExportableFields(): Promise<ExportFieldGroup[]> {
    const response = await request.get('/api/v1/export/fields')
    if (!response.data.success) {
      throw new Error(response.data.message || '获取导出字段失败')
    }
    return response.data.data
  }

  // 获取导出模板
  static async getExportTemplate(): Promise<any> {
    const response = await request.get('/api/v1/export/template')
    if (!response.data.success) {
      throw new Error(response.data.message || '获取导出模板失败')
    }
    return response.data.data
  }

  // 清理过期的导出文件
  static async cleanupExpiredFiles(): Promise<{ cleaned: number }> {
    const response = await request.post('/api/v1/export/cleanup')
    if (!response.data.success) {
      throw new Error(response.data.message || '清理失败')
    }
    return response.data.data
  }
}

// 导出字段分组配置
export const ExportFieldGroups: ExportFieldGroup[] = [
  {
    group: 'basic',
    label: '基础信息',
    fields: [
      { field: ExportableFields.COMPANY_NAME, label: FieldDisplayNames[ExportableFields.COMPANY_NAME], required: true },
      { field: ExportableFields.POSITION_TITLE, label: FieldDisplayNames[ExportableFields.POSITION_TITLE], required: true },
      { field: ExportableFields.APPLICATION_DATE, label: FieldDisplayNames[ExportableFields.APPLICATION_DATE], required: true },
      { field: ExportableFields.STATUS, label: FieldDisplayNames[ExportableFields.STATUS], required: true },
    ]
  },
  {
    group: 'job_details',
    label: '职位详情',
    fields: [
      { field: ExportableFields.JOB_DESCRIPTION, label: FieldDisplayNames[ExportableFields.JOB_DESCRIPTION] },
      { field: ExportableFields.SALARY_RANGE, label: FieldDisplayNames[ExportableFields.SALARY_RANGE] },
      { field: ExportableFields.WORK_LOCATION, label: FieldDisplayNames[ExportableFields.WORK_LOCATION] },
    ]
  },
  {
    group: 'interview',
    label: '面试信息',
    fields: [
      { field: ExportableFields.INTERVIEW_TIME, label: FieldDisplayNames[ExportableFields.INTERVIEW_TIME] },
      { field: ExportableFields.INTERVIEW_LOCATION, label: FieldDisplayNames[ExportableFields.INTERVIEW_LOCATION] },
      { field: ExportableFields.INTERVIEW_TYPE, label: FieldDisplayNames[ExportableFields.INTERVIEW_TYPE] },
    ]
  },
  {
    group: 'contact',
    label: '联系信息',
    fields: [
      { field: ExportableFields.HR_NAME, label: FieldDisplayNames[ExportableFields.HR_NAME] },
      { field: ExportableFields.HR_PHONE, label: FieldDisplayNames[ExportableFields.HR_PHONE] },
      { field: ExportableFields.HR_EMAIL, label: FieldDisplayNames[ExportableFields.HR_EMAIL] },
      { field: ExportableFields.CONTACT_INFO, label: FieldDisplayNames[ExportableFields.CONTACT_INFO] },
    ]
  },
  {
    group: 'reminders',
    label: '提醒跟进',
    fields: [
      { field: ExportableFields.REMINDER_TIME, label: FieldDisplayNames[ExportableFields.REMINDER_TIME] },
      { field: ExportableFields.FOLLOW_UP_DATE, label: FieldDisplayNames[ExportableFields.FOLLOW_UP_DATE] },
    ]
  },
  {
    group: 'metadata',
    label: '其他信息',
    fields: [
      { field: ExportableFields.NOTES, label: FieldDisplayNames[ExportableFields.NOTES] },
      { field: ExportableFields.CREATED_AT, label: FieldDisplayNames[ExportableFields.CREATED_AT] },
      { field: ExportableFields.UPDATED_AT, label: FieldDisplayNames[ExportableFields.UPDATED_AT] },
    ]
  }
]

// 默认导出字段
export const DefaultExportFields: ExportableField[] = [
  ExportableFields.COMPANY_NAME,
  ExportableFields.POSITION_TITLE,
  ExportableFields.APPLICATION_DATE,
  ExportableFields.STATUS,
  ExportableFields.SALARY_RANGE,
  ExportableFields.WORK_LOCATION,
  ExportableFields.HR_NAME,
  ExportableFields.HR_PHONE,
  ExportableFields.INTERVIEW_TIME,
  ExportableFields.NOTES
]

// 任务状态显示配置
export const TaskStatusConfig = {
  pending: {
    label: '等待处理',
    color: 'default',
    icon: 'ClockCircleOutlined'
  },
  processing: {
    label: '正在处理',
    color: 'processing',
    icon: 'LoadingOutlined'
  },
  completed: {
    label: '已完成',
    color: 'success',
    icon: 'CheckCircleOutlined'
  },
  failed: {
    label: '失败',
    color: 'error',
    icon: 'ExclamationCircleOutlined'
  },
  cancelled: {
    label: '已取消',
    color: 'default',
    icon: 'StopOutlined'
  },
  expired: {
    label: '已过期',
    color: 'warning',
    icon: 'ClockCircleOutlined'
  }
}

// 工具函数：格式化文件大小
export const formatFileSize = (bytes?: string | number): string => {
  if (!bytes) return '--'
  
  const size = typeof bytes === 'string' ? parseFloat(bytes) : bytes
  if (size === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(size) / Math.log(k))
  
  return parseFloat((size / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// 工具函数：格式化剩余时间
export const formatRemainingTime = (expiresAt?: string): string => {
  if (!expiresAt) return '--'
  
  const now = new Date()
  const expiry = new Date(expiresAt)
  const diff = expiry.getTime() - now.getTime()
  
  if (diff <= 0) return '已过期'
  
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  
  if (hours > 24) {
    const days = Math.floor(hours / 24)
    return `${days}天`
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}