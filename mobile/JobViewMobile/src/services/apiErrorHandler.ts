import { AxiosError } from 'axios';

// Error types
export interface ApiError {
  code: string;
  message: string;
  details?: any;
  statusCode?: number;
  timestamp: Date;
  retryable: boolean;
}

export interface RetryConfig {
  maxRetries: number;
  baseDelay: number;
  maxDelay: number;
  backoffFactor: number;
  retryableErrors: string[];
}

export class ApiErrorHandler {
  private static defaultRetryConfig: RetryConfig = {
    maxRetries: 3,
    baseDelay: 1000,
    maxDelay: 30000,
    backoffFactor: 2,
    retryableErrors: [
      'NETWORK_ERROR',
      'TIMEOUT_ERROR',
      'SERVER_ERROR',
      'RATE_LIMITED',
    ],
  };

  static isRetryableError(error: ApiError): boolean {
    return this.defaultRetryConfig.retryableErrors.includes(error.code);
  }

  static parseError(error: any): ApiError {
    const timestamp = new Date();

    // Axios error
    if (error.isAxiosError) {
      const axiosError = error as AxiosError;
      const statusCode = axiosError.response?.status;
      const responseData = axiosError.response?.data as any;

      let code = 'UNKNOWN_ERROR';
      let message = '未知错误';
      let retryable = false;

      if (axiosError.code === 'ECONNABORTED') {
        code = 'TIMEOUT_ERROR';
        message = '请求超时，请稍后重试';
        retryable = true;
      } else if (axiosError.code === 'NETWORK_ERROR' || !axiosError.response) {
        code = 'NETWORK_ERROR';
        message = '网络连接失败，请检查网络设置';
        retryable = true;
      } else if (statusCode) {
        switch (statusCode) {
          case 400:
            code = 'BAD_REQUEST';
            message = responseData?.message || '请求参数错误';
            retryable = false;
            break;
          case 401:
            code = 'UNAUTHORIZED';
            message = '认证失败，请重新登录';
            retryable = false;
            break;
          case 403:
            code = 'FORBIDDEN';
            message = '权限不足，无法执行此操作';
            retryable = false;
            break;
          case 404:
            code = 'NOT_FOUND';
            message = '请求的资源不存在';
            retryable = false;
            break;
          case 409:
            code = 'CONFLICT';
            message = responseData?.message || '数据冲突';
            retryable = false;
            break;
          case 422:
            code = 'VALIDATION_ERROR';
            message = responseData?.message || '数据验证失败';
            retryable = false;
            break;
          case 429:
            code = 'RATE_LIMITED';
            message = '请求过于频繁，请稍后重试';
            retryable = true;
            break;
          case 500:
          case 502:
          case 503:
          case 504:
            code = 'SERVER_ERROR';
            message = '服务器错误，请稍后重试';
            retryable = true;
            break;
          default:
            code = 'HTTP_ERROR';
            message = responseData?.message || `HTTP错误 ${statusCode}`;
            retryable = statusCode >= 500;
        }
      }

      return {
        code,
        message,
        details: responseData,
        statusCode,
        timestamp,
        retryable,
      };
    }

    // Network or other errors
    if (error.message?.includes('Network')) {
      return {
        code: 'NETWORK_ERROR',
        message: '网络连接失败，请检查网络设置',
        timestamp,
        retryable: true,
      };
    }

    // Generic error
    return {
      code: 'UNKNOWN_ERROR',
      message: error.message || '未知错误',
      details: error,
      timestamp,
      retryable: false,
    };
  }

  static async executeWithRetry<T>(
    operation: () => Promise<T>,
    config: Partial<RetryConfig> = {}
  ): Promise<T> {
    const retryConfig = { ...this.defaultRetryConfig, ...config };
    let lastError: ApiError;

    for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error: any) {
        lastError = this.parseError(error);

        // Don't retry if it's the last attempt or error is not retryable
        if (attempt === retryConfig.maxRetries || !this.isRetryableError(lastError)) {
          throw lastError;
        }

        // Calculate delay with exponential backoff
        const delay = Math.min(
          retryConfig.baseDelay * Math.pow(retryConfig.backoffFactor, attempt),
          retryConfig.maxDelay
        );

        // Add jitter to prevent thundering herd
        const jitter = Math.random() * 0.1 * delay;
        const totalDelay = delay + jitter;

        console.log(`API call failed (attempt ${attempt + 1}), retrying in ${totalDelay}ms...`, lastError);
        await this.delay(totalDelay);
      }
    }

    throw lastError!;
  }

  private static delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  static getErrorMessage(error: any): string {
    if (error && typeof error === 'object' && 'message' in error) {
      return error.message;
    }
    if (typeof error === 'string') {
      return error;
    }
    return '未知错误';
  }

  static getErrorCode(error: any): string {
    if (error && typeof error === 'object' && 'code' in error) {
      return error.code;
    }
    return 'UNKNOWN_ERROR';
  }

  static isNetworkError(error: any): boolean {
    const code = this.getErrorCode(error);
    return ['NETWORK_ERROR', 'TIMEOUT_ERROR'].includes(code);
  }

  static isAuthError(error: any): boolean {
    const code = this.getErrorCode(error);
    return ['UNAUTHORIZED', 'FORBIDDEN'].includes(code);
  }

  static isServerError(error: any): boolean {
    const code = this.getErrorCode(error);
    return code === 'SERVER_ERROR';
  }

  static isValidationError(error: any): boolean {
    const code = this.getErrorCode(error);
    return ['BAD_REQUEST', 'VALIDATION_ERROR'].includes(code);
  }
}

// Utility function for API calls with error handling
export const apiCall = <T>(
  operation: () => Promise<T>,
  retryConfig?: Partial<RetryConfig>
): Promise<T> => {
  return ApiErrorHandler.executeWithRetry(operation, retryConfig);
};

// Error boundary for API errors
export class ApiErrorBoundary {
  private static errorHandlers = new Map<string, (error: ApiError) => void>();

  static registerErrorHandler(errorCode: string, handler: (error: ApiError) => void) {
    this.errorHandlers.set(errorCode, handler);
  }

  static unregisterErrorHandler(errorCode: string) {
    this.errorHandlers.delete(errorCode);
  }

  static handleError(error: ApiError) {
    const handler = this.errorHandlers.get(error.code);
    if (handler) {
      handler(error);
    } else {
      // Default error handling
      console.error('Unhandled API error:', error);
    }
  }
}

export default ApiErrorHandler;