import axios, { type AxiosResponse } from 'axios'
import { message } from 'ant-design-vue'
import type { APIResponse } from '../types'

// 网络状态检测和重连机制
let networkCheckInterval: NodeJS.Timeout | null = null
let isNetworkChecking = false

const startNetworkMonitoring = () => {
  if (networkCheckInterval) return
  
  networkCheckInterval = setInterval(() => {
    if (!navigator.onLine && !isNetworkChecking) {
      isNetworkChecking = true
      message.warning('网络连接已断开，正在尝试重连...')
    } else if (navigator.onLine && isNetworkChecking) {
      isNetworkChecking = false
      message.success('网络连接已恢复')
    }
  }, 5000)
}

const stopNetworkMonitoring = () => {
  if (networkCheckInterval) {
    clearInterval(networkCheckInterval)
    networkCheckInterval = null
  }
}

// 启动网络监控
startNetworkMonitoring()

// 页面可见性变化时的处理
document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible') {
    startNetworkMonitoring()
  } else {
    stopNetworkMonitoring()
  }
})

// 网络状态检测
const checkNetworkStatus = (): boolean => {
  return navigator.onLine
}

// 获取友好的错误信息
const getFriendlyErrorMessage = (error: any): string => {
  if (!navigator.onLine) {
    return '网络连接已断开，请检查网络设置'
  }
  
  if (error.code === 'NETWORK_ERROR' || error.message === 'Network Error') {
    return '网络连接失败，请稍后重试'
  }
  
  if (error.code === 'ECONNABORTED' || error.message.includes('timeout')) {
    return '请求超时，请检查网络连接'
  }
  
  return error.response?.data?.message || error.message || '请求失败'
}

// 创建axios实例
const request = axios.create({
  baseURL: 'http://localhost:8010',
  timeout: 15000, // 增加超时时间到15秒
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求重试队列
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: any) => void
}> = []

