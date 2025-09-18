import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { mockAuthAPI } from '@services/mockAuthService';
// import { authAPI } from '@services/authService'; // Use this for real API
import { User } from '@types';

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

// Storage keys
const STORAGE_KEYS = {
  AUTH_TOKEN: '@jobview/auth_token',
  REFRESH_TOKEN: '@jobview/refresh_token',
  USER_INFO: '@jobview/user_info',
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
      const response = await mockAuthAPI.login({ username, password });

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

      // Store in AsyncStorage
      await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, accessToken);
      await AsyncStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
      await AsyncStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(user));

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
      const response = await mockAuthAPI.register(userData);

      if (!response.success) {
        return rejectWithValue(response.message || '注册失败');
      }

      return response.message || '注册成功';
    } catch (error: any) {
      return rejectWithValue(error.message || '注册失败');
    }
  }
);

export const logoutUser = createAsyncThunk(
  'auth/logoutUser',
  async () => {
    try {
      // Clear local storage
      await AsyncStorage.multiRemove([
        STORAGE_KEYS.AUTH_TOKEN,
        STORAGE_KEYS.REFRESH_TOKEN,
        STORAGE_KEYS.USER_INFO,
      ]);
    } catch (error: any) {
      console.error('Logout error:', error);
    }
  }
);

export const loadStoredAuth = createAsyncThunk(
  'auth/loadStoredAuth',
  async (_, { rejectWithValue }) => {
    try {
      const [token, refreshToken, userStr] = await AsyncStorage.multiGet([
        STORAGE_KEYS.AUTH_TOKEN,
        STORAGE_KEYS.REFRESH_TOKEN,
        STORAGE_KEYS.USER_INFO,
      ]);

      if (token[1] && refreshToken[1] && userStr[1]) {
        const user = JSON.parse(userStr[1]);

        return {
          user,
          accessToken: token[1],
          refreshToken: refreshToken[1],
        };
      }

      return null;
    } catch (error) {
      return rejectWithValue('Failed to load stored auth');
    }
  }
);

export const forgotPassword = createAsyncThunk(
  'auth/forgotPassword',
  async (email: string, { rejectWithValue }) => {
    try {
      const response = await mockAuthAPI.forgotPassword(email);

      if (!response.success) {
        return rejectWithValue(response.message || '发送重置邮件失败');
      }

      return response.message || '重置邮件已发送';
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
    setUser: (state, action: PayloadAction<User>) => {
      state.user = action.payload;
    },
    setTokens: (state, action: PayloadAction<{ accessToken: string; refreshToken: string; expiresAt: Date }>) => {
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
    clearAuth: () => initialState,
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
        state.error = null;
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Register
    builder
      .addCase(registerUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(registerUser.fulfilled, (state) => {
        state.isLoading = false;
        state.error = null;
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Logout
    builder
      .addCase(logoutUser.fulfilled, (state) => {
        state.isAuthenticated = false;
        state.user = null;
        state.tokens = {
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
        };
        state.error = null;
      });

    // Load stored auth
    builder
      .addCase(loadStoredAuth.fulfilled, (state, action) => {
        if (action.payload) {
          state.isAuthenticated = true;
          state.user = action.payload.user;
          state.tokens = {
            accessToken: action.payload.accessToken,
            refreshToken: action.payload.refreshToken,
            expiresAt: null,
          };
        }
      });

    // Forgot password
    builder
      .addCase(forgotPassword.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(forgotPassword.fulfilled, (state) => {
        state.isLoading = false;
        state.error = null;
      })
      .addCase(forgotPassword.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });
  },
});