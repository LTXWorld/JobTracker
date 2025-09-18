import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';

// Import both mock and real slices
import { createAuthSlice as createMockAuthSlice } from './authSlice';
import { createAuthSlice as createRealAuthSlice } from './apiAuthSlice';
import { createApplicationSlice as createMockApplicationSlice } from './applicationSlice';
import { createApplicationSlice as createRealApplicationSlice } from './apiApplicationSlice';

interface ApiConfigState {
  useRealAPI: boolean;
  baseUrl: string;
  isConnected: boolean;
  lastConnectionCheck: Date | null;
  connectionError: string | null;
}

const initialState: ApiConfigState = {
  useRealAPI: __DEV__ ? false : true, // Use mock in development, real in production
  baseUrl: __DEV__ ? 'http://10.0.2.2:8080' : 'https://api.jobview.app',
  isConnected: false,
  lastConnectionCheck: null,
  connectionError: null,
};

// Async thunk to test API connection
export const testApiConnection = createAsyncThunk(
  'apiConfig/testConnection',
  async (baseUrl: string, { rejectWithValue }) => {
    try {
      // Test connection to the API
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

      const response = await fetch(`${baseUrl}/api/v1/health`, {
        method: 'GET',
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
        },
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      return {
        isConnected: true,
        serverInfo: data,
        baseUrl,
      };
    } catch (error: any) {
      if (error.name === 'AbortError') {
        return rejectWithValue('连接超时');
      }
      return rejectWithValue(error.message || '无法连接到服务器');
    }
  }
);

export const switchToRealAPI = createAsyncThunk(
  'apiConfig/switchToRealAPI',
  async (baseUrl: string, { dispatch, rejectWithValue }) => {
    try {
      // Test connection first
      const result = await dispatch(testApiConnection(baseUrl));

      if (testApiConnection.fulfilled.match(result)) {
        return {
          useRealAPI: true,
          baseUrl,
          isConnected: true,
        };
      } else {
        throw new Error(result.payload as string);
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '切换到真实API失败');
    }
  }
);

export const apiConfigSlice = createSlice({
  name: 'apiConfig',
  initialState,
  reducers: {
    setUseRealAPI: (state, action: PayloadAction<boolean>) => {
      state.useRealAPI = action.payload;
    },
    setBaseUrl: (state, action: PayloadAction<string>) => {
      state.baseUrl = action.payload;
    },
    setConnectionStatus: (state, action: PayloadAction<boolean>) => {
      state.isConnected = action.payload;
    },
    clearConnectionError: (state) => {
      state.connectionError = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(testApiConnection.pending, (state) => {
        state.connectionError = null;
      })
      .addCase(testApiConnection.fulfilled, (state, action) => {
        state.isConnected = action.payload.isConnected;
        state.baseUrl = action.payload.baseUrl;
        state.lastConnectionCheck = new Date();
        state.connectionError = null;
      })
      .addCase(testApiConnection.rejected, (state, action) => {
        state.isConnected = false;
        state.connectionError = action.payload as string;
        state.lastConnectionCheck = new Date();
      })
      .addCase(switchToRealAPI.fulfilled, (state, action) => {
        state.useRealAPI = action.payload.useRealAPI;
        state.baseUrl = action.payload.baseUrl;
        state.isConnected = action.payload.isConnected;
        state.connectionError = null;
      })
      .addCase(switchToRealAPI.rejected, (state, action) => {
        state.connectionError = action.payload as string;
      });
  },
});

// Factory function to create the appropriate slices based on API configuration
export const createSlicesForApiMode = (useRealAPI: boolean) => {
  if (useRealAPI) {
    return {
      authSlice: createRealAuthSlice(),
      applicationSlice: createRealApplicationSlice(),
    };
  } else {
    return {
      authSlice: createMockAuthSlice(),
      applicationSlice: createMockApplicationSlice(),
    };
  }
};

export const { setUseRealAPI, setBaseUrl, setConnectionStatus, clearConnectionError } = apiConfigSlice.actions;

export default apiConfigSlice.reducer;