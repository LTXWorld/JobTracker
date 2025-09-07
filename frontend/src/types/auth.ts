// 用户信息接口
export interface User {
  id: number
  username: string
  email: string
  avatar?: string // 头像URL
  full_name?: string // 全名
  bio?: string // 个人简介
  phone?: string // 电话号码
  location?: string // 所在位置
  website?: string // 个人网站
  created_at: string
  updated_at?: string
  last_login_at?: string // 最后登录时间
  is_active?: boolean // 账户是否激活
}

// 登录凭证接口
export interface LoginCredentials {
  username: string
  password: string
  remember_me?: boolean // 记住我选项
  captcha?: string // 验证码
}

// 注册数据接口
export interface RegisterData {
  username: string
  email: string
  password: string
  confirmPassword: string
}

// 认证响应接口
export interface AuthResponse {
  token: string
  refresh_token: string
  user: User
}

// Token响应接口
export interface TokenResponse {
  token: string
  refresh_token?: string // 刷新时可能会返回新的refresh token
}

// 更新用户资料数据接口
export interface UpdateProfileData {
  username?: string
  email?: string
  full_name?: string
  bio?: string
  phone?: string
  location?: string
  website?: string
  avatar?: File | string // 支持文件上传或URL
}

// 修改密码接口
export interface ChangePasswordData {
  current_password: string
  new_password: string
}

// API错误响应接口
export interface AuthError {
  code: number
  message: string
  details?: string
}

// 检查可用性响应接口
export interface AvailabilityResponse {
  available: boolean
  message?: string
}

// 认证统计信息接口
export interface AuthStats {
  total_users: number
  active_users: number
  new_registrations_today: number
  login_attempts_today: number
}

// 用户会话信息接口
export interface UserSession {
  device: string
  location?: string
  ip_address: string
  user_agent: string
  login_time: string
  last_active: string
  is_current: boolean
}

// 密码强度检查接口
export interface PasswordStrength {
  score: number // 0-4
  feedback: string[]
  suggestions: string[]
}

// 认证设置接口
export interface AuthSettings {
  two_factor_enabled: boolean
  session_timeout: number // 分钟
  remember_login_days: number
  password_expiry_days: number
}