// 处理队列中的请求
const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token)
    }
  })
  
  failedQueue = []
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 检查网络状态
    if (!checkNetworkStatus()) {
      message.error('网络连接已断开，请检查网络设置')
      return Promise.reject(new Error('网络连接已断开'))
    }
    
    // 从 sessionStorage 获取 access token
    const accessToken = sessionStorage.getItem('access_token')
    
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    
    // 添加请求时间戳，用于调试
    ;(config as any).metadata = { startTime: new Date() }
    
    return config
  },
  (error) => {
    const errorMessage = getFriendlyErrorMessage(error)
    message.error(errorMessage)
    return Promise.reject(new Error(errorMessage))
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // 对于文件下载（blob）响应，直接返回原始响应，避免按JSON格式解析
    const respType = (response.request && response.request.responseType) || response.config.responseType
    if (respType === 'blob') {
      return response
    }

    const data: APIResponse = response.data as any
    if (data && (data.code === 200 || data.code === 201)) {
      // 转换后端响应格式为前端期望的格式
      return {
        ...response,
        data: {
          success: true,
          message: data.message,
          data: data.data
        }
      }
    } else {
      throw new Error((data as any)?.message || '请求失败')
    }
  },
  async (error) => {
    const originalRequest = error.config
    
    // 处理 401 未授权错误 - 智能处理策略
    if (error.response?.status === 401 && !originalRequest._retry) {
      // 检查是否为非关键请求，如果是，则不立即跳转登录
      const isNonCriticalRequest = originalRequest.url?.includes('/statistics') || 
                                   originalRequest.url?.includes('/profile') ||
                                   originalRequest.url?.includes('/validate')
      
      if (isRefreshing) {
        // 如果正在刷新token，将请求加入队列
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then((token) => {
          // 更新请求头中的token
          if (token) {
            originalRequest.headers.Authorization = `Bearer ${token}`
          }
          return request(originalRequest)
        }).catch(err => {
          // 对于非关键请求，返回错误但不跳转登录
          if (isNonCriticalRequest) {
            console.warn('非关键请求失败，不影响用户操作:', originalRequest.url)
            return Promise.reject(new Error('请求失败，请稍后重试'))
          }
          return Promise.reject(err)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      const refreshToken = localStorage.getItem('refresh_token')
      
      if (refreshToken) {
        try {
          // 尝试刷新token
          const response = await axios.post('http://localhost:8010/api/auth/refresh', {
            refresh_token: refreshToken
          })
          
          const newAccessToken = response.data.data.token
          const newRefreshToken = response.data.data.refresh_token
          
          // 更新存储的token
          sessionStorage.setItem('access_token', newAccessToken)
          if (newRefreshToken) {
            localStorage.setItem('refresh_token', newRefreshToken)
          }
          
          // 更新原始请求的Authorization头
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
          
          // 处理队列中的请求
          processQueue(null, newAccessToken)
          
          // 重试原始请求
          return request(originalRequest)
        } catch (refreshError) {
          // 刷新token失败，根据请求类型决定处理方式
          processQueue(refreshError, null)
          
          if (isNonCriticalRequest) {
            // 非关键请求，不跳转登录，只显示错误
            console.warn('非关键请求token刷新失败:', originalRequest.url)
            return Promise.reject(new Error('认证失败，请稍后重试'))
          } else {
            // 关键请求，清除认证信息并跳转到登录页
            clearAuthData()
            redirectToLogin()
            return Promise.reject(refreshError)
          }
        } finally {
          isRefreshing = false
        }
      } else {
        // 没有刷新token，根据请求类型决定处理方式
        if (isNonCriticalRequest) {
          console.warn('非关键请求没有refresh token:', originalRequest.url)
          return Promise.reject(new Error('认证失败，请稍后重试'))
        } else {
          // 直接跳转登录
          clearAuthData()
          redirectToLogin()
        }
        return Promise.reject(error)
      }
    }

    // 计算请求耗时（用于性能监控）
    const requestDuration = (originalRequest as any).metadata?.startTime 
      ? new Date().getTime() - (originalRequest as any).metadata.startTime.getTime()
      : 0
    
    // 如果请求超过5秒，记录警告
    if (requestDuration > 5000) {
      console.warn(`慢请求检测: ${originalRequest.url} 耗时 ${requestDuration}ms`)
    }

    // 处理其他错误状态码
    let errorMessage = getFriendlyErrorMessage(error)
    
    // 根据状态码显示不同的错误信息
    if (error.response?.status) {
      switch (error.response.status) {
        case 400:
          errorMessage = error.response.data?.message || '请求参数错误'
          break
        case 403:
          errorMessage = '权限不足，无法访问该资源'
          break
        case 404:
          errorMessage = '请求的资源不存在'
          break
        case 409:
          errorMessage = error.response.data?.message || '资源冲突'
          break
        case 422:
          errorMessage = error.response.data?.message || '数据验证失败'
          break
        case 429:
          errorMessage = '请求过于频繁，请稍后重试'
          break
        case 500:
          errorMessage = '服务器内部错误，请稍后重试'
          break
        case 502:
          errorMessage = '网关错误，服务器暂时不可用'
          break
        case 503:
          errorMessage = '服务暂时不可用，请稍后重试'
          break
        case 504:
          errorMessage = '网关超时，请检查网络连接'
          break
      }
    }

    // 显示错误提示（除了401，因为会自动处理）
    if (error.response?.status !== 401) {
      message.error(errorMessage)
    }

    // 记录错误日志用于调试
    console.error('API请求错误:', {
      url: originalRequest?.url,
      method: originalRequest?.method,
      status: error.response?.status,
      duration: requestDuration,
      message: errorMessage,
      error: error.response?.data
    })

    return Promise.reject(new Error(errorMessage))
  }
)

// 清除认证数据
const clearAuthData = () => {
  sessionStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('user')
  localStorage.removeItem('last_token_validation')
}

// 跳转到登录页
const redirectToLogin = () => {
  // 避免在登录页面重复跳转
  if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
    const currentPath = window.location.pathname + window.location.search
    // 使用更温和的跳转方式，避免突然的页面刷新
    if (window.history.length > 1) {
      window.history.replaceState(null, '', `/login?redirect=${encodeURIComponent(currentPath)}`)
    } else {
      window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`
    }
  }
}

export default request
