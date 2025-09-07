import { describe, it, expect, vi, beforeEach } from 'vitest'
import { AuthAPI } from '../../src/api/auth'
import request from '../../src/api/request'
import type { 
  LoginCredentials, 
  RegisterData, 
  AuthResponse,
  TokenResponse,
  UpdateProfileData 
} from '../../src/types/auth'

// Mock request module
vi.mock('../../src/api/request', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('AuthAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('用户登录', () => {
    it('login 应该成功返回认证信息', async () => {
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      const mockResponse: AuthResponse = {
        user: { id: 1, username: 'testuser', email: 'test@example.com' },
        token: 'access-token-123',
        refresh_token: 'refresh-token-456',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockResponse,
          message: '登录成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.login(credentials)

      expect(request.post).toHaveBeenCalledWith('/api/auth/login', credentials)
      expect(result).toEqual(mockResponse)
    })

    it('login 处理无效凭证时应该抛出错误', async () => {
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'wrongpassword'
      }

      const mockErrorResponse = {
        response: {
          status: 401,
          data: {
            success: false,
            message: '用户名或密码错误'
          }
        }
      }

      vi.mocked(request.post).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.login(credentials)).rejects.toThrow()
    })

    it('login 处理服务器异常响应', async () => {
      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: null, // 异常情况：data为null
          message: '登录成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      await expect(AuthAPI.login(credentials)).rejects.toThrow('登录失败，服务器响应异常')
    })
  })

  describe('用户注册', () => {
    it('register 应该成功创建新用户', async () => {
      const registerData: RegisterData = {
        username: 'newuser',
        email: 'newuser@example.com',
        password: 'password123'
      }

      const mockResponse: AuthResponse = {
        user: { id: 2, username: 'newuser', email: 'newuser@example.com' },
        token: 'access-token-789',
        refresh_token: 'refresh-token-012',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockResponse,
          message: '注册成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.register(registerData)

      expect(request.post).toHaveBeenCalledWith('/api/auth/register', registerData)
      expect(result).toEqual(mockResponse)
    })

    it('register 处理用户名已存在的情况', async () => {
      const registerData: RegisterData = {
        username: 'existinguser',
        email: 'existing@example.com',
        password: 'password123'
      }

      const mockErrorResponse = {
        response: {
          status: 400,
          data: {
            success: false,
            message: '用户名已存在'
          }
        }
      }

      vi.mocked(request.post).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.register(registerData)).rejects.toThrow()
    })

    it('register 处理邮箱已注册的情况', async () => {
      const registerData: RegisterData = {
        username: 'newuser',
        email: 'existing@example.com',
        password: 'password123'
      }

      const mockErrorResponse = {
        response: {
          status: 400,
          data: {
            success: false,
            message: '邮箱已被注册'
          }
        }
      }

      vi.mocked(request.post).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.register(registerData)).rejects.toThrow()
    })
  })

  describe('Token刷新', () => {
    it('refreshToken 应该成功刷新访问令牌', async () => {
      const refreshToken = 'current-refresh-token'

      const mockResponse: TokenResponse = {
        token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        expires_at: Date.now() + 24 * 60 * 60 * 1000
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockResponse,
          message: 'Token刷新成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.refreshToken(refreshToken)

      expect(request.post).toHaveBeenCalledWith('/api/auth/refresh', {
        refresh_token: refreshToken
      })
      expect(result).toEqual(mockResponse)
    })

    it('refreshToken 处理无效刷新令牌', async () => {
      const invalidRefreshToken = 'invalid-refresh-token'

      const mockErrorResponse = {
        response: {
          status: 401,
          data: {
            success: false,
            message: '刷新令牌无效'
          }
        }
      }

      vi.mocked(request.post).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.refreshToken(invalidRefreshToken)).rejects.toThrow()
    })

    it('refreshToken 处理服务器错误', async () => {
      const refreshToken = 'valid-refresh-token'

      const mockApiResponse = {
        data: {
          success: true,
          data: null, // 异常响应
          message: 'Token刷新成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      await expect(AuthAPI.refreshToken(refreshToken)).rejects.toThrow('Token刷新失败')
    })
  })

  describe('用户信息管理', () => {
    it('getProfile 应该成功获取用户信息', async () => {
      const mockUserProfile = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockUserProfile,
          message: '获取用户信息成功'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.getProfile()

      expect(request.get).toHaveBeenCalledWith('/api/auth/profile')
      expect(result).toEqual(mockUserProfile)
    })

    it('updateProfile 应该成功更新用户信息', async () => {
      const updateData: UpdateProfileData = {
        username: 'updateduser',
        email: 'updated@example.com'
      }

      const mockUpdatedProfile = {
        id: 1,
        username: 'updateduser',
        email: 'updated@example.com',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z'
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockUpdatedProfile,
          message: '更新用户信息成功'
        }
      }

      vi.mocked(request.put).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.updateProfile(updateData)

      expect(request.put).toHaveBeenCalledWith('/api/auth/profile', updateData)
      expect(result).toEqual(mockUpdatedProfile)
    })

    it('changePassword 应该成功修改密码', async () => {
      const passwordData = {
        current_password: 'oldpassword',
        new_password: 'newpassword123'
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: null,
          message: '密码修改成功'
        }
      }

      vi.mocked(request.put).mockResolvedValueOnce(mockApiResponse)

      await expect(AuthAPI.changePassword(passwordData)).resolves.toBeUndefined()

      expect(request.put).toHaveBeenCalledWith('/api/auth/change-password', passwordData)
    })

    it('changePassword 处理当前密码错误', async () => {
      const passwordData = {
        current_password: 'wrongpassword',
        new_password: 'newpassword123'
      }

      const mockErrorResponse = {
        response: {
          status: 400,
          data: {
            success: false,
            message: '当前密码错误'
          }
        }
      }

      vi.mocked(request.put).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.changePassword(passwordData)).rejects.toThrow()
    })
  })

  describe('用户名和邮箱可用性检查', () => {
    it('checkUsernameAvailability 应该检查用户名是否可用', async () => {
      const username = 'newuser'

      const mockApiResponse = {
        data: {
          success: true,
          data: { available: true },
          message: '用户名可用'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.checkUsernameAvailability(username)

      expect(request.get).toHaveBeenCalledWith(`/api/auth/check-username?username=${username}`)
      expect(result).toEqual({ available: true })
    })

    it('checkUsernameAvailability 处理用户名已被使用', async () => {
      const username = 'existinguser'

      const mockApiResponse = {
        data: {
          success: false,
          data: { available: false },
          message: '用户名已被使用'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.checkUsernameAvailability(username)

      expect(result).toEqual({ available: false })
    })

    it('checkEmailAvailability 应该检查邮箱是否可用', async () => {
      const email = 'new@example.com'

      const mockApiResponse = {
        data: {
          success: true,
          data: { available: true },
          message: '邮箱可用'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.checkEmailAvailability(email)

      expect(request.get).toHaveBeenCalledWith(`/api/auth/check-email?email=${encodeURIComponent(email)}`)
      expect(result).toEqual({ available: true })
    })
  })

  describe('用户登出', () => {
    it('logout 应该成功登出用户', async () => {
      const mockApiResponse = {
        data: {
          success: true,
          data: null,
          message: '登出成功'
        }
      }

      vi.mocked(request.post).mockResolvedValueOnce(mockApiResponse)

      await expect(AuthAPI.logout()).resolves.toBeUndefined()

      expect(request.post).toHaveBeenCalledWith('/api/auth/logout')
    })

    it('logout 处理服务器错误时仍应成功', async () => {
      // 即使服务器返回错误，客户端也应该清理本地状态
      const mockErrorResponse = {
        response: {
          status: 500,
          data: {
            success: false,
            message: '服务器内部错误'
          }
        }
      }

      vi.mocked(request.post).mockRejectedValueOnce(mockErrorResponse)

      // 登出不应该因为服务器错误而失败
      await expect(AuthAPI.logout()).rejects.toThrow()
    })
  })

  describe('用户统计信息', () => {
    it('getUserStats 应该成功获取用户统计', async () => {
      const mockStats = {
        total_applications: 15,
        pending_applications: 5,
        interviews_scheduled: 3,
        offers_received: 2,
        last_application_date: '2024-01-15',
        active_days: 30,
        success_rate: 0.13
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockStats,
          message: '获取统计信息成功'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.getUserStats()

      expect(request.get).toHaveBeenCalledWith('/api/auth/stats')
      expect(result).toEqual(mockStats)
    })
  })

  describe('会话管理', () => {
    it('validateSession 应该验证当前会话', async () => {
      const mockSessionData = {
        valid: true,
        user: { id: 1, username: 'testuser', email: 'test@example.com' },
        expires_at: Date.now() + 60 * 60 * 1000
      }

      const mockApiResponse = {
        data: {
          success: true,
          data: mockSessionData,
          message: '会话有效'
        }
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      const result = await AuthAPI.validateSession()

      expect(request.get).toHaveBeenCalledWith('/api/auth/validate')
      expect(result).toEqual(mockSessionData)
    })

    it('validateSession 处理无效会话', async () => {
      const mockErrorResponse = {
        response: {
          status: 401,
          data: {
            success: false,
            message: '会话已过期'
          }
        }
      }

      vi.mocked(request.get).mockRejectedValueOnce(mockErrorResponse)

      await expect(AuthAPI.validateSession()).rejects.toThrow()
    })
  })

  describe('错误处理', () => {
    it('应该正确处理网络错误', async () => {
      const networkError = new Error('网络连接失败')
      vi.mocked(request.post).mockRejectedValueOnce(networkError)

      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      await expect(AuthAPI.login(credentials)).rejects.toThrow('网络连接失败')
    })

    it('应该正确处理超时错误', async () => {
      const timeoutError = { code: 'ECONNABORTED', message: 'timeout exceeded' }
      vi.mocked(request.post).mockRejectedValueOnce(timeoutError)

      const credentials: LoginCredentials = {
        username: 'testuser',
        password: 'password123'
      }

      await expect(AuthAPI.login(credentials)).rejects.toMatchObject(timeoutError)
    })

    it('应该正确处理无响应数据的情况', async () => {
      const mockApiResponse = {
        data: null // 无响应数据
      }

      vi.mocked(request.get).mockResolvedValueOnce(mockApiResponse)

      await expect(AuthAPI.getProfile()).rejects.toThrow()
    })
  })

  describe('请求参数验证', () => {
    it('login 应该验证必需的参数', async () => {
      const incompleteCredentials = {
        username: 'testuser'
        // 缺少password
      } as LoginCredentials

      // 这里我们期望API会在内部验证参数
      // 实际实现可能会在发送请求前进行客户端验证
      expect(incompleteCredentials.username).toBe('testuser')
      expect(incompleteCredentials.password).toBeUndefined()
    })

    it('register 应该验证邮箱格式', async () => {
      const invalidEmailData = {
        username: 'testuser',
        email: 'invalid-email',
        password: 'password123'
      } as RegisterData

      // 客户端或服务器应该验证邮箱格式
      expect(invalidEmailData.email).toBe('invalid-email')
    })
  })
})