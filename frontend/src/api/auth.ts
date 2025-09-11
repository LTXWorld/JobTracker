import request from './request'
import type { 
  LoginCredentials, 
  RegisterData, 
  AuthResponse, 
  TokenResponse, 
  User,
  UpdateProfileData,
  AvailabilityResponse,
  AuthStats,
  UserSession
} from '../types/auth'

/**
 * 认证相关的API服务类
 */
export class AuthAPI {
  // 基础API路径 - 与后端路由保持一致
  private static readonly AUTH_BASE_URL = '/api/auth'

  /**
   * 用户登录
   */
  static async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await request.post(`${this.AUTH_BASE_URL}/login`, credentials)
    if (!response.data.data) {
      throw new Error('登录失败，服务器响应异常')
    }
    return response.data.data
  }

  /**
   * 用户注册
   */
  static async register(userData: RegisterData): Promise<AuthResponse> {
    // 确保密码匹配
    if (userData.password !== userData.confirmPassword) {
      throw new Error('两次输入的密码不匹配')
    }
    
    // 移除确认密码字段，后端不需要这个字段
    const { confirmPassword, ...registrationData } = userData
    
    const response = await request.post(`${this.AUTH_BASE_URL}/register`, registrationData)
    if (!response.data.data) {
      throw new Error('注册失败，服务器响应异常')
    }
    return response.data.data
  }

  /**
   * 刷新访问令牌
   */
  static async refreshToken(refreshToken: string): Promise<TokenResponse> {
    const response = await request.post(`${this.AUTH_BASE_URL}/refresh`, {
      refresh_token: refreshToken
    })
    if (!response.data.data) {
      throw new Error('刷新令牌失败')
    }
    return response.data.data
  }

  /**
   * 获取用户信息
   */
  static async getUserProfile(): Promise<User> {
    const response = await request.get(`${this.AUTH_BASE_URL}/profile`)
    if (!response.data.data) {
      throw new Error('获取用户信息失败')
    }
    return response.data.data
  }

  /**
   * 更新用户资料
   */
  static async updateProfile(data: UpdateProfileData): Promise<User> {
    const response = await request.put(`${this.AUTH_BASE_URL}/profile`, data)
    if (!response.data.data) {
      throw new Error('更新用户资料失败')
    }
    return response.data.data
  }

  /**
   * 修改密码
   */
  static async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    // 与后端路由保持一致: /api/auth/change-password
    const response = await request.put(`${this.AUTH_BASE_URL}/change-password`, {
      current_password: currentPassword,
      new_password: newPassword
    })
    // 响应拦截器已将 200/201 统一转换为 success=true
    if (!response.data?.success) {
      throw new Error(response.data?.message || '修改密码失败')
    }
  }

  /**
   * 用户登出
   */
  static async logout(): Promise<void> {
    const response = await request.post(`${this.AUTH_BASE_URL}/logout`)
    // 对于成功响应，拦截器会包装为 { success: true }
    if (!response.data?.success) {
      throw new Error(response.data?.message || '登出失败')
    }
  }

  /**
   * 验证访问令牌
   */
  static async validateToken(): Promise<User> {
    const response = await request.get(`${this.AUTH_BASE_URL}/validate`)
    if (!response.data.data) {
      throw new Error('Token验证失败')
    }
    return response.data.data
  }

  /**
   * 上传用户头像
   */
  static async uploadAvatar(file: File): Promise<{ avatar_url: string }> {
    const formData = new FormData()
    formData.append('avatar', file)
    
    const response = await request.post(`${this.AUTH_BASE_URL}/avatar`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    if (!response.data.data) {
      throw new Error('上传头像失败')
    }
    return response.data.data
  }

  /**
   * 获取用户会话信息
   */
  static async getUserSessions(): Promise<UserSession[]> {
    try {
      const response = await request.get(`${this.AUTH_BASE_URL}/sessions`)
      return response.data.data || []
    } catch (error) {
      console.error('获取用户会话信息失败:', error)
      return []
    }
  }

  /**
   * 退出指定会话
   */
  static async terminateSession(sessionId: string): Promise<void> {
    const response = await request.delete(`${this.AUTH_BASE_URL}/sessions/${sessionId}`)
    if (!response.data?.success) {
      throw new Error(response.data?.message || '退出会话失败')
    }
  }

  /**
   * 检查用户名是否可用
   */
  static async checkUsernameAvailability(username: string): Promise<AvailabilityResponse> {
    try {
      const response = await request.get(`${this.AUTH_BASE_URL}/check-username?username=${encodeURIComponent(username)}`)
      return response.data.data || { available: false }
    } catch (error) {
      console.error('检查用户名可用性失败:', error)
      return { available: false, message: '检查失败' }
    }
  }

  /**
   * 检查邮箱是否可用
   */
  static async checkEmailAvailability(email: string): Promise<AvailabilityResponse> {
    try {
      const response = await request.get(`${this.AUTH_BASE_URL}/check-email?email=${encodeURIComponent(email)}`)
      return response.data.data || { available: false }
    } catch (error) {
      console.error('检查邮箱可用性失败:', error)
      return { available: false, message: '检查失败' }
    }
  }

  /**
   * 获取认证统计信息
   */
  static async getAuthStats(): Promise<AuthStats> {
    try {
      const response = await request.get(`${this.AUTH_BASE_URL}/stats`)
      return response.data.data || {
        total_users: 0,
        active_users: 0,
        new_registrations_today: 0,
        login_attempts_today: 0
      }
    } catch (error) {
      console.error('获取认证统计失败:', error)
      return {
        total_users: 0,
        active_users: 0,
        new_registrations_today: 0,
        login_attempts_today: 0
      }
    }
  }

  /**
   * 认证服务健康检查
   */
  static async healthCheck(): Promise<boolean> {
    try {
      const response = await request.get(`${this.AUTH_BASE_URL}/health`)
      return response.data?.success === true
    } catch {
      return false
    }
  }
}
