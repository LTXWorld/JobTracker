import { describe, it, expect, vi, beforeEach } from 'vitest'
import { AuthAPI } from '../../src/api/auth'
import request from '../../src/api/request'
import type { LoginCredentials, RegisterData, UpdateProfileData } from '../../src/types/auth'

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

  describe('登录流程', () => {
    it('login 成功返回认证信息', async () => {
      const credentials: LoginCredentials = { username: 'tester', password: 'secret' }
      const payload = {
        user: { id: 1, username: 'tester', email: 'tester@example.com' },
        token: 'access-token',
        refresh_token: 'refresh-token',
        expires_at: Date.now() + 3600_000
      }
      vi.mocked(request.post).mockResolvedValueOnce({
        data: { success: true, data: payload, message: '登录成功' }
      })

      const result = await AuthAPI.login(credentials)

      expect(request.post).toHaveBeenCalledWith('/auth/login', credentials)
      expect(result).toEqual(payload)
    })

    it('login 遇到缺失数据时抛出错误', async () => {
      const credentials: LoginCredentials = { username: 'tester', password: 'secret' }
      vi.mocked(request.post).mockResolvedValueOnce({
        data: { success: true, data: null, message: '登录成功' }
      })

      await expect(AuthAPI.login(credentials)).rejects.toThrow('登录失败，服务器响应异常')
    })

    it('login 传播底层网络错误', async () => {
      const credentials: LoginCredentials = { username: 'tester', password: 'secret' }
      vi.mocked(request.post).mockRejectedValueOnce(new Error('网络连接失败'))

      await expect(AuthAPI.login(credentials)).rejects.toThrow('网络连接失败')
    })
  })

  describe('注册流程', () => {
    it('register 成功创建用户', async () => {
      const registerData: RegisterData = {
        username: 'newuser',
        email: 'new@example.com',
        password: 'secret',
        confirmPassword: 'secret'
      }
      const payload = {
        user: { id: 2, username: 'newuser', email: 'new@example.com' },
        token: 'token',
        refresh_token: 'refresh',
        expires_at: Date.now() + 3600_000
      }
      vi.mocked(request.post).mockResolvedValueOnce({
        data: { success: true, data: payload, message: '注册成功' }
      })

      const result = await AuthAPI.register(registerData)

      expect(request.post).toHaveBeenCalledWith('/auth/register', {
        username: 'newuser',
        email: 'new@example.com',
        password: 'secret'
      })
      expect(result).toEqual(payload)
    })

    it('register 两次密码不一致时抛出错误', async () => {
      const registerData = {
        username: 'user',
        email: 'user@example.com',
        password: 'abc',
        confirmPassword: 'def'
      } as RegisterData

      await expect(AuthAPI.register(registerData)).rejects.toThrow('两次输入的密码不匹配')
      expect(request.post).not.toHaveBeenCalled()
    })
  })

  describe('令牌与会话', () => {
    it('refreshToken 成功返回新令牌', async () => {
      const payload = { token: 'new-access', refresh_token: 'new-refresh' }
      vi.mocked(request.post).mockResolvedValueOnce({
        data: { success: true, data: payload, message: 'ok' }
      })

      const result = await AuthAPI.refreshToken('old-refresh')

      expect(request.post).toHaveBeenCalledWith('/auth/refresh', { refresh_token: 'old-refresh' })
      expect(result).toEqual(payload)
    })

    it('validateToken 返回用户信息', async () => {
      const payload = { id: 1, username: 'tester', email: 'tester@example.com' }
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { success: true, data: payload, message: 'valid' }
      })

      const result = await AuthAPI.validateToken()

      expect(request.get).toHaveBeenCalledWith('/auth/validate')
      expect(result).toEqual(payload)
    })
  })

  describe('用户资料', () => {
    it('getUserProfile 返回用户详情', async () => {
      const payload = { id: 1, username: 'tester', email: 'tester@example.com' }
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { success: true, data: payload, message: 'ok' }
      })

      const result = await AuthAPI.getUserProfile()

      expect(request.get).toHaveBeenCalledWith('/auth/profile')
      expect(result).toEqual(payload)
    })

    it('updateProfile 成功更新用户资料', async () => {
      const payload: UpdateProfileData = { username: 'updated', email: 'updated@example.com' }
      vi.mocked(request.put).mockResolvedValueOnce({
        data: { success: true, data: payload, message: 'ok' }
      })

      const result = await AuthAPI.updateProfile(payload)

      expect(request.put).toHaveBeenCalledWith('/auth/profile', payload)
      expect(result).toEqual(payload)
    })

    it('changePassword 成功时不抛错', async () => {
      vi.mocked(request.put).mockResolvedValueOnce({ data: { success: true, message: 'ok' } })

      await expect(AuthAPI.changePassword('old', 'new')).resolves.toBeUndefined()

      expect(request.put).toHaveBeenCalledWith('/auth/password', {
        current_password: 'old',
        new_password: 'new'
      })
    })
  })

  describe('登出与可用性检查', () => {
    it('logout 成功执行', async () => {
      vi.mocked(request.post).mockResolvedValueOnce({ data: { success: true, message: 'ok' } })

      await expect(AuthAPI.logout()).resolves.toBeUndefined()
      expect(request.post).toHaveBeenCalledWith('/auth/logout')
    })

    it('checkUsernameAvailability 返回接口结果', async () => {
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { success: true, data: { available: true } }
      })

      const result = await AuthAPI.checkUsernameAvailability('newuser')

      expect(request.get).toHaveBeenCalledWith('/auth/check-username?username=newuser')
      expect(result).toEqual({ available: true })
    })

    it('checkEmailAvailability 遇到错误时回退默认值', async () => {
      vi.mocked(request.get).mockRejectedValueOnce(new Error('network'))

      const result = await AuthAPI.checkEmailAvailability('exist@example.com')

      expect(result.available).toBe(false)
      expect(result.message).toBe('检查失败')
    })
  })

  describe('统计信息', () => {
    it('getAuthStats 正常返回数据', async () => {
      const stats = {
        total_users: 10,
        active_users: 5,
        new_registrations_today: 2,
        retention_rate: '80%'
      }
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { success: true, data: stats }
      })

      const result = await AuthAPI.getAuthStats()

      expect(request.get).toHaveBeenCalledWith('/auth/stats')
      expect(result).toEqual(stats)
    })
  })
})
