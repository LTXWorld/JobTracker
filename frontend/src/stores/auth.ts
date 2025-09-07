import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import { AuthAPI } from '../api/auth'
import type { 
  User, 
  LoginCredentials, 
  RegisterData, 
  AuthResponse,
  UpdateProfileData 
} from '../types/auth'

// Token验证常量
const TOKEN_VALIDATION_INTERVAL = 5 * 60 * 1000 // 5分钟
const TOKEN_GRACE_PERIOD = 2 * 60 * 1000 // 2分钟宽限期

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const isAuthenticated = ref<boolean>(false)
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const loading = ref<boolean>(false)
  const lastTokenValidation = ref<number>(0) // 上次token验证时间
  const tokenValidationAttempts = ref<number>(0) // 验证失败尝试次数

  // 计算属性
  const isLoggedIn = computed(() => isAuthenticated.value && !!user.value)
  const userName = computed(() => user.value?.username || '')
  const userEmail = computed(() => user.value?.email || '')

  // 初始化认证状态
  const initAuth = () => {
    const storedAccessToken = sessionStorage.getItem('access_token')
    const storedRefreshToken = localStorage.getItem('refresh_token')
    const storedUser = localStorage.getItem('user')
    const storedLastValidation = localStorage.getItem('last_token_validation')

    if (storedAccessToken && storedRefreshToken && storedUser) {
      try {
        accessToken.value = storedAccessToken
        refreshToken.value = storedRefreshToken
        user.value = JSON.parse(storedUser)
        isAuthenticated.value = true
        lastTokenValidation.value = storedLastValidation ? parseInt(storedLastValidation) : 0
      } catch (error) {
        console.error('初始化认证状态失败:', error)
        clearAuth()
      }
    }
  }

  // 清除认证数据
  const clearAuth = () => {
    isAuthenticated.value = false
    user.value = null
    accessToken.value = null
    refreshToken.value = null
    lastTokenValidation.value = 0
    tokenValidationAttempts.value = 0
    
    // 清除本地存储
    sessionStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    localStorage.removeItem('last_token_validation')
    // 保留记住的用户名，不在清除认证时删除
  }

  // 保存认证数据
  const saveAuth = (authResponse: AuthResponse) => {
    accessToken.value = authResponse.token
    refreshToken.value = authResponse.refresh_token
    user.value = authResponse.user
    isAuthenticated.value = true
    lastTokenValidation.value = Date.now() // 记录验证时间
    tokenValidationAttempts.value = 0 // 重置失败次数

    // 存储到本地
    sessionStorage.setItem('access_token', authResponse.token)
    localStorage.setItem('refresh_token', authResponse.refresh_token)
    localStorage.setItem('user', JSON.stringify(authResponse.user))
    localStorage.setItem('last_token_validation', lastTokenValidation.value.toString())
  }

  // 用户登录
  const login = async (credentials: LoginCredentials): Promise<boolean> => {
    loading.value = true
    try {
      const response = await AuthAPI.login(credentials)
      saveAuth(response)
      
      // 处理记住我功能
      if (credentials.remember_me) {
        localStorage.setItem('remembered_username', credentials.username)
      } else {
        localStorage.removeItem('remembered_username')
      }
      
      message.success('登录成功')
      return true
    } catch (error) {
      message.error(`登录失败: ${(error as Error).message}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 用户注册
  const register = async (userData: RegisterData): Promise<boolean> => {
    loading.value = true
    try {
      const response = await AuthAPI.register(userData)
      saveAuth(response)
      message.success('注册成功，已自动登录')
      return true
    } catch (error) {
      message.error(`注册失败: ${(error as Error).message}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 用户登出
  const logout = async (): Promise<void> => {
    loading.value = true
    try {
      if (refreshToken.value) {
        await AuthAPI.logout()
      }
      clearAuth()
      message.success('已成功登出')
    } catch (error) {
      console.error('登出失败:', error)
      // 即使登出API失败，也要清除本地数据
      clearAuth()
      message.warning('登出完成，但服务器连接异常')
    } finally {
      loading.value = false
    }
  }

  // 刷新访问令牌
  const refreshAccessToken = async (): Promise<boolean> => {
    if (!refreshToken.value) {
      return false
    }

    try {
      const response = await AuthAPI.refreshToken(refreshToken.value)
      accessToken.value = response.token
      sessionStorage.setItem('access_token', response.token)
      
      // 如果响应包含新的刷新令牌，也要更新
      if (response.refresh_token) {
        refreshToken.value = response.refresh_token
        localStorage.setItem('refresh_token', response.refresh_token)
      }
      
      return true
    } catch (error) {
      console.error('刷新令牌失败:', error)
      clearAuth()
      return false
    }
  }

  // 获取用户信息
  const fetchUserProfile = async (): Promise<void> => {
    if (!isAuthenticated.value) return
    
    loading.value = true
    try {
      const profile = await AuthAPI.getUserProfile()
      user.value = profile
      localStorage.setItem('user', JSON.stringify(profile))
    } catch (error) {
      console.error('获取用户信息失败:', error)
      message.error('获取用户信息失败')
    } finally {
      loading.value = false
    }
  }

  // 更新用户资料
  const updateProfile = async (data: UpdateProfileData): Promise<boolean> => {
    if (!isAuthenticated.value || !user.value) return false
    
    loading.value = true
    try {
      const updatedUser = await AuthAPI.updateProfile(data)
      user.value = updatedUser
      localStorage.setItem('user', JSON.stringify(updatedUser))
      message.success('资料更新成功')
      return true
    } catch (error) {
      message.error(`更新资料失败: ${(error as Error).message}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 修改密码
  const changePassword = async (currentPassword: string, newPassword: string): Promise<boolean> => {
    if (!isAuthenticated.value) return false
    
    loading.value = true
    try {
      await AuthAPI.changePassword(currentPassword, newPassword)
      message.success('密码修改成功，请重新登录')
      clearAuth()
      return true
    } catch (error) {
      message.error(`修改密码失败: ${(error as Error).message}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 验证当前token是否有效
  const validateToken = async (): Promise<boolean> => {
    if (!accessToken.value) return false
    
    try {
      await AuthAPI.validateToken()
      lastTokenValidation.value = Date.now()
      tokenValidationAttempts.value = 0
      localStorage.setItem('last_token_validation', lastTokenValidation.value.toString())
      return true
    } catch (error) {
      console.error('Token验证失败:', error)
      tokenValidationAttempts.value++
      return false
    }
  }

  // 检查是否需要验证token（智能验证策略）
  const shouldValidateToken = (): boolean => {
    if (!accessToken.value) return false
    
    const now = Date.now()
    const timeSinceLastValidation = now - lastTokenValidation.value
    
    // 如果上次验证在5分钟内，且没有验证失败记录，则不需要重新验证
    if (timeSinceLastValidation < TOKEN_VALIDATION_INTERVAL && tokenValidationAttempts.value === 0) {
      return false
    }
    
    return true
  }

  // 检查token是否在最近是有效的（用于网络错误容错）
  const isTokenRecentlyValid = (): boolean => {
    if (!accessToken.value) return false
    
    const now = Date.now()
    const timeSinceLastValidation = now - lastTokenValidation.value
    
    // 如果上次验证在2分钟内且成功，则认为最近是有效的
    return timeSinceLastValidation < TOKEN_GRACE_PERIOD && tokenValidationAttempts.value === 0
  }

  return {
    // 状态
    isAuthenticated,
    user,
    accessToken,
    refreshToken,
    loading,
    lastTokenValidation,
    
    // 计算属性
    isLoggedIn,
    userName,
    userEmail,
    
    // 方法
    initAuth,
    clearAuth,
    login,
    register,
    logout,
    refreshAccessToken,
    fetchUserProfile,
    updateProfile,
    changePassword,
    validateToken,
    shouldValidateToken,
    isTokenRecentlyValid
  }
})