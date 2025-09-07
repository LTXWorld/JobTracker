import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../../src/stores/auth'
import { AuthAPI } from '../../src/api/auth'
import type { AuthResponse, LoginCredentials, RegisterData } from '../../src/types/auth'

// Mock AuthAPI
vi.mock('../../src/api/auth', () => ({
  AuthAPI: {
    login: vi.fn(),
    register: vi.fn(),
    refreshToken: vi.fn(),
    logout: vi.fn(),
    getProfile: vi.fn(),
    updateProfile: vi.fn(),
    changePassword: vi.fn()
  }
}))

// Mock ant-design-vue message
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

describe('AuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    // 清理localStorage和sessionStorage
    localStorage.clear()
    sessionStorage.clear()
    // 清理所有mock
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  describe('初始状态', () => {
    it('应该有正确的初始状态', () => {
      const authStore = useAuthStore()
      
      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
      expect(authStore.refreshToken).toBe(null)
      expect(authStore.loading).toBe(false)
      expect(authStore.isLoggedIn).toBe(false)
      expect(authStore.userName).toBe('')
      expect(authStore.userEmail).toBe('')
    })

    it('应该从本地存储恢复认证状态', () => {
      // 模拟已存储的认证信息
      const mockUser = { id: 1, username: 'testuser', email: 'test@example.com' }
      const mockAccessToken = 'mock-access-token'
      const mockRefreshToken = 'mock-refresh-token'

      sessionStorage.setItem('access_token', mockAccessToken)
      localStorage.setItem('refresh_token', mockRefreshToken)
      localStorage.setItem('user', JSON.stringify(mockUser))
      localStorage.setItem('last_token_validation', Date.now().toString())

      const authStore = useAuthStore()
      authStore.initAuth()

      expect(authStore.isAuthenticated).toBe(true)
      expect(authStore.user).toEqual(mockUser)
      expect(authStore.accessToken).toBe(mockAccessToken)
      expect(authStore.refreshToken).toBe(mockRefreshToken)
      expect(authStore.isLoggedIn).toBe(true)
    })

    it('处理无效的存储数据时应该清理认证状态', () => {
      // 模拟无效的存储数据
      sessionStorage.setItem('access_token', 'token')
      localStorage.setItem('refresh_token', 'token')
      localStorage.setItem('user', 'invalid-json')

      const authStore = useAuthStore()
      authStore.initAuth()

      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
    })
  })

  describe('用户登录', () => {
    it('登录成功应该更新认证状态', async () => {
      const authStore = useAuthStore()
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }
      
      const mockResponse: AuthResponse = {
        user: { id: 1, username: 'testuser', email: 'test@example.com' },
        token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      vi.mocked(AuthAPI.login).mockResolvedValueOnce(mockResponse)

      const result = await authStore.login(credentials)

      expect(result).toBe(true)
      expect(authStore.isAuthenticated).toBe(true)
      expect(authStore.user).toEqual(mockResponse.user)
      expect(authStore.accessToken).toBe(mockResponse.token)
      expect(authStore.refreshToken).toBe(mockResponse.refresh_token)
      
      // 验证存储
      expect(sessionStorage.getItem('access_token')).toBe(mockResponse.token)
      expect(localStorage.getItem('refresh_token')).toBe(mockResponse.refresh_token)
      expect(JSON.parse(localStorage.getItem('user')!)).toEqual(mockResponse.user)
    })

    it('登录失败应该返回false且不更新状态', async () => {
      const authStore = useAuthStore()
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'wrongpassword'
      }

      const mockError = new Error('用户名或密码错误')
      vi.mocked(AuthAPI.login).mockRejectedValueOnce(mockError)

      const result = await authStore.login(credentials)

      expect(result).toBe(false)
      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
    })

    it('登录时应该设置loading状态', async () => {
      const authStore = useAuthStore()
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      let loadingDuringCall = false
      
      vi.mocked(AuthAPI.login).mockImplementationOnce(() => {
        loadingDuringCall = authStore.loading
        return Promise.resolve({
          user: { id: 1, username: 'testuser', email: 'test@example.com' },
          token: 'token',
          refresh_token: 'refresh-token',
          expires_at: Date.now() + 24 * 60 * 60 * 1000
        } as AuthResponse)
      })

      await authStore.login(credentials)

      expect(loadingDuringCall).toBe(true)
      expect(authStore.loading).toBe(false) // 调用完成后应该重置
    })
  })

  describe('用户注册', () => {
    it('注册成功应该更新认证状态', async () => {
      const authStore = useAuthStore()
      const registerData: RegisterData = {
        username: 'newuser',
        email: 'newuser@example.com',
        password: 'password123'
      }
      
      const mockResponse: AuthResponse = {
        user: { id: 1, username: 'newuser', email: 'newuser@example.com' },
        token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      vi.mocked(AuthAPI.register).mockResolvedValueOnce(mockResponse)

      const result = await authStore.register(registerData)

      expect(result).toBe(true)
      expect(authStore.isAuthenticated).toBe(true)
      expect(authStore.user).toEqual(mockResponse.user)
      expect(AuthAPI.register).toHaveBeenCalledWith(registerData)
    })

    it('注册失败应该返回false', async () => {
      const authStore = useAuthStore()
      const registerData: RegisterData = {
        username: 'existinguser',
        email: 'existing@example.com',
        password: 'password123'
      }

      const mockError = new Error('用户名已存在')
      vi.mocked(AuthAPI.register).mockRejectedValueOnce(mockError)

      const result = await authStore.register(registerData)

      expect(result).toBe(false)
      expect(authStore.isAuthenticated).toBe(false)
    })
  })

  describe('Token刷新机制', () => {
    it('应该能成功刷新token', async () => {
      const authStore = useAuthStore()
      
      // 设置初始状态
      authStore.refreshToken = 'current-refresh-token'
      authStore.isAuthenticated = true

      const mockResponse = {
        token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      vi.mocked(AuthAPI.refreshToken).mockResolvedValueOnce(mockResponse)

      const result = await authStore.refreshAccessToken()

      expect(result).toBe(true)
      expect(authStore.accessToken).toBe(mockResponse.token)
      expect(authStore.refreshToken).toBe(mockResponse.refresh_token)
      expect(AuthAPI.refreshToken).toHaveBeenCalledWith('current-refresh-token')
    })

    it('Token刷新失败应该清理认证状态', async () => {
      const authStore = useAuthStore()
      
      // 设置初始状态
      authStore.refreshToken = 'invalid-refresh-token'
      authStore.isAuthenticated = true
      authStore.user = { id: 1, username: 'test', email: 'test@example.com' }

      vi.mocked(AuthAPI.refreshToken).mockRejectedValueOnce(new Error('refresh token无效'))

      const result = await authStore.refreshAccessToken()

      expect(result).toBe(false)
      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
      expect(authStore.refreshToken).toBe(null)
    })

    it('智能验证策略应该正确工作', () => {
      const authStore = useAuthStore()
      
      // 模拟最近验证过的情况
      authStore.lastTokenValidation = Date.now() - 3 * 60 * 1000 // 3分钟前
      authStore.tokenValidationAttempts = 0
      
      const shouldValidate = authStore.shouldValidateToken()
      expect(shouldValidate).toBe(false) // 不应该验证，因为还在5分钟间隔内

      // 模拟需要验证的情况
      authStore.lastTokenValidation = Date.now() - 6 * 60 * 1000 // 6分钟前
      
      const shouldValidateNow = authStore.shouldValidateToken()
      expect(shouldValidateNow).toBe(true) // 应该验证，已超过5分钟间隔
    })

    it('网络错误容错机制应该生效', async () => {
      const authStore = useAuthStore()
      authStore.refreshToken = 'valid-token'
      authStore.isAuthenticated = true

      // 模拟网络错误
      const networkError = new Error('Network Error')
      vi.mocked(AuthAPI.refreshToken).mockRejectedValueOnce(networkError)

      const result = await authStore.refreshAccessToken()

      expect(result).toBe(false)
      // 网络错误时不应该完全清理状态，可能是临时问题
      expect(authStore.tokenValidationAttempts).toBeGreaterThan(0)
    })
  })

  describe('用户登出', () => {
    it('登出应该清理所有认证状态和存储', async () => {
      const authStore = useAuthStore()
      
      // 设置已认证状态
      authStore.isAuthenticated = true
      authStore.user = { id: 1, username: 'test', email: 'test@example.com' }
      authStore.accessToken = 'access-token'
      authStore.refreshToken = 'refresh-token'
      
      sessionStorage.setItem('access_token', 'access-token')
      localStorage.setItem('refresh_token', 'refresh-token')
      localStorage.setItem('user', JSON.stringify(authStore.user))

      vi.mocked(AuthAPI.logout).mockResolvedValueOnce(undefined)

      await authStore.logout()

      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
      expect(authStore.refreshToken).toBe(null)
      
      // 验证存储已清理
      expect(sessionStorage.getItem('access_token')).toBe(null)
      expect(localStorage.getItem('refresh_token')).toBe(null)
      expect(localStorage.getItem('user')).toBe(null)
    })

    it('即使API调用失败也应该清理本地状态', async () => {
      const authStore = useAuthStore()
      
      authStore.isAuthenticated = true
      authStore.user = { id: 1, username: 'test', email: 'test@example.com' }

      vi.mocked(AuthAPI.logout).mockRejectedValueOnce(new Error('服务器错误'))

      await authStore.logout()

      // 即使API失败，本地状态也应该被清理
      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
    })
  })

  describe('计算属性', () => {
    it('isLoggedIn应该正确反映认证状态', () => {
      const authStore = useAuthStore()

      // 初始状态
      expect(authStore.isLoggedIn).toBe(false)

      // 只有isAuthenticated为true
      authStore.isAuthenticated = true
      expect(authStore.isLoggedIn).toBe(false) // 仍然为false，因为没有user

      // 添加用户信息
      authStore.user = { id: 1, username: 'test', email: 'test@example.com' }
      expect(authStore.isLoggedIn).toBe(true)

      // 移除用户信息
      authStore.user = null
      expect(authStore.isLoggedIn).toBe(false)
    })

    it('userName和userEmail应该正确返回用户信息', () => {
      const authStore = useAuthStore()

      expect(authStore.userName).toBe('')
      expect(authStore.userEmail).toBe('')

      authStore.user = { id: 1, username: 'testuser', email: 'test@example.com' }

      expect(authStore.userName).toBe('testuser')
      expect(authStore.userEmail).toBe('test@example.com')
    })
  })

  describe('错误处理', () => {
    it('应该正确处理API错误', async () => {
      const authStore = useAuthStore()
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      const errorMessage = 'API请求失败'
      vi.mocked(AuthAPI.login).mockRejectedValueOnce(new Error(errorMessage))

      const result = await authStore.login(credentials)

      expect(result).toBe(false)
      expect(authStore.loading).toBe(false)
      expect(authStore.isAuthenticated).toBe(false)
    })

    it('应该处理无效的响应数据', async () => {
      const authStore = useAuthStore()
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      // 模拟无效的响应数据
      vi.mocked(AuthAPI.login).mockResolvedValueOnce(null as any)

      const result = await authStore.login(credentials)

      expect(result).toBe(false)
      expect(authStore.isAuthenticated).toBe(false)
    })
  })

  describe('清理功能', () => {
    it('clearAuth应该重置所有状态', () => {
      const authStore = useAuthStore()
      
      // 设置一些状态
      authStore.isAuthenticated = true
      authStore.user = { id: 1, username: 'test', email: 'test@example.com' }
      authStore.accessToken = 'token'
      authStore.refreshToken = 'refresh-token'
      authStore.loading = true
      authStore.lastTokenValidation = Date.now()
      authStore.tokenValidationAttempts = 5

      sessionStorage.setItem('access_token', 'token')
      localStorage.setItem('refresh_token', 'refresh-token')
      localStorage.setItem('user', JSON.stringify(authStore.user))

      authStore.clearAuth()

      expect(authStore.isAuthenticated).toBe(false)
      expect(authStore.user).toBe(null)
      expect(authStore.accessToken).toBe(null)
      expect(authStore.refreshToken).toBe(null)
      expect(authStore.loading).toBe(false)
      expect(authStore.lastTokenValidation).toBe(0)
      expect(authStore.tokenValidationAttempts).toBe(0)
      
      // 验证存储已清理
      expect(sessionStorage.getItem('access_token')).toBe(null)
      expect(localStorage.getItem('refresh_token')).toBe(null)
      expect(localStorage.getItem('user')).toBe(null)
    })
  })
})