import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

let request: typeof import('../../src/api/request').default
let axiosCreateMock: ReturnType<typeof vi.fn>
let axiosPostMock: ReturnType<typeof vi.fn>
let capturedRequestInterceptor: ((config: any) => any) | undefined
let capturedRequestError: ((error: any) => any) | undefined
let capturedResponseSuccess: ((response: any) => any) | undefined
let capturedResponseError: ((error: any) => any) | undefined
let messageMocks: any
let addEventListenerSpy: ReturnType<typeof vi.fn>

const setupModule = async () => {
  vi.resetModules()
  capturedRequestInterceptor = undefined
  capturedRequestError = undefined
  capturedResponseSuccess = undefined
  capturedResponseError = undefined

  axiosPostMock = vi.fn()
  axiosCreateMock = vi.fn((config) => ({
    defaults: { ...config },
    interceptors: {
      request: {
        use: vi.fn((onFulfilled, onRejected) => {
          capturedRequestInterceptor = onFulfilled
          capturedRequestError = onRejected
        })
      },
      response: {
        use: vi.fn((onFulfilled, onRejected) => {
          capturedResponseSuccess = onFulfilled
          capturedResponseError = onRejected
        })
      }
    },
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }))

  vi.doMock('axios', () => ({
    default: {
      create: axiosCreateMock,
      post: axiosPostMock
    }
  }))

  messageMocks = {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
  vi.doMock('ant-design-vue', () => ({
    message: messageMocks
  }))

  // Mock navigator / document before模块执行
  Object.defineProperty(globalThis, 'navigator', {
    configurable: true,
    value: { onLine: true }
  })

  addEventListenerSpy = vi.fn()
  const removeEventListenerSpy = vi.fn()
  Object.defineProperty(document, 'addEventListener', {
    configurable: true,
    value: addEventListenerSpy
  })
  Object.defineProperty(document, 'removeEventListener', {
    configurable: true,
    value: removeEventListenerSpy
  })
  Object.defineProperty(document, 'visibilityState', {
    configurable: true,
    value: 'visible'
  })

  Object.defineProperty(window, 'sessionStorage', {
    configurable: true,
    value: {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    }
  })

  vi.useFakeTimers()
  ;({ default: request } = await import('../../src/api/request'))
}

const teardownModule = () => {
  vi.useRealTimers()
  vi.clearAllMocks()
  sessionStorage.clear()
}

describe('api/request', () => {
  beforeEach(async () => {
    await setupModule()
  })

  afterEach(() => {
    teardownModule()
  })

  it('初始化时使用正确的基础配置', () => {
    expect(axiosCreateMock).toHaveBeenCalledTimes(1)
    const config = axiosCreateMock.mock.calls[0][0]
    expect(config.baseURL).toBe('/api')
    expect(config.timeout).toBe(15000)
    expect(config.headers['Content-Type']).toBe('application/json')
    expect(addEventListenerSpy).toHaveBeenCalledWith('visibilitychange', expect.any(Function))
  })

  it('请求拦截器会在有 token 时附加 Authorization 头', async () => {
    expect(capturedRequestInterceptor).toBeDefined()
    const getItemSpy = vi.spyOn(window.sessionStorage, 'getItem').mockReturnValue('mock-token')
    const config = { headers: {} as Record<string, string> }
    await capturedRequestInterceptor?.(config)
    expect(config.headers.Authorization).toBe('Bearer mock-token')
    getItemSpy.mockRestore()
  })

  it('请求拦截器在离线时拒绝请求', async () => {
    Object.defineProperty(navigator, 'onLine', { configurable: true, value: false })
    const config = { headers: {} as Record<string, string> }
    await expect(capturedRequestInterceptor?.(config)).rejects.toThrow('网络连接已断开')
  })

  it('响应拦截器转换成功响应结构', () => {
    expect(capturedResponseSuccess).toBeDefined()
    const mockResponse = {
      data: { code: 200, message: 'ok', data: { foo: 'bar' } },
      config: {},
      request: {}
    }
    const result = capturedResponseSuccess?.(mockResponse)
    expect(result?.data.success).toBe(true)
    expect(result?.data.data).toEqual({ foo: 'bar' })
  })

  it('响应拦截器在非 2xx 数据时抛错', () => {
    const mockResponse = {
      data: { code: 500, message: 'error', data: null },
      config: {},
      request: {}
    }
    expect(() => capturedResponseSuccess?.(mockResponse)).toThrow('error')
  })

  it('响应错误拦截器返回友好错误信息', async () => {
    const error = {
      config: { url: '/test', method: 'get' },
      response: { status: 500, data: { message: '服务器内部错误' } },
      message: 'Internal Error'
    }
    await expect(capturedResponseError?.(error)).rejects.toThrow('服务器内部错误，请稍后重试')
    expect(messageMocks.error).toHaveBeenCalledWith('服务器内部错误，请稍后重试')
  })
})
