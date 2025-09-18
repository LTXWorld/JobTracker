import axios, { AxiosResponse } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { API_CONFIG, STORAGE_KEYS } from '@config';

// API Types
interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  success: boolean;
  data: {
    user: {
      id: string;
      username: string;
      email: string;
      firstName?: string;
      lastName?: string;
    };
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
  message?: string;
}

interface RefreshTokenResponse {
  success: boolean;
  data: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
}

// Create axios instance
const apiClient = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  async (config) => {
    const token = await AsyncStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = await AsyncStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);

        if (refreshToken) {
          const response = await axios.post(`${API_CONFIG.BASE_URL}/api/v1/auth/refresh`, {
            refreshToken,
          });

          const { accessToken, refreshToken: newRefreshToken } = response.data.data;

          // Update stored tokens
          await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, accessToken);
          await AsyncStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, newRefreshToken);

          // Retry original request
          originalRequest.headers.Authorization = `Bearer ${accessToken}`;
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, logout user
        await AsyncStorage.multiRemove([
          STORAGE_KEYS.AUTH_TOKEN,
          STORAGE_KEYS.REFRESH_TOKEN,
          STORAGE_KEYS.USER_INFO,
        ]);

        // Navigate to login screen
        // This will be handled by the navigation system
      }
    }

    return Promise.reject(error);
  }
);

// Auth API functions
export const authAPI = {
  // Login
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    try {
      const response: AxiosResponse<LoginResponse> = await apiClient.post('/api/v1/auth/login', credentials);

      // Store tokens and user info
      if (response.data.success && response.data.data) {
        const { user, accessToken, refreshToken } = response.data.data;

        await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, accessToken);
        await AsyncStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
        await AsyncStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(user));
      }

      return response.data;
    } catch (error: any) {
      console.error('Login API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '登录失败，请稍后重试',
      };
    }
  },

  // Logout
  logout: async (): Promise<void> => {
    try {
      // Call logout API to invalidate token on server
      await apiClient.post('/api/v1/auth/logout');
    } catch (error) {
      console.error('Logout API Error:', error);
    } finally {
      // Clear local storage regardless of API result
      await AsyncStorage.multiRemove([
        STORAGE_KEYS.AUTH_TOKEN,
        STORAGE_KEYS.REFRESH_TOKEN,
        STORAGE_KEYS.USER_INFO,
      ]);
    }
  },

  // Register
  register: async (userData: {
    username: string;
    email: string;
    password: string;
    firstName?: string;
    lastName?: string;
  }) => {
    try {
      const response = await apiClient.post('/api/v1/auth/register', userData);
      return response.data;
    } catch (error: any) {
      console.error('Register API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '注册失败，请稍后重试',
      };
    }
  },

  // Refresh token
  refreshToken: async (refreshToken: string): Promise<RefreshTokenResponse> => {
    try {
      const response = await apiClient.post('/api/v1/auth/refresh', { refreshToken });
      return response.data;
    } catch (error: any) {
      console.error('Refresh Token API Error:', error);
      throw error;
    }
  },

  // Get current user info
  getCurrentUser: async () => {
    try {
      const response = await apiClient.get('/api/v1/auth/me');
      return response.data;
    } catch (error: any) {
      console.error('Get Current User API Error:', error);
      throw error;
    }
  },

  // Forgot password
  forgotPassword: async (email: string) => {
    try {
      const response = await apiClient.post('/api/v1/auth/forgot-password', { email });
      return response.data;
    } catch (error: any) {
      console.error('Forgot Password API Error:', error);
      throw {
        success: false,
        message: error.response?.data?.message || '发送重置邮件失败，请稍后重试',
      };
    }
  },
};

// General API client for other endpoints
export { apiClient };

export default authAPI;