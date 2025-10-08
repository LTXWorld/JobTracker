import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../../src/stores/auth'
import { AuthAPI } from '../../src/api/auth'
import type { AuthResponse, LoginCredentials, RegisterData, TokenResponse } from '../../src/types/auth'

vi.mock('../../src/api/auth', () => ({
  AuthAPI: {
    login: vi.fn(),
    register: vi.fn(),
    refreshToken: vi.fn(),
    logout: vi.fn(),
    getUserProfile: vi.fn(),
    updateProfile: vi.fn(),
    changePassword: vi.fn(),
    validateToken: vi.fn()
  }
}))

vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn()
  }
}))

describe('AuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    const createStorage = () => {
      let store: Record<string, string> = {}
      return {
        getItem: vi.fn((key: string) => (key in store ? store[key] : null)),
        setItem: vi.fn((key: string, value: string) => {
          store[key] = String(value)
        }),
        removeItem: vi.fn((key: string) => {
          delete store[key]
        }),
        clear: vi.fn(() => {
          store = {}
        }),
        key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
        get length() {
          return Object.keys(store).length
        }
      }
    }

    Object.defineProperty(window, 'sessionStorage', {
      configurable: true,
      value: createStorage()
    })
    Object.defineProperty(window, 'localStorage', {
      configurable: true,
      value: createStorage()
    })
    vi.clearAllMocks()
  })

  const createAuthResponse = (): AuthResponse => ({
    token: 'access-token',
    refresh_token: 'refresh-token',
    user: {
      id: 1,
      username: 'tester',
      email: 'tester@example.com',
      created_at: '2024-01-01T00:00:00Z'
    }
  })

  it('初始化时状态为空', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
    expect(store.accessToken).toBeNull()
    expect(store.refreshToken).toBeNull()
    expect(store.isLoggedIn).toBe(false)
  })

  it('initAuth 读取存储的认证信息', () => {
    const response = createAuthResponse()
    sessionStorage.setItem('access_token', response.token)
    localStorage.setItem('refresh_token', response.refresh_token)
    localStorage.setItem('user', JSON.stringify(response.user))
    localStorage.setItem('last_token_validation', '1700000000000')

    const store = useAuthStore()
    store.initAuth()

    expect(store.isAuthenticated).toBe(true)
    expect(store.user?.username).toBe('tester')
    expect(store.accessToken).toBe(response.token)
    expect(store.refreshToken).toBe(response.refresh_token)
  })

  it('login 成功后保存认证信息', async () => {
    const store = useAuthStore()
    const credentials: LoginCredentials = { username: 'tester', password: 'secret' }
    const authResponse = createAuthResponse()
    vi.mocked(AuthAPI.login).mockResolvedValueOnce(authResponse)

    const result = await store.login(credentials)

    expect(result).toBe(true)
    expect(store.isAuthenticated).toBe(true)
    expect(sessionStorage.getItem('access_token')).toBe('access-token')
    expect(localStorage.getItem('refresh_token')).toBe('refresh-token')
  })

  it('login 失败时不改变状态', async () => {
    const store = useAuthStore()
    const credentials: LoginCredentials = { username: 'tester', password: 'secret' }
    vi.mocked(AuthAPI.login).mockRejectedValueOnce(new Error('网络错误'))

    const result = await store.login(credentials)

    expect(result).toBe(false)
    expect(store.isAuthenticated).toBe(false)
  })

  it('logout 会清理认证信息', async () => {
    const store = useAuthStore()
    const authResponse = createAuthResponse()
    store.$patch({
      isAuthenticated: true,
      user: authResponse.user,
      accessToken: authResponse.token,
      refreshToken: authResponse.refresh_token
    })
    sessionStorage.setItem('access_token', authResponse.token)
    localStorage.setItem('refresh_token', authResponse.refresh_token)
    localStorage.setItem('user', JSON.stringify(authResponse.user))
    vi.mocked(AuthAPI.logout).mockResolvedValueOnce()

    await store.logout()

    expect(store.isAuthenticated).toBe(false)
    expect(sessionStorage.getItem('access_token')).toBeNull()
    expect(localStorage.getItem('refresh_token')).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('refreshAccessToken 成功更新 accessToken', async () => {
    const store = useAuthStore()
    const authResponse = createAuthResponse()
    store.$patch({
      isAuthenticated: true,
      user: authResponse.user,
      accessToken: authResponse.token,
      refreshToken: authResponse.refresh_token
    })
    sessionStorage.setItem('access_token', authResponse.token)
    localStorage.setItem('refresh_token', authResponse.refresh_token)
    const newTokens: TokenResponse = { token: 'new-token', refresh_token: 'new-refresh' }
    vi.mocked(AuthAPI.refreshToken).mockResolvedValueOnce(newTokens)

    const result = await store.refreshAccessToken()

    expect(result).toBe(true)
    expect(store.accessToken).toBe('new-token')
    expect(sessionStorage.getItem('access_token')).toBe('new-token')
    expect(localStorage.getItem('refresh_token')).toBe('new-refresh')
  })

  it('shouldValidateToken 根据时间间隔判断', () => {
    const store = useAuthStore()
    const authResponse = createAuthResponse()
    store.$patch({
      isAuthenticated: true,
      user: authResponse.user,
      accessToken: authResponse.token,
      refreshToken: authResponse.refresh_token
    })

    // 刚验证过，无需再次验证
    store.lastTokenValidation = Date.now()
    expect(store.shouldValidateToken()).toBe(false)

    // 距离上次验证超过5分钟，需要验证
    store.lastTokenValidation = Date.now() - 10 * 60 * 1000
    expect(store.shouldValidateToken()).toBe(true)
  })

  it('isTokenRecentlyValid 在宽限期内返回 true', () => {
    const store = useAuthStore()
    const authResponse = createAuthResponse()
    store.$patch({
      isAuthenticated: true,
      user: authResponse.user,
      accessToken: authResponse.token,
      refreshToken: authResponse.refresh_token
    })

    store.lastTokenValidation = Date.now()
    expect(store.isTokenRecentlyValid()).toBe(true)

    store.lastTokenValidation = Date.now() - 10 * 60 * 1000
    expect(store.isTokenRecentlyValid()).toBe(false)
  })
})
