import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { authAPI } from '../services/authService';
import { User } from '../types';
import { STORAGE_KEYS } from '../config';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  tokens: {
    accessToken: string | null;
    refreshToken: string | null;
    expiresAt: Date | null;
  };
  isLoading: boolean;
  error: string | null;
  biometricAvailable: boolean;
  biometricEnabled: boolean;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  tokens: {
    accessToken: null,
    refreshToken: null,
    expiresAt: null,
  },
  isLoading: false,
  error: null,
  biometricAvailable: false,
  biometricEnabled: false,
};

// Create default preferences
const createDefaultPreferences = () => ({
  theme: 'system' as const,
  language: 'zh' as const,
  notifications: {
    enabled: true,
    interviewReminders: true,
    followUpReminders: true,
    statusUpdates: true,
    sound: true,
    vibration: true,
  },
  biometricEnabled: false,
});

// Async thunks
export const loginUser = createAsyncThunk(
  'auth/loginUser',
  async ({ username, password }: { username: string; password: string }, { rejectWithValue }) => {
    try {
      const response = await authAPI.login({ username, password });

      if (!response.success) {
        return rejectWithValue(response.message || '登录失败');
      }

      if (!response.data) {
        return rejectWithValue('登录数据异常');
      }

      const { user: userData, accessToken, refreshToken, expiresIn } = response.data;
      const expiresAt = new Date(Date.now() + expiresIn * 1000);

      // Create full user object with preferences
      const user: User = {
        ...userData,
        id: userData.id,
        createdAt: new Date(),
        updatedAt: new Date(),
        preferences: createDefaultPreferences(),
      };

      return {
        user,
        accessToken,
        refreshToken,
        expiresAt,
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '登录失败');
    }
  }
);

export const registerUser = createAsyncThunk(
  'auth/registerUser',
  async (userData: {
    username: string;
    email: string;
    password: string;
    firstName?: string;
    lastName?: string;
  }, { rejectWithValue }) => {
    try {
      const response = await authAPI.register(userData);

      if (!response.success) {
        return rejectWithValue(response.message || '注册失败');
      }

      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '注册失败');
    }
  }
);

export const logoutUser = createAsyncThunk(
  'auth/logoutUser',
  async (_, { rejectWithValue }) => {
    try {
      await authAPI.logout();
      return true;
    } catch (error: any) {
      // Even if API call fails, we should still clear local data
      console.warn('Logout API failed:', error);
      return true;
    }
  }
);

export const loadStoredAuth = createAsyncThunk(
  'auth/loadStoredAuth',
  async (_, { rejectWithValue }) => {
    try {
      const [tokenData, userDataStr] = await Promise.all([
        AsyncStorage.getItem(STORAGE_KEYS.AUTH_TOKEN),
        AsyncStorage.getItem(STORAGE_KEYS.USER_INFO),
      ]);

      if (!tokenData || !userDataStr) {
        return null;
      }

      const user: User = JSON.parse(userDataStr);

      // Verify token is still valid by calling the API
      try {
        const response = await authAPI.getCurrentUser();
        if (response.success && response.data) {
          // Update user data with latest from server
          const updatedUser = {
            ...user,
            ...response.data,
            preferences: user.preferences || createDefaultPreferences(),
          };

          await AsyncStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(updatedUser));

          return {
            user: updatedUser,
            accessToken: tokenData,
          };
        }
      } catch (error) {
        // Token is invalid, clear storage
        await AsyncStorage.multiRemove([
          STORAGE_KEYS.AUTH_TOKEN,
          STORAGE_KEYS.REFRESH_TOKEN,
          STORAGE_KEYS.USER_INFO,
        ]);
        return null;
      }

      return null;
    } catch (error: any) {
      return rejectWithValue(error.message || '加载存储的认证信息失败');
    }
  }
);

export const refreshAuthToken = createAsyncThunk(
  'auth/refreshAuthToken',
  async (_, { rejectWithValue, getState }) => {
    try {
      const refreshToken = await AsyncStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);

      if (!refreshToken) {
        return rejectWithValue('无刷新令牌');
      }

      const response = await authAPI.refreshToken(refreshToken);

      if (!response.success) {
        return rejectWithValue(response.message || '刷新令牌失败');
      }

      const { accessToken, refreshToken: newRefreshToken, expiresIn } = response.data;
      const expiresAt = new Date(Date.now() + expiresIn * 1000);

      // Update stored tokens
      await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, accessToken);
      await AsyncStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, newRefreshToken);

      return {
        accessToken,
        refreshToken: newRefreshToken,
        expiresAt,
      };
    } catch (error: any) {
      // Clear stored tokens on refresh failure
      await AsyncStorage.multiRemove([
        STORAGE_KEYS.AUTH_TOKEN,
        STORAGE_KEYS.REFRESH_TOKEN,
        STORAGE_KEYS.USER_INFO,
      ]);
      return rejectWithValue(error.message || '刷新令牌失败');
    }
  }
);

export const updateUser = createAsyncThunk(
  'auth/updateUser',
  async (updates: Partial<User>, { rejectWithValue, getState }) => {
    try {
      // This would call a user update API endpoint
      // For now, we'll just update locally
      const state = getState() as any;
      const currentUser = state.auth.user;

      if (!currentUser) {
        return rejectWithValue('用户未登录');
      }

      const updatedUser = {
        ...currentUser,
        ...updates,
        updatedAt: new Date(),
      };

      await AsyncStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(updatedUser));

      return updatedUser;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新用户信息失败');
    }
  }
);

export const forgotPassword = createAsyncThunk(
  'auth/forgotPassword',
  async (email: string, { rejectWithValue }) => {
    try {
      const response = await authAPI.forgotPassword(email);

      if (!response.success) {
        return rejectWithValue(response.message || '发送重置邮件失败');
      }

      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '发送重置邮件失败');
    }
  }
);

export const createAuthSlice = () => createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
    setUser: (state, action: PayloadAction<User | null>) => {
      state.user = action.payload;
      state.isAuthenticated = !!action.payload;
    },
    setTokens: (state, action: PayloadAction<{
      accessToken: string | null;
      refreshToken: string | null;
      expiresAt: Date | null;
    }>) => {
      state.tokens = action.payload;
    },
    setBiometricAvailable: (state, action: PayloadAction<boolean>) => {
      state.biometricAvailable = action.payload;
    },
    setBiometricEnabled: (state, action: PayloadAction<boolean>) => {
      state.biometricEnabled = action.payload;
      if (state.user) {
        state.user.preferences.biometricEnabled = action.payload;
      }
    },
    clearAuth: (state) => {
      state.isAuthenticated = false;
      state.user = null;
      state.tokens = {
        accessToken: null,
        refreshToken: null,
        expiresAt: null,
      };
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    // Login
    builder
      .addCase(loginUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.isLoading = false;
        state.isAuthenticated = true;
        state.user = action.payload.user;
        state.tokens = {
          accessToken: action.payload.accessToken,
          refreshToken: action.payload.refreshToken,
          expiresAt: action.payload.expiresAt,
        };
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
        state.isAuthenticated = false;
        state.user = null;
      });

    // Register
    builder
      .addCase(registerUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(registerUser.fulfilled, (state) => {
        state.isLoading = false;
        // Registration successful, user needs to login
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Logout
    builder
      .addCase(logoutUser.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(logoutUser.fulfilled, (state) => {
        state.isLoading = false;
        state.isAuthenticated = false;
        state.user = null;
        state.tokens = {
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
        };
        state.error = null;
      })
      .addCase(logoutUser.rejected, (state) => {
        state.isLoading = false;
        // Still clear auth state even if logout API fails
        state.isAuthenticated = false;
        state.user = null;
        state.tokens = {
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
        };
      });

    // Load stored auth
    builder
      .addCase(loadStoredAuth.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(loadStoredAuth.fulfilled, (state, action) => {
        state.isLoading = false;
        if (action.payload) {
          state.isAuthenticated = true;
          state.user = action.payload.user;
          state.tokens.accessToken = action.payload.accessToken;
        }
      })
      .addCase(loadStoredAuth.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Refresh token
    builder
      .addCase(refreshAuthToken.fulfilled, (state, action) => {
        state.tokens = {
          accessToken: action.payload.accessToken,
          refreshToken: action.payload.refreshToken,
          expiresAt: action.payload.expiresAt,
        };
      })
      .addCase(refreshAuthToken.rejected, (state) => {
        // Clear auth state on refresh failure
        state.isAuthenticated = false;
        state.user = null;
        state.tokens = {
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
        };
      });

    // Update user
    builder
      .addCase(updateUser.fulfilled, (state, action) => {
        state.user = action.payload;
      })
      .addCase(updateUser.rejected, (state, action) => {
        state.error = action.payload as string;
      });

    // Forgot password
    builder
      .addCase(forgotPassword.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(forgotPassword.fulfilled, (state) => {
        state.isLoading = false;
      })
      .addCase(forgotPassword.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });
  },
});