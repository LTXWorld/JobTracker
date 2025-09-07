import request from './request'
import type { 
  JobApplication, 
  CreateJobApplicationRequest, 
  UpdateJobApplicationRequest,
  JobApplicationStatistics
} from '../types'

export class JobApplicationAPI {
  // 获取所有投递记录
  static async getAll(): Promise<JobApplication[]> {
    const response = await request.get('/api/v1/applications')
    return response.data.data || []
  }

  // 根据ID获取投递记录
  static async getById(id: number): Promise<JobApplication> {
    const response = await request.get(`/api/v1/applications/${id}`)
    if (!response.data.data) {
      throw new Error('投递记录不存在')
    }
    return response.data.data
  }

  // 创建新的投递记录
  static async create(data: CreateJobApplicationRequest): Promise<JobApplication> {
    const response = await request.post('/api/v1/applications', data)
    if (!response.data.data) {
      throw new Error('创建失败')
    }
    return response.data.data
  }

  // 更新投递记录
  static async update(id: number, data: UpdateJobApplicationRequest): Promise<JobApplication> {
    const response = await request.put(`/api/v1/applications/${id}`, data)
    if (!response.data.data) {
      throw new Error('更新失败')
    }
    return response.data.data
  }

  // 删除投递记录
  static async delete(id: number): Promise<void> {
    await request.delete(`/api/v1/applications/${id}`)
  }

  // 获取统计信息
  static async getStatistics(): Promise<JobApplicationStatistics> {
    const response = await request.get('/api/v1/applications/statistics')
    if (!response.data.data) {
      throw new Error('获取统计信息失败')
    }
    return response.data.data
  }

  // 健康检查
  static async healthCheck(): Promise<boolean> {
    try {
      // 通过前端代理访问健康检查
      const response = await fetch('/health')
      return response.ok
    } catch {
      return false
    }
  }
}