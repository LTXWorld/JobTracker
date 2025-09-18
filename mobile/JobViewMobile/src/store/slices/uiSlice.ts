import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface UIState {
  theme: 'light' | 'dark' | 'system';
  systemTheme: 'light' | 'dark';
  isLoading: boolean;
  loadingMessage: string;
  error: string | null;
  networkStatus: 'online' | 'offline';
  activeScreen: string;
  modal: {
    isVisible: boolean;
    type: string | null;
    data: any;
  };
  toast: {
    isVisible: boolean;
    message: string;
    type: 'success' | 'error' | 'warning' | 'info';
  };
}

const initialState: UIState = {
  theme: 'system',
  systemTheme: 'light',
  isLoading: false,
  loadingMessage: '',
  error: null,
  networkStatus: 'online',
  activeScreen: '',
  modal: {
    isVisible: false,
    type: null,
    data: null,
  },
  toast: {
    isVisible: false,
    message: '',
    type: 'info',
  },
};

export const createUiSlice = () => createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'system'>) => {
      state.theme = action.payload;
    },
    setSystemTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
      state.systemTheme = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean | { loading: boolean; message?: string }>) => {
      if (typeof action.payload === 'boolean') {
        state.isLoading = action.payload;
        if (!action.payload) {
          state.loadingMessage = '';
        }
      } else {
        state.isLoading = action.payload.loading;
        state.loadingMessage = action.payload.message || '';
      }
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
    setNetworkStatus: (state, action: PayloadAction<'online' | 'offline'>) => {
      state.networkStatus = action.payload;
    },
    setActiveScreen: (state, action: PayloadAction<string>) => {
      state.activeScreen = action.payload;
    },
    showModal: (state, action: PayloadAction<{ type: string; data?: any }>) => {
      state.modal = {
        isVisible: true,
        type: action.payload.type,
        data: action.payload.data || null,
      };
    },
    hideModal: (state) => {
      state.modal = {
        isVisible: false,
        type: null,
        data: null,
      };
    },
    showToast: (state, action: PayloadAction<{ message: string; type?: UIState['toast']['type'] }>) => {
      state.toast = {
        isVisible: true,
        message: action.payload.message,
        type: action.payload.type || 'info',
      };
    },
    hideToast: (state) => {
      state.toast = {
        ...state.toast,
        isVisible: false,
      };
    },
    resetUI: () => initialState,
  },
